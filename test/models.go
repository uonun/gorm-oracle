package test

// Customer table comment
type Customer struct {
	CustomerName string `gorm:"column:CUSTOMER_NAME" json:"customer_name"`
	CustomerID   int64  `gorm:"column:CUSTOMER_ID;primary_key;sequence:CUSTOMERS_S" json:"customer_id"`
	Address      string `gorm:"column:ADDRESS" json:"address"`
	City         string `gorm:"column:CITY" json:"city"`
	State        string `gorm:"column:STATE" json:"state"`
	ZipCode      string `gorm:"column:ZIP_CODE" json:"zip_code"`
	Age          int32  `gorm:"column:AGE" json:"age"`
}

// TableName sets the insert table name for this struct type
func (e *Customer) TableName() string {
	return "Customers"
}
