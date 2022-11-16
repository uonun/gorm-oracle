package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestUpdateRawExists(t *testing.T) {
	c, db := createOne(t)

	customer_name := uuid.New().String()
	tx := checkTxError(t,
		db.Exec(`UPDATE CUSTOMERS SET customer_name=? WHERE customer_id=? `,
			customer_name, c.CustomerID))

	if tx.RowsAffected < 1 {
		t.Errorf("there must be 1 row updated.")
	}

	var updatedC Customer
	tx = checkTxError(t, db.
		Where("customer_name=? and customer_id=?", customer_name, c.CustomerID).
		Find(&updatedC))

	if tx.RowsAffected < 1 {
		t.Errorf("there must be 1 row found after updated.")
	}

	if updatedC.CustomerID != c.CustomerID || updatedC.CustomerName != customer_name {
		t.Errorf("row not updated.")
	}
}

func TestUpdateRawNoneExists(t *testing.T) {

	var customer_name = "customer_name2"
	var address = "address1"

	db := getDb(t)

	// NOTE: Anonymous parameters only, passed by order
	tx := checkTxError(t,
		db.Exec(`UPDATE CUSTOMERS SET customer_name=?, address=? WHERE 1=2 `,
			customer_name, address))

	if tx.RowsAffected > 0 {
		t.Errorf("should not update any row.")
	}
}

func TestUpdateModel(t *testing.T) {
	c, db := createOne(t)

	c.CustomerName = uuid.New().String()
	c.Age = rand.Int31()

	tx := checkTxError(t, db.Updates(&c))
	if tx.RowsAffected < 1 {
		t.Errorf("there must be 1 row updated.")
	}

	var updatedC Customer
	tx = checkTxError(t, db.
		Where("customer_name=? and age=? and customer_id=?", c.CustomerName, c.Age, c.CustomerID).
		Find(&updatedC))

	if tx.RowsAffected < 1 {
		t.Errorf("there must be 1 row found after updated.")
	}

	if updatedC.CustomerID != c.CustomerID ||
		updatedC.CustomerName != c.CustomerName ||
		updatedC.Age != c.Age {
		t.Errorf("row not updated.")
	}
}

func TestUpdateModels(t *testing.T) {
	// create N rows
	count := 4
	batchId := uuid.NewString()
	cs := make([]CustomerReturning, count)
	for i := 0; i < count; i++ {
		cs[i] = getRandomCustomerReturning(fmt.Sprintf("TestUpdateModels:batch-%s:", batchId))
	}

	db := getDb(t)
	checkTxError(t, db.Create(&cs))

	// update them
	newState := uuid.New().String()
	newTime := time.Now()
	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		ids[i] = cs[i].CustomerID
	}

	tx2 := checkTxError(t, db.Where("customer_id IN ?", ids).Updates(Customer{State: newState, CreatedTime: newTime}))
	if tx2.RowsAffected != int64(count) {
		t.Errorf("there must be %d rows updated.", count)
	}

	var updatedCs []Customer
	tx3 := checkTxError(t, db.Where("state=? and created_time=?", newState, newTime).Find(&updatedCs))
	if tx3.RowsAffected != int64(count) || len(updatedCs) != count {
		t.Errorf("there must be %d rows found after updated.", count)
	}

	contains := func(elems []int64, elem int64) bool {
		for _, e := range ids {
			if elem == e {
				return true
			}
		}
		return false
	}

	for i := 0; i < count; i++ {
		if !contains(ids, updatedCs[i].CustomerID) {
			t.Errorf("it must be the created rows which updated.")
		}
	}
}

// createOne create a row to be updated.
func createOne(t *testing.T) (CustomerReturningPrimaryKey, *gorm.DB) {
	c := getRandomCustomerReturningPrimaryKey("createForUpdateTesting")
	db := getDb(t)
	checkTxError(t, db.Create(&c))
	return c, db
}
