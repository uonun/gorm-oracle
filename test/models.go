package test

// Customer table comment
type Customer struct {
	CustomerID   int64  `gorm:"primary_key;sequence:customers_s;column:customer_id" json:"customer_id"`
	CustomerName string `gorm:"column:customer_name" json:"customer_name"`
	Address      string `gorm:"column:address" json:"address"`
	City         string `gorm:"column:city" json:"city"`
	State        string `gorm:"column:state" json:"state"`
	ZipCode      string `gorm:"column:zip_code" json:"zip_code"`
	Age          int32  `gorm:"column:age" json:"age"`
}

// TableName sets the insert table name for this struct type
func (e *Customer) TableName() string {
	return "Customers"
}
