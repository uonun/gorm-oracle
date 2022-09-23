package test

import "time"

// Customer table comment
type Customer struct {
	CustomerName string    `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	CustomerID   int64     `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S" json:"customer_id"`
	Address      string    `gorm:"column:ADDRESS" json:"address"`
	City         string    `gorm:"column:CITY" json:"city"`
	State        string    `gorm:"column:STATE" json:"state"`
	ZipCode      string    `gorm:"column:ZIP_CODE" json:"zip_code"`
	CreatedTime  time.Time `gorm:"column:CREATED_TIME" json:"created_time"`
	Age          int32     `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (e *Customer) TableName() string {
	return "Customers"
}
