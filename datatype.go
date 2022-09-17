package oracle

import (
	"fmt"
	"strings"

	"gorm.io/gorm/schema"
)

// Selecting a Datatype
// https://docs.oracle.com/cd/A58617_01/server.804/a58241/ch5.htm

func (dialector Dialector) getSchemaFloatType(field *schema.Field) string {
	if field.Precision > 0 {
		return fmt.Sprintf("NUMBER(%d, %d)", field.Precision, field.Scale)
	}

	return "NUMBER"
}

func (dialector Dialector) getSchemaStringType(field *schema.Field) string {
	size := field.Size
	if size == 0 {
		if dialector.DefaultStringSize > 0 {
			size = int(dialector.DefaultStringSize)
		}
	}

	// TODO: 根据不同字段存储方式选择合适的类型
	if size <= 2000 {
		return fmt.Sprintf("CHAR(%d)", size)
	} else if size <= 4000 {
		return fmt.Sprintf("VARCHAR2(%d)", size)
	}

	return "CLOB"
}

func (dialector Dialector) getSchemaTimeType(field *schema.Field) string {
	if field.NotNull || field.PrimaryKey {
		return "DATE"
	}
	return "DATE NULL"
}

func (dialector Dialector) getSchemaBytesType(field *schema.Field) string {
	return "BLOB"
}

func (dialector Dialector) getSchemaIntAndUnitType(field *schema.Field) string {
	// https://blog.csdn.net/yzsind/article/details/7948226
	// https://docs.oracle.com/cd/E17952_01/mysql-8.0-en/integer-types.html
	// https://docs.oracle.com/database/121/DRDAS/data_type.htm#DRDAS241
	sqlType := "NUMBER"
	switch {
	case field.Size <= 8:
		sqlType = "NUMBER(3,0)"
	case field.Size <= 16:
		sqlType = "NUMBER(5,0)"
	case field.Size <= 24:
		sqlType = "NUMBER(7,0)"
	case field.Size <= 32:
		sqlType = "NUMBER(10,0)"
	}

	// TODO: test
	if field.AutoIncrement && dialector.Config.SupportIdentity {
		sqlType += " GENERATED ALWAYS as IDENTITY(START with 1 INCREMENT by 1)"
	}

	return sqlType
}

func (dialector Dialector) getSchemaCustomType(field *schema.Field) string {
	sqlType := string(field.DataType)

	// TODO: test
	if field.AutoIncrement && !strings.Contains(strings.ToLower(sqlType), " auto_increment") {
		sqlType += " GENERATED ALWAYS as IDENTITY(START with 1 INCREMENT by 1)"
	}

	return sqlType
}
