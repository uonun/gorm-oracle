package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	_ "github.com/sijms/go-ora/v2"
	"github.com/uonun/gorm-oracle/callbacks"

	"gorm.io/gorm"
	cbs "gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

/*
type Dialector interface {
  Name() string
  Initialize(*DB) error
  Migrator(db *DB) Migrator
  DataTypeOf(*schema.Field) string
  DefaultValueOf(*schema.Field) clause.Expression
  BindVarTo(writer clause.Writer, stmt *Statement, v interface{})
  QuoteTo(clause.Writer, string)
  Explain(sql string, vars ...interface{}) string
}
*/

// https://gorm.io/docs/write_driver.html

const (
	dialectorName        string = "oracle"
	ctxKeyIsBatchInsert  string = "is_batch_insert"
	ctxKeyNextFieldIndex string = "next_field_index"
)

type Config struct {
	DriverName    string
	ServerVersion string
	DSN           string
	Conn          gorm.ConnPool

	InitializeWithVersion bool

	// SupportIdentity 为 true 时支持 IDENTITY 关键字
	// See: https://docs.oracle.com/database/121/DRDAA/migr_tools_feat.htm#DRDAA109
	SupportIdentity bool

	DefaultStringSize uint
	// DefaultDatetimePrecision *int
	// DisableDatetimePrecision      bool
	DontSupportRenameIndex  bool
	DontSupportRenameColumn bool
	// DontSupportForShareClause     bool
	DontSupportNullAsDefaultValue bool
}

var (
	// CreateClauses create clauses
	CreateClauses = []string{"INSERT", "VALUES", "ON CONFLICT", "RETURNING"}

	supportReturning bool
)

type Dialector struct {
	*Config
}

func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}}
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

func (dialector Dialector) Name() string {
	return dialectorName
}

func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	ctx := context.Background()

	// register callbacks
	cbsCfg := &cbs.Config{
		CreateClauses: CreateClauses,
	}
	supportReturning = utils.Contains(CreateClauses, "RETURNING")

	cbs.RegisterDefaultCallbacks(db, cbsCfg)

	if dialector.DriverName == "" {
		dialector.DriverName = dialectorName
	}

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
		if err != nil {
			return errors.Wrapf(err, "sql.Open failed")
		}
	}

	if dialector.Config.InitializeWithVersion {
		err = db.ConnPool.QueryRowContext(ctx, "SELECT * FROM v$version	WHERE banner LIKE 'Oracle%'").Scan(&dialector.ServerVersion)
		if err != nil {
			return errors.Wrapf(err, "db.ConnPool.QueryRowContext failed")
		}

		// https://en.wikipedia.org/wiki/Oracle_Database
		if strings.Contains(dialector.ServerVersion, "12c") ||
			strings.Contains(dialector.ServerVersion, "18c") ||
			strings.Contains(dialector.ServerVersion, "19c") ||
			strings.Contains(dialector.ServerVersion, "21c") {
			dialector.Config.SupportIdentity = true
		}
	}

	if err = db.Callback().Create().Replace("gorm:create", callbacks.Create(cbsCfg)); err != nil {
		return
	}

	for k, v := range dialector.ClauseBuilders() {
		db.ClauseBuilders[k] = v
	}

	return
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: dialector,
			},
		},
		Dialector: dialector,
	}
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "boolean"
	case schema.Int, schema.Uint:
		return dialector.getSchemaIntAndUnitType(field)
	case schema.Float:
		return dialector.getSchemaFloatType(field)
	case schema.String:
		return dialector.getSchemaStringType(field)
	case schema.Time:
		return dialector.getSchemaTimeType(field)
	case schema.Bytes:
		return dialector.getSchemaBytesType(field)
	default:
		return dialector.getSchemaCustomType(field)
	}
}

func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}

const (
	// ClauseOnConflict for clause.ClauseBuilder ON CONFLICT key
	ClauseOnConflict = "ON CONFLICT"
	ClauseReturning  = "RETURNING"
	// ClauseValues for clause.ClauseBuilder VALUES key
	ClauseValues = "VALUES"
	// ClauseValues for clause.ClauseBuilder FOR key
	ClauseFor    = "FOR"
	ClauseInsert = "INSERT"
)

type fieldSet struct {
	idx int
	f   schema.Field
	c   clause.Column
}

func (dialector Dialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	clauseBuilders := map[string]clause.ClauseBuilder{
		ClauseOnConflict: func(c clause.Clause, builder clause.Builder) {
			onConflict, ok := c.Expression.(clause.OnConflict)
			if !ok {
				c.Build(builder)
				return
			}

			builder.WriteString("ON DUPLICATE KEY UPDATE ")
			if len(onConflict.DoUpdates) == 0 {
				if s := builder.(*gorm.Statement).Schema; s != nil {
					var column clause.Column
					onConflict.DoNothing = false

					if s.PrioritizedPrimaryField != nil {
						column = clause.Column{Name: s.PrioritizedPrimaryField.DBName}
					} else if len(s.DBNames) > 0 {
						column = clause.Column{Name: s.DBNames[0]}
					}

					if column.Name != "" {
						onConflict.DoUpdates = []clause.Assignment{{Column: column, Value: column}}
					}
				}
			}

			for idx, assignment := range onConflict.DoUpdates {
				if idx > 0 {
					builder.WriteByte(',')
				}

				builder.WriteQuoted(assignment.Column)
				builder.WriteByte('=')
				if column, ok := assignment.Value.(clause.Column); ok && column.Table == "excluded" {
					column.Table = ""
					builder.WriteString("VALUES(")
					builder.WriteQuoted(column)
					builder.WriteByte(')')
				} else {
					builder.AddVar(builder, assignment.Value)
				}
			}
		},
		ClauseReturning: func(c clause.Clause, builder clause.Builder) {
			returning, ok := c.Expression.(clause.Returning)
			if !ok {
				c.Build(builder)
				return
			}

			stmt := builder.(*gorm.Statement)
			isBatchInsert, _ := stmt.Context.Value(ctxKeyIsBatchInsert).(bool)
			if isBatchInsert {
				// do nothing
			} else {
				// RETURNING id INTO l_id;
				builder.WriteString("RETURNING ")

				colCount := len(returning.Columns)
				type fieldSet struct {
					f   *schema.Field
					idx int
				}
				returningFields := make([]fieldSet, colCount)
				for i := 0; i < colCount; i++ {
					colName := returning.Columns[i].Name
					if i > 0 {
						builder.WriteByte(',')
					}
					builder.WriteString(colName)

					for idx, f := range stmt.Schema.Fields {
						if f.DBName == colName {
							returningFields[i] = fieldSet{f, idx}
							break
						}
					}
				}

				builder.WriteString(" INTO ")

				for idx, fs := range returningFields {
					if idx > 0 {
						builder.WriteByte(',')
					}
					v := stmt.ReflectValue
					vv := v.Field(fs.idx).Addr().Interface()
					stmt.Vars = append(stmt.Vars, vv)
					builder.WriteString(fmt.Sprintf(":o%d", fs.idx))
				}
			}
		},
		ClauseInsert: func(c clause.Clause, builder clause.Builder) {
			insertClause, ok := c.Expression.(clause.Insert)
			if !ok {
				c.Build(builder)
				return
			}

			stmt := builder.(*gorm.Statement)
			// batch insert
			isBatchInsert := false
			k := stmt.ReflectValue.Kind()
			if k == reflect.Slice || k == reflect.Array {
				isBatchInsert = true
				stmt.Context = context.WithValue(stmt.Context, ctxKeyIsBatchInsert, true)
			}

			if isReturning, _ := hasReturning(stmt); isReturning && isBatchInsert {
				// BUILDING SQL: DECLARE
			} else {
				// BUILDING SQL: INSERT INTO
				insertClause.MergeClause(&c)
				c.Build(builder)
			}
		},
		ClauseValues: func(c clause.Clause, builder clause.Builder) {
			values, ok := c.Expression.(clause.Values)
			if !ok {
				c.Build(builder)
				return
			}

			stmt := builder.(*gorm.Statement)
			values = dialector.AddSequenceColumn(stmt, values)
			values.MergeClause(&c)

			colCount := len(values.Columns)
			valCount := len(values.Values)

			if isBatchInsert, _ := stmt.Context.Value(ctxKeyIsBatchInsert).(bool); isBatchInsert {
				if isReturning, _ := hasReturning(stmt); isReturning {

					//  DECLARE
					//    TYPE t_forall_test_tab IS TABLE OF allen_test%ROWTYPE;
					//    r  t_forall_test_tab := t_forall_test_tab();
					//    l_size NUMBER := 3;
					//  BEGIN
					//    -- Populate collection.
					//    FOR i IN 1 .. l_size LOOP
					//      r.extend;
					//      r(r.last).id := allen_test_s.nextval;
					//      :id1 := r(r.last).id;
					//      r(r.last).col1 := TO_CHAR(i);
					//      r(r.last).col2 := 'Description: ' || TO_CHAR(i);
					//    END LOOP;
					//    -- Time bulk inserts.
					//    FORALL i IN r.first .. r.last
					//      INSERT INTO allen_test VALUES r (i);
					//    COMMIT;
					//  END;

					allFields := make([]fieldSet, colCount)
					for i := 0; i < colCount; i++ {
						colName := values.Columns[i].Name
						for idx, f := range stmt.Schema.Fields {
							if f.DBName == colName {
								allFields[i] = fieldSet{idx, *f, values.Columns[i]}
								break
							}
						}
					}
					colInsert := ""
					colCount := len(values.Columns)
					for i := 0; i < colCount; i++ {
						if i > 0 {
							colInsert += ","
						}
						colInsert += values.Columns[i].Name
					}

					builder.WriteString("DECLARE\n")
					builder.WriteString(fmt.Sprintf("\tTYPE t IS TABLE OF %s%%ROWTYPE;\n", stmt.Schema.Table))
					builder.WriteString("\tr t := t();\n")
					builder.WriteString("BEGIN\n")
					for i := 0; i < valCount; i++ {
						builder.WriteString("\tr.extend;\n")
						for j := 0; j < colCount; j++ {
							fs := allFields[j]
							para := ""
							if seqName, isSeq := fs.f.TagSettings["SEQUENCE"]; isSeq {
								builder.WriteString(fmt.Sprintf("\t:p%d_%d := %s.NEXTVAL;\n", i, fs.idx, seqName))
								para = fmt.Sprintf(":p%d_%d", i, fs.idx)
							} else {
								para = fmt.Sprintf(":p%d_%d", i, fs.idx)
							}
							builder.WriteString(fmt.Sprintf("\tr(r.last).%s := %s;\n", fs.f.DBName, para))
						}
						stmt.Vars = append(stmt.Vars, values.Values[i]...)
					}
					builder.WriteString("\tFORALL i IN r.first .. r.last\n")
					builder.WriteString(fmt.Sprintf("\t\tINSERT INTO %s VALUES r (i);\n", stmt.Schema.Table))
					// builder.WriteString("\tDBMS_OUTPUT.PUT_LINE(TO_Char(SQL%ROWCOUNT)||' rows affected.');\n")
					builder.WriteString("\tCOMMIT;\n")
					builder.WriteString("END;")
				} else {
					// INSERT INTO tableName (id,col1,col2)
					// 	SELECT tableSequence.NEXTVAL, col1, col2
					// 		FROM (
					// 			SELECT :p0_0 AS col1, :p0_1 AS col2 FROM DUAL UNION ALL
					// 			SELECT :p1_0 AS col1, :p1_1 AS col2 FROM DUAL UNION ALL
					// 			SELECT :p2_0 AS col1, :p2_1 AS col2 FROM DUAL
					// 		)

					colInsert := ""
					colSelect := ""
					colCount := len(values.Columns)
					for i := 0; i < colCount; i++ {
						if i > 0 {
							colInsert += ","
							colSelect += ","
						}

						field := stmt.Schema.Fields[i]

						colInsert += values.Columns[i].Name
						if seqName, isSeq := field.TagSettings["SEQUENCE"]; isSeq {
							colSelect += fmt.Sprintf("%s.NEXTVAL", seqName)
						} else {
							colSelect += field.Name
						}

					}

					builder.WriteString(fmt.Sprintf("(%s) SELECT %s FROM (", colInsert, colSelect))

					valCount := len(values.Values)
					for i := 0; i < valCount; i++ {
						builder.WriteString("SELECT ")
						stmt.AddVar(builder, values.Values[i]...)
						builder.WriteString(" FROM DUAL ")

						if i < valCount-1 {
							builder.WriteString("UNION ALL ")
						}
					}

					builder.WriteString(")")
				}
			} else {
				c.Build(builder)
			}
		},
	}

	return clauseBuilders
}

// hasReturning see: gorm/callbacks/helper.go:L96
func hasReturning(stmt *gorm.Statement) (bool, gorm.ScanMode) {
	if c, ok := stmt.Clauses["RETURNING"]; ok {
		returning, _ := c.Expression.(clause.Returning)
		if len(returning.Columns) == 0 || (len(returning.Columns) == 1 && returning.Columns[0].Name == "*") {
			return true, 0
		}
		return true, gorm.ScanUpdate
	}
	return false, 0
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "NULL"}
}

// AddSequenceColumn add sequence column
func (dialector Dialector) AddSequenceColumn(stmt *gorm.Statement, values clause.Values) clause.Values {
	dbNameCount := len(stmt.Schema.DBNames)

	isColumnExists := func(cols []clause.Column, colName string) bool {
		if cols != nil {
			for i := 0; i < len(cols); i++ {
				if cols[i].Name == colName {
					return true
				}
			}
		}
		return false
	}

	isBatchInsert, _ := stmt.Context.Value(ctxKeyIsBatchInsert).(bool)

	for i := 0; i < dbNameCount; i++ {
		db := stmt.Schema.DBNames[i]
		field := stmt.Schema.LookUpField(db)
		if _, isSeq := field.TagSettings["SEQUENCE"]; isSeq {

			var idx int
			for i, f := range stmt.Schema.Fields {
				if f.DBName == field.DBName {
					idx = i
					break
				}
			}

			v := stmt.ReflectValue
			// exists:
			//	case 1: `autoIncrement` column
			//		need to insert/append the column & value.
			//  case 2: db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "CUSTOMER_ID"},},})
			// 		need to overwrite the exists value.
			exists := isColumnExists(values.Columns, db)
			// insert at index i
			if i < len(values.Columns) {
				if !exists {
					values.Columns = append(values.Columns[:i+1], values.Columns[i:]...)
					values.Columns[i] = clause.Column{Name: db}
				}
				for j := 0; j < len(values.Values); j++ {
					if !exists {
						values.Values[j] = append(values.Values[j][:i+1], values.Values[j][i:]...)
					}
					if isBatchInsert {
						values.Values[j][i] = v.Index(j).Field(idx).Addr().Interface()
					} else {
						values.Values[j][i] = v.Field(idx).Addr().Interface()
					}
				}
			} else { // append at the end
				if !exists {
					values.Columns = append(values.Columns, clause.Column{Name: db})
				}
				for j := 0; j < len(values.Values); j++ {
					if !exists {
						values.Values[j] = append(values.Values[j], nil)
					}
					if isBatchInsert {
						values.Values[j][i] = v.Index(j).Field(idx).Addr().Interface()
					} else {
						values.Values[j][i] = v.Field(idx).Addr().Interface()
					}
				}
			}
		}
	}
	return values
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {

	isInsert := utils.Contains(stmt.BuildClauses, "INSERT")
	if !isInsert {
		writer.WriteString(fmt.Sprintf(":p%d", len(stmt.Vars)))
		return
	}

	writer.WriteString(dialector.BindVarParameter(stmt))
}

func (dialector Dialector) BindVarParameter(stmt *gorm.Statement) string {

	// the value for sequence column need to be removed
	// use the index to record the removed one in the looping
	idx, ok := stmt.Context.Value(ctxKeyNextFieldIndex).(int)
	if ok {
		stmt.Context = context.WithValue(stmt.Context, ctxKeyNextFieldIndex, idx+1)
	} else {
		stmt.Context = context.WithValue(stmt.Context, ctxKeyNextFieldIndex, 1)
	}

	fieldIndex := idx % len(stmt.Schema.Fields)
	valueIndex := idx / len(stmt.Schema.Fields)

	field := stmt.Schema.Fields[fieldIndex]
	if seqName, isSeq := field.TagSettings["SEQUENCE"]; isSeq {

		// for sequence Columns, need no value to bind to parameters
		removing := len(stmt.Vars) - 1
		stmt.Vars = append(stmt.Vars[:removing], stmt.Vars[removing+1:]...)

		isBatchInsert, ok := stmt.Context.Value(ctxKeyIsBatchInsert).(bool)
		if ok && isBatchInsert {
			if isReturning, _ := callbacks.HasReturning(stmt.DB, supportReturning); !isReturning {
				// BUILDING SQL: INSERT INTO ...
				// placeholder for sequence column
				return "NULL"
			}
		} else {
			return fmt.Sprintf("%s.NEXTVAL", seqName)
		}

	} else {
		isBatchInsert, ok := stmt.Context.Value(ctxKeyIsBatchInsert).(bool)
		if ok && isBatchInsert {
			if isReturning, _ := callbacks.HasReturning(stmt.DB, supportReturning); !isReturning {
				// BUILDING SQL: INSERT INTO ...
				return fmt.Sprintf(":p%d_%d AS %s", valueIndex, fieldIndex, field.Name)
			}
		} else {
			return fmt.Sprintf(":p%d_%d", valueIndex, fieldIndex)
		}
	}

	return ""
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	writer.WriteString(str)
}

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}
