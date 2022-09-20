package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/sijms/go-ora/v2"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
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

	// DontSupportIdentity 为 ture 时表明不支持 IDENTITY 关键字
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
	CreateClauses = []string{"INSERT", "VALUES", "ON CONFLICT"}
	// QueryClauses query clauses
	QueryClauses = []string{}
	// UpdateClauses update clauses
	UpdateClauses = []string{"UPDATE", "SET", "WHERE", "ORDER BY", "LIMIT"}
	// DeleteClauses delete clauses
	DeleteClauses = []string{"DELETE", "FROM", "WHERE", "ORDER BY", "LIMIT"}

	// defaultDatetimePrecision = 3
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
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		CreateClauses: CreateClauses,
		QueryClauses:  QueryClauses,
		UpdateClauses: UpdateClauses,
		DeleteClauses: DeleteClauses,
	})

	if dialector.DriverName == "" {
		dialector.DriverName = dialectorName
	}

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
		if err != nil {
			return err
		}
	}

	if !dialector.Config.InitializeWithVersion {
		err = db.ConnPool.QueryRowContext(ctx, "SELECT * FROM v$version	WHERE banner LIKE 'Oracle%'").Scan(&dialector.ServerVersion)
		if err != nil {
			return err
		}

		// https://en.wikipedia.org/wiki/Oracle_Database
		if strings.Contains(dialector.ServerVersion, "12c") ||
			strings.Contains(dialector.ServerVersion, "18c") ||
			strings.Contains(dialector.ServerVersion, "19c") ||
			strings.Contains(dialector.ServerVersion, "21c") {
			dialector.Config.SupportIdentity = true
		}
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
	// ClauseValues for clause.ClauseBuilder VALUES key
	ClauseValues = "VALUES"
	// ClauseValues for clause.ClauseBuilder FOR key
	ClauseFor    = "FOR"
	ClauseInsert = "INSERT"
)

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
		ClauseInsert: func(c clause.Clause, builder clause.Builder) {
			insertClause, ok := c.Expression.(clause.Insert)
			if ok {
				stmt := builder.(*gorm.Statement)
				k := stmt.ReflectValue.Kind()
				if k == reflect.Slice || k == reflect.Array {
					insertClause.Modifier = "ALL"
					stmt.Context = context.WithValue(stmt.Context, ctxKeyIsBatchInsert, true)
				}
				insertClause.MergeClause(&c)
			}
			c.Build(builder)
		},
		ClauseValues: func(c clause.Clause, builder clause.Builder) {
			values, _ := c.Expression.(clause.Values)

			stmt := builder.(*gorm.Statement)
			values = dialector.AddSequenceColumn(stmt, values)
			values.MergeClause(&c)

			if isBatchInsert, _ := stmt.Context.Value(ctxKeyIsBatchInsert).(bool); isBatchInsert {
				// https://database.guide/4-ways-to-insert-multiple-rows-in-oracle/
				// INSERT ALL
				// 		INTO TABLE (C1, C2) VALUES (V1,V2)
				// 		INTO TABLE (C1, C2) VALUES (V1,V2)
				// 		INTO TABLE (C1, C2) VALUES (V1,V2)
				// SELECT 1 FORM DUAL

				colCount := len(values.Columns)
				colClausePart1 := " ("
				for i := 0; i < colCount; i++ {
					colClausePart1 += values.Columns[i].Name
					if i < colCount-1 {
						colClausePart1 += ","
					}
				}
				colClausePart1 += ") VALUES ("
				colClausePart2 := ")"
			
				valCount := len(values.Values)
				for i := 0; i < valCount; i++ {
					if i > 0 {
						builder.WriteString(fmt.Sprintf(" INTO %s ", stmt.Schema.Table))
					}
					builder.WriteString(colClausePart1)
					stmt.AddVar(builder, values.Values[i]...)
					builder.WriteString(colClausePart2)
				}

				builder.WriteString(" SELECT 1 FROM DUAL ")
			} else {
				c.Build(builder)
			}
		},
	}

	// if dialector.Config.DontSupportForShareClause {
	// 	clauseBuilders[ClauseFor] = func(c clause.Clause, builder clause.Builder) {
	// 		if values, ok := c.Expression.(clause.Locking); ok && strings.EqualFold(values.Strength, "SHARE") {
	// 			builder.WriteString("LOCK IN SHARE MODE")
	// 			return
	// 		}
	// 		c.Build(builder)
	// 	}
	// }

	return clauseBuilders
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {

	isInsert := utils.Contains(stmt.BuildClauses, "INSERT")
	if !isInsert {
		writer.WriteString(fmt.Sprintf(":p%d", len(stmt.Vars)))
		return
	}

	writer.WriteString(dialector.BindVarParameter(stmt))
}

// AddSequenceColumn add sequence column
func (dialector Dialector) AddSequenceColumn(stmt *gorm.Statement, values clause.Values) clause.Values {
	dbNameCount := len(stmt.Schema.DBNames)
	for i := 0; i < dbNameCount; i++ {
		db := stmt.Schema.DBNames[i]
		field := stmt.Schema.FieldsByDBName[db]
		if _, isSeq := field.TagSettings["SEQUENCE"]; isSeq {
			if i < len(values.Columns) {
				values.Columns = append(values.Columns[:i+1], values.Columns[i:]...)
				values.Columns[i] = clause.Column{Name: db}
				for j := 0; j < len(values.Values); j++ {
					values.Values[j] = append(values.Values[j][:i+1], values.Values[j][i:]...)
					values.Values[j][i] = field.DefaultValueInterface
				}
			} else {
				values.Columns = append(values.Columns, clause.Column{Name: db})
				values.Values[i] = append(values.Values[i], field.DefaultValueInterface)
			}
		}
	}
	return values
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
		if ok && isBatchInsert && valueIndex > 0 {
			return fmt.Sprintf("%s.nextval + %d", seqName, valueIndex)
		} else {
			return fmt.Sprintf("%s.nextval", seqName)
		}
	} else {
		return fmt.Sprintf(":p%d_%s", valueIndex, field.Name)
	}
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)

	for _, v := range []byte(str) {
		switch v {
		case '`':
			continuousBacktick++
			if continuousBacktick == 2 {
				writer.WriteString("``")
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				writer.WriteByte('`')
			}
			writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				// writer.WriteByte('`')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				// writer.WriteString("``")
			}

			writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		writer.WriteString("``")
	}
	// writer.WriteByte('`')
}

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}
