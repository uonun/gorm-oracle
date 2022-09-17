package test

import (
	"math/rand"
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
		db.Exec("INSERT INTO CUSTOMERS VALUES (customers_s.nextval,:1,:2,:3,:4,:5,:6)",
			customer_name, address, city, state, zip, age))

}

func TestInsertModel(t *testing.T) {
	c := Customer{
		CustomerName: uuid.New().String(),
		Address:      uuid.New().String(),
		City:         uuid.New().String(),
		Age:          rand.Int31(),
	}

	db := getDb(t)
	checkTxError(t, db.Create(&c))
}

// not support yet
// func TestInsertModels(t *testing.T) {
// 	count := 10
// 	cs := make([]Customer, count)
// 	for i := 0; i < count; i++ {
// 		cs[i] = Customer{
// 			CustomerName: uuid.New().String(),
// 			Address:      uuid.New().String(),
// 			City:         uuid.New().String(),
// 			Age:          rand.Int31(),
// 		}
// 	}

// 	db := getDb(t)
// 	tx := checkTxError(t, db.Create(&cs))
// 	if tx.RowsAffected != int64(count) {
// 		t.Errorf("batch insert affected %d rows, %d expected", tx.RowsAffected, count)
// 	}
// }
