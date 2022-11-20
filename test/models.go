package test

import (
	"context"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CustomerModel interface {
	CustomerWithSequenceButNotReturning | Customer | CustomerOfNClob | CustomerOfUDT | CustomerWithPrimaryKey | CustomerWithHook
	GetCustomerID() int64
}

///------Generate sequence(autoIncrement value)----------------------------------------------------------------
// CustomerWithSequenceButNotReturning table comment
// - use `sequence` to specify the sequence name.
// - need db.Clauses(clause.Returning{...}) to return new value of sequence column.
type CustomerWithSequenceButNotReturning struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// test case: the sequence column is not the first one
	CustomerID int64 `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S" json:"customer_id"`
	// use go_ora.NVarChar/go_ora.NClob for NVARCHAR/NCLOB columns
	Address     go_ora.NVarChar `gorm:"column:ADDRESS" json:"address"`
	City        string          `gorm:"column:CITY" json:"city"`
	State       string          `gorm:"column:STATE" json:"state"`
	ZipCode     string          `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time       `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32           `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerWithSequenceButNotReturning) TableName() string {
	return "Customers"
}

///-------Auto returning auto-generated sequence---------------------------------------------------------------
// Customer table comment
// - use `autoIncrement` to specify the auto generated sequence value.
type Customer struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// use go_ora.NVarChar/go_ora.NClob for NVARCHAR/NCLOB columns
	Address     go_ora.NVarChar `gorm:"column:ADDRESS" json:"address"`
	City        string          `gorm:"column:CITY" json:"city"`
	State       string          `gorm:"column:STATE" json:"state"`
	ZipCode     string          `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time       `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32           `gorm:"column:AGE" json:"age"`
	// test case: the returning column `autoIncrement` is at the end
	CustomerID int64 `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement" json:"customer_id"`
}

// TableName sets the insert table name for this struct type
func (c *Customer) TableName() string {
	return "Customers"
}

func (c Customer) GetCustomerID() int64 {
	return c.CustomerID
}

///--------go_ora.NVarChar/go_ora.NClob for NCLOB columns--------------------------------------------------------
// CustomerOfNClob table comment
// use `type` for go_ora.NClob
type CustomerOfNClob struct {
	CustomerID   int64  `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement" json:"customer_id"`
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// use go_ora.NVarChar/go_ora.NClob for NVARCHAR/NCLOB columns
	Address     *go_ora.NClob `gorm:"column:ADDRESS;type:nclob" json:"address"`
	City        string        `gorm:"column:CITY" json:"city"`
	State       string        `gorm:"column:STATE" json:"state"`
	ZipCode     string        `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time     `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32         `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerOfNClob) TableName() string {
	return "Customers"
}

func (c CustomerOfNClob) GetCustomerID() int64 {
	return c.CustomerID
}

///----------------------------------------------------------------------
// CustomerOfUDT table comment
type CustomerOfUDT struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// Address      string       `gorm:"column:ADDRESS" json:"address"`
	Address     *NClobContent `gorm:"column:ADDRESS" json:"address"`
	City        string        `gorm:"column:CITY" json:"city"`
	State       string        `gorm:"column:STATE" json:"state"`
	ZipCode     string        `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time     `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32         `gorm:"column:AGE" json:"age"`
	// test case: the returning column `autoIncrement` is at the end
	CustomerID int64 `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement" json:"customer_id"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerOfUDT) TableName() string {
	return "Customers"
}

type NClobContent struct {
	String go_ora.NClob
}

// func (c *ContentNClob) Scan(value interface{}) error {
// 	bytes, ok := value.([]byte)
// 	if !ok {
// 		return errors.New(fmt.Sprint("Failed to unmarshal CONTENT value:", value))
// 	}

// 	c.String = go_ora.NClob{String: string(bytes), Valid: true}
// 	return nil
// }

// func (c *ContentNClob) Value() (driver.Value, error) {
// 	return c.String, nil
// }

func (c *NClobContent) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return gorm.Expr(":p_nclob_string", c.String)
}

// func (*ContentNClob) GormDataType() string {
// 	return "nclob"
// }

// func (*ContentNClob) GormDBDataType(db *gorm.DB, field *schema.Field) string {
// 	return "NCLOB"
// }

// // MarshalJSON implements json.Marshaler to convert Time to json serialization.
// func (c *ContentNClob) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(c.String)
// }

// // UnmarshalJSON implements json.Unmarshaler to deserialize json data.
// func (c *ContentNClob) UnmarshalJSON(data []byte) error {
// 	// ignore null
// 	if string(data) == "null" {
// 		return nil
// 	}
// 	c.String = string(data)
// 	c.Valid = true
// 	return nil
// }

///----------------------------------------------------------------------

// CustomerWithPrimaryKey table comment
// - use `primaryKey` to create WHERE condition.
type CustomerWithPrimaryKey struct {
	CustomerName string          `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	Address      go_ora.NVarChar `gorm:"column:ADDRESS" json:"address"`
	City         string          `gorm:"column:CITY" json:"city"`
	State        string          `gorm:"column:STATE" json:"state"`
	ZipCode      string          `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime  time.Time       `gorm:"column:CREATED_TIME" json:"created_time"`
	Age          int32           `gorm:"column:AGE" json:"age"`
	// test case: the returning column `autoIncrement` is at the end
	CustomerID int64 `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement;primaryKey" json:"customer_id"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerWithPrimaryKey) TableName() string {
	return "Customers"
}

///----Customer callbacks------------------------------------------------------------------
// CustomerWithHook table comment
type CustomerWithHook struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// test case: the sequence column is not the first
	CustomerID  int64           `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S" json:"customer_id"`
	Address     go_ora.NVarChar `gorm:"column:ADDRESS" json:"address"`
	City        string          `gorm:"column:CITY" json:"city"`
	State       string          `gorm:"column:STATE" json:"state"`
	ZipCode     string          `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time       `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32           `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerWithHook) TableName() string {
	return "Customers"
}

func (c *CustomerWithHook) BeforeCreate(tx *gorm.DB) (err error) {
	c.State = "HOOK:" + c.State
	return nil
}
