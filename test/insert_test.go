package test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
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
	tx := checkTxError(t, db.Create(&c))

	var rowsExpected int64 = 1
	if tx.RowsAffected != rowsExpected {
		t.Errorf("%d rows affected, %d expected", tx.RowsAffected, rowsExpected)
	}
}

func TestInsertModels(t *testing.T) {
	count := 4
	batchId := uuid.NewString()
	cs := make([]Customer, count)
	for i := 0; i < count; i++ {
		cs[i] = getRandomCustomer(fmt.Sprintf("TestInsertModels:batch-%s:", batchId))
	}

	db := getDb(t)
	tx := checkTxError(t, db.Create(&cs).Clauses(clause.Returning{}))
	if tx.RowsAffected != int64(count) {
		t.Errorf("batch insert %d rows affected, %d expected", tx.RowsAffected, count)
	}
}

func TestInsertModelsWithReturningClause(t *testing.T) {
	count := 4
	batchId := uuid.NewString()
	cs := make([]Customer, count)
	for i := 0; i < count; i++ {
		cs[i] = getRandomCustomer(fmt.Sprintf("TestInsertModels:batch-%s:", batchId))
	}

	db := getDb(t)
	tx := checkTxError(t, db.Clauses(clause.Returning{
		Columns: []clause.Column{
			{Name: "CUSTOMER_ID"},
		},
	}).Create(&cs))

	// NOT SUPPORTED NOW
	// var rowsExpected = int64(count)
	// if tx.RowsAffected != rowsExpected {
	// 	t.Errorf("batch insert %d rows affected, %d expected", tx.RowsAffected, rowsExpected)
	// }

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		cid := cs[i].CustomerID
		ids[i] = strconv.FormatInt(cid, 10)
		if cid == 0 {
			t.Errorf("not returning created value: %s", tx.Error)
		}
	}
	t.Logf("created: %s", strings.Join(ids, ","))
}

func TestInsertModelsWithReturningClause2(t *testing.T) {
	count := 4
	batchId := uuid.NewString()
	cs := make([]CustomerHook, count)
	for i := 0; i < count; i++ {
		cs[i] = getRandomCustomerHook(fmt.Sprintf("TestInsertModels:batch-%s:", batchId))
	}

	db := getDb(t)
	tx := checkTxError(t, db.Clauses(clause.Returning{
		Columns: []clause.Column{
			{Name: "CUSTOMER_ID"},
			// {Name: "STATE"}, // not need returning this column
		},
	}).Create(&cs))

	// NOT SUPPORTED NOW
	// var rowsExpected = int64(count)
	// if tx.RowsAffected != rowsExpected {
	// 	t.Errorf("batch insert %d rows affected, %d expected", tx.RowsAffected, rowsExpected)
	// }

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		cid := cs[i].CustomerID
		ids[i] = strconv.FormatInt(cid, 10)
		if cid == 0 || !strings.HasPrefix(cs[i].State, "HOOK") {
			t.Errorf("not returning created value: %s", tx.Error)
		}
	}
	t.Logf("created: %s", strings.Join(ids, ","))
}

func TestInsertReturningModel(t *testing.T) {
	c := getRandomCustomerReturning("TestInsertModelReturning")

	db := getDb(t)
	checkTxError(t, db.Create(&c))

	// NOT SUPPORTED NOW
	// var rowsExpected int64 = 1
	// if tx.RowsAffected != rowsExpected {
	// 	t.Errorf("%d rows affected, %d expected", tx.RowsAffected, rowsExpected)
	// }

	if c.CustomerID <= 0 {
		t.Errorf("create failed or not returning new id")
	}

	t.Logf("created: %d", c.CustomerID)
}

func TestInsertReturningModels(t *testing.T) {
	count := 4
	batchId := uuid.NewString()
	cs := make([]CustomerReturning, count)
	for i := 0; i < count; i++ {
		cs[i] = getRandomCustomerReturning(fmt.Sprintf("TestInsertModelsReturning:batch-%s:", batchId))
	}

	db := getDb(t)
	tx := checkTxError(t, db.Create(&cs))

	// NOT SUPPORTED NOW
	// if tx.RowsAffected != int64(count) {
	// 	t.Errorf("batch insert %d rows affected, %d expected", tx.RowsAffected, count)
	// }

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		cid := cs[i].CustomerID
		ids[i] = strconv.FormatInt(cid, 10)
		if cid == 0 {
			t.Errorf("not returning created value: %s", tx.Error)
		}
	}
	t.Logf("created: %s", strings.Join(ids, ","))
}
