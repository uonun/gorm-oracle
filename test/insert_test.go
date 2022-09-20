package test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestInsertRaw(t *testing.T) {

	var customer_name = "customer_name2"
	var address = "address1"
	var city = "city1"
	var state = "state1"
	var zip = "zip1"
	var age = 10

	db := getDb(t)

	// NOTE: Anonymous parameters only, passed by order
	checkTxError(t,
		db.Exec(`INSERT INTO CUSTOMERS 
				(customer_id,customer_name,address,city,state,zip_code,age) VALUES 
				(customers_s.nextval,:1,:2,:3,:4,:5,:6)`,
			customer_name, address, city, state, zip, age))

}

func TestInsertModel(t *testing.T) {
	c := getRandomCustomer("TestInsertModel")

	db := getDb(t)
	checkTxError(t, db.Create(&c))
}

func TestInsertModels(t *testing.T) {
	count := 4
	batchId := uuid.NewString()
	cs := make([]Customer, count)
	for i := 0; i < count; i++ {
		cs[i] = getRandomCustomer(fmt.Sprintf("TestInsertModels:batch-%s:", batchId))
	}

	db := getDb(t)
	tx := checkTxError(t, db.Create(&cs))
	if tx.RowsAffected != int64(count) {
		t.Errorf("batch insert affected %d rows, %d expected", tx.RowsAffected, count)
	}
}
