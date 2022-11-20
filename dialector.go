package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
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

var (
	// CreateClauses create clauses
	CreateClauses = []string{"INSERT", "VALUES", "ON CONFLICT", "RETURNING"}

	// defaultDatetimePrecision = 3
)

type Dialector struct {
	*Config
}

// Name implements gorm.Dialector interface
func (dialector Dialector) Name() string {
	return dialectorName
}

// Initialize implements gorm.Dialector interface
func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	ctx := context.Background()

	// register callbacks
	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		CreateClauses: CreateClauses,
	})

	if dialector.DriverName == "" {
		dialector.DriverName = dialectorName
	}

	if dialector.connPool != nil {
		db.ConnPool = dialector.connPool
	} else {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
		if err != nil {
			return errors.Wrapf(err, "sql.Open failed")
		}
	}

	if !dialector.Config.SkipInitializeWithVersion {
		err = db.ConnPool.QueryRowContext(ctx, "SELECT * FROM v$version	WHERE banner LIKE 'Oracle%'").Scan(&dialector.serverVersion)
		if err != nil {
			return errors.Wrapf(err, "db.ConnPool.QueryRowContext failed")
		}

		// https://en.wikipedia.org/wiki/Oracle_Database
		// TEST：tested: Oracle Database 11g Enterprise Edition Release 11.2.0.4.0 - 64bit Production
		if strings.Contains(dialector.serverVersion, "12c") ||
			strings.Contains(dialector.serverVersion, "18c") ||
			strings.Contains(dialector.serverVersion, "19c") ||
			strings.Contains(dialector.serverVersion, "21c") {
			dialector.Config.supportIdentity = true
			dialector.Config.supportOffsetFetch = true
		}
	}

	for k, v := range dialector.ClauseBuilders() {
		db.ClauseBuilders[k] = v
	}

	return
}

// Migrator implements gorm.Dialector interface
func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:                          db,
				Dialector:                   dialector,
				CreateIndexAfterCreateTable: true,
			},
		},
		Dialector: dialector,
	}
}

// DataTypeOf implements gorm.Dialector interface
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

// DefaultValueOf implements gorm.Dialector interface
func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "NULL"}
}

// BindVarTo implements gorm.Dialector interface
func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {

	isInsert := utils.Contains(stmt.BuildClauses, "INSERT")
	if !isInsert {
		writer.WriteString(fmt.Sprintf(":p%d", len(stmt.Vars)))
		return
	}

	writer.WriteString(dialector.bindVarParameter(stmt))
}

// QuoteTo implements gorm.Dialector interface
func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	writer.WriteString(str)
}

// Explain implements gorm.Dialector interface
func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}

// SavePoint implements gorm.SavePointerDialectorInterface interface
func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

// RollbackTo implements gorm.SavePointerDialectorInterface interface
func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}
