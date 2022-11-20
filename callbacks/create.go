package callbacks

import (
	"reflect"

	"gorm.io/gorm"
	cbs "gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils"
)

// Create create hook
func Create(config *cbs.Config) func(db *gorm.DB) {
	supportReturning := utils.Contains(config.CreateClauses, "RETURNING")

	return func(db *gorm.DB) {
		if db.Error != nil {
			return
		}

		if db.Statement.Schema != nil {
			if !db.Statement.Unscoped {
				for _, c := range db.Statement.Schema.CreateClauses {
					db.Statement.AddClause(c)
				}
			}

			if supportReturning && len(db.Statement.Schema.FieldsWithDefaultDBValue) > 0 {
				if _, ok := db.Statement.Clauses["RETURNING"]; !ok {
					fromColumns := make([]clause.Column, 0, len(db.Statement.Schema.FieldsWithDefaultDBValue))
					for _, field := range db.Statement.Schema.FieldsWithDefaultDBValue {
						fromColumns = append(fromColumns, clause.Column{Name: field.DBName})
					}
					db.Statement.AddClause(clause.Returning{Columns: fromColumns})
				}
			}
		}

		if db.Statement.SQL.Len() == 0 {
			db.Statement.SQL.Grow(180)
			db.Statement.AddClauseIfNotExists(clause.Insert{})
			db.Statement.AddClause(cbs.ConvertToCreateValues(db.Statement))

			db.Statement.Build(db.Statement.BuildClauses...)
		}

		isDryRun := !db.DryRun && db.Error == nil
		if !isDryRun {
			return
		}

		ok, mode := HasReturning(db, supportReturning)
		if ok {
			if c, ok := db.Statement.Clauses["ON CONFLICT"]; ok {
				if onConflict, _ := c.Expression.(clause.OnConflict); onConflict.DoNothing {
					mode |= gorm.ScanOnConflictDoNothing
				}
			}

			rows, err := db.Statement.ConnPool.QueryContext(
				db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...,
			)
			if db.AddError(err) == nil {
				defer func() {
					db.AddError(rows.Close())
				}()
				gorm.Scan(rows, db, mode)
			}

			return
		}

		result, err := db.Statement.ConnPool.ExecContext(
			db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...,
		)
		if err != nil {
			db.AddError(err)
			return
		}

		db.RowsAffected, _ = result.RowsAffected()
		if db.RowsAffected != 0 && db.Statement.Schema != nil &&
			db.Statement.Schema.PrioritizedPrimaryField != nil &&
			db.Statement.Schema.PrioritizedPrimaryField.HasDefaultValue {
			insertID, err := result.LastInsertId()
			insertOk := err == nil && insertID > 0
			if !insertOk {
				db.AddError(err)
				return
			}

			switch db.Statement.ReflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				if config.LastInsertIDReversed {
					for i := db.Statement.ReflectValue.Len() - 1; i >= 0; i-- {
						rv := db.Statement.ReflectValue.Index(i)
						if reflect.Indirect(rv).Kind() != reflect.Struct {
							break
						}

						_, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv)
						if isZero {
							db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, rv, insertID))
							insertID -= db.Statement.Schema.PrioritizedPrimaryField.AutoIncrementIncrement
						}
					}
				} else {
					for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
						rv := db.Statement.ReflectValue.Index(i)
						if reflect.Indirect(rv).Kind() != reflect.Struct {
							break
						}

						if _, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv); isZero {
							db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, rv, insertID))
							insertID += db.Statement.Schema.PrioritizedPrimaryField.AutoIncrementIncrement
						}
					}
				}
			case reflect.Struct:
				_, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
				if isZero {
					db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, db.Statement.ReflectValue, insertID))
				}
			}
		}
	}
}
