package test

import (
	"time"

	"gorm.io/gorm"
)

// Customer table comment
// use `sequence` to creating sequence columns.
type Customer struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// test case: the sequence column is not the first
	CustomerID  int64     `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S" json:"customer_id"`
	Address     string    `gorm:"column:ADDRESS" json:"address"`
	City        string    `gorm:"column:CITY" json:"city"`
	State       string    `gorm:"column:STATE" json:"state"`
	ZipCode     string    `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32     `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (c *Customer) TableName() string {
	return "Customers"
}

// CustomerReturning table comment
// use `autoIncrement` to returning sequence columns.
type CustomerReturning struct {
	CustomerName string    `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	Address      string    `gorm:"column:ADDRESS" json:"address"`
	City         string    `gorm:"column:CITY" json:"city"`
	State        string    `gorm:"column:STATE" json:"state"`
	ZipCode      string    `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime  time.Time `gorm:"column:CREATED_TIME" json:"created_time"`
	Age          int32     `gorm:"column:AGE" json:"age"`
	// test case: the returning column `autoIncrement` is at the end
	CustomerID int64 `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement" json:"customer_id"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerReturning) TableName() string {
	return "Customers"
}

// CustomerReturningPrimaryKey table comment
// use `primaryKey` to create WHERE condition.
type CustomerReturningPrimaryKey struct {
	CustomerName string    `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	Address      string    `gorm:"column:ADDRESS" json:"address"`
	City         string    `gorm:"column:CITY" json:"city"`
	State        string    `gorm:"column:STATE" json:"state"`
	ZipCode      string    `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime  time.Time `gorm:"column:CREATED_TIME" json:"created_time"`
	Age          int32     `gorm:"column:AGE" json:"age"`
	// test case: the returning column `autoIncrement` is at the end
	CustomerID int64 `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement;primaryKey" json:"customer_id"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerReturningPrimaryKey) TableName() string {
	return "Customers"
}

// CustomerHook table comment
// use `sequence` to creating sequence columns.
type CustomerHook struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	// test case: the sequence column is not the first
	CustomerID  int64     `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S" json:"customer_id"`
	Address     string    `gorm:"column:ADDRESS" json:"address"`
	City        string    `gorm:"column:CITY" json:"city"`
	State       string    `gorm:"column:STATE" json:"state"`
	ZipCode     string    `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime time.Time `gorm:"column:CREATED_TIME" json:"created_time"`
	Age         int32     `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (c *CustomerHook) TableName() string {
	return "Customers"
}

func (c *CustomerHook) BeforeCreate(tx *gorm.DB) (err error) {
	c.State = "HOOK:" + c.State
	return nil
}
