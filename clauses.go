package oracle

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

const (
	// ClauseOnConflict for clause.ClauseBuilder ON CONFLICT key
	ClauseOnConflict = "ON CONFLICT"
	ClauseReturning  = "RETURNING"
	// ClauseValues for clause.ClauseBuilder VALUES key
	ClauseValues = "VALUES"
	ClauseLimit  = "LIMIT"
	// ClauseValues for clause.ClauseBuilder FOR key
	ClauseFor    = "FOR"
	ClauseInsert = "INSERT"
)

type fieldSet struct {
	idx int
	f   schema.Field
	c   clause.Column
}

func (d Dialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	clauseBuilders := map[string]clause.ClauseBuilder{
		ClauseOnConflict: d.HandleOnConflict,
		ClauseReturning:  d.HandleReturning,
		ClauseInsert:     d.HandleInsert,
		ClauseValues:     d.HandleValues,
		ClauseLimit:      d.HandleLimit,
	}

	return clauseBuilders
}

func (d Dialector) HandleOnConflict(c clause.Clause, builder clause.Builder) {
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
}

func (d Dialector) HandleValues(c clause.Clause, builder clause.Builder) {
	values, ok := c.Expression.(clause.Values)
	if !ok {
		c.Build(builder)
		return
	}

	stmt := builder.(*gorm.Statement)
	values = d.addSequenceColumn(stmt, values)
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
}

func (d Dialector) HandleReturning(c clause.Clause, builder clause.Builder) {
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
}

func (d Dialector) HandleInsert(c clause.Clause, builder clause.Builder) {
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
}

func (d Dialector) HandleLimit(c clause.Clause, builder clause.Builder) {
	if !d.Config.SupportOffsetFetch {
		c.Build(builder)
		return
	}

	if limit, ok := c.Expression.(clause.Limit); ok {
		if stmt, ok := builder.(*gorm.Statement); ok {
			if _, ok := stmt.Clauses["ORDER BY"]; !ok {
				s := stmt.Schema
				builder.WriteString("ORDER BY ")
				if s != nil && s.PrioritizedPrimaryField != nil {
					builder.WriteQuoted(s.PrioritizedPrimaryField.DBName)
					builder.WriteByte(' ')
				} else {
					builder.WriteString("(SELECT NULL FROM DUAL)")
				}
			}
		}

		if offset := limit.Offset; offset > 0 {
			builder.WriteString(" OFFSET ")
			builder.WriteString(strconv.Itoa(offset))
			builder.WriteString(" ROWS")
		}
		if limit := limit.Limit; limit > 0 {
			builder.WriteString(" FETCH NEXT ")
			builder.WriteString(strconv.Itoa(limit))
			builder.WriteString(" ROWS ONLY")
		}
	}
}

// addSequenceColumn add sequence column
func (dialector Dialector) addSequenceColumn(stmt *gorm.Statement, values clause.Values) clause.Values {
	dbNameCount := len(stmt.Schema.DBNames)

	isExists := func(cols []clause.Column, colName string) bool {
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
			exists := isExists(values.Columns, db)
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

func (dialector Dialector) bindVarParameter(stmt *gorm.Statement) string {

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
			if isReturning, _ := hasReturning(stmt); !isReturning {
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
			if isReturning, _ := hasReturning(stmt); !isReturning {
				// BUILDING SQL: INSERT INTO ...
				return fmt.Sprintf(":p%d_%d AS %s", valueIndex, fieldIndex, field.Name)
			}
		} else {
			return fmt.Sprintf(":p%d_%d", valueIndex, fieldIndex)
		}
	}

	return ""
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
