package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTxInsertRaw(t *testing.T) {

	var customer_name1 = fmt.Sprintf("TestTxInsertRaw:row1:%s", uuid.New().String())
	var customer_name2 = fmt.Sprintf("TestTxInsertRaw:row2:%s", uuid.New().String())
	var address = "address1"
	var date = time.Now()
	var age = 101
	var sql = `INSERT INTO CUSTOMERS (customer_id,customer_name,address,created_time,age) 
		VALUES (CUSTOMERS_S.nextval,:1,:2,:3,:4)`
	var vals1 = []interface{}{customer_name1, address, date, age}
	var vals2 = []interface{}{customer_name2, address, date, age}

	db := getDb(t).Begin()
	defer func() {
		db.Rollback()
	}()

	// insert customer_name1
	checkTxError(t, db.Exec(sql, vals1...))

	db.SavePoint("point1")

	// insert customer_name2
	checkTxError(t, db.Exec(sql, vals2...))

	// query, there must be 2 rows found.
	var row1 CustomerWithSequenceButNotReturning
	row1ret := checkTxError(t, db.Where("CUSTOMER_NAME = ? ", customer_name1).Find(&row1))
	if row1ret.RowsAffected != 1 || row1.CustomerName != customer_name1 {
		t.Fatalf("TestTxInsertRaw: can not query inserted row-customer_name1 in tx")
	}
	var row2 CustomerWithSequenceButNotReturning
	row2ret := checkTxError(t, db.Where("CUSTOMER_NAME = ? ", customer_name2).Find(&row2))
	if row2ret.RowsAffected != 1 || row2.CustomerName != customer_name2 {
		t.Fatalf("TestTxInsertRaw: can not query inserted row-customer_name2 in tx")
	}

	// rollback
	db.RollbackTo("point1")

	// customer_name2 commited.
	db.Commit()

	newdb := getDb(t)
	var commitedRow CustomerWithSequenceButNotReturning
	// query again. must be 1 row only
	newdbret := checkTxError(t, newdb.Where("CUSTOMER_NAME = ? or CUSTOMER_NAME = ? ", customer_name1, customer_name2).Find(&commitedRow))
	if newdbret.RowsAffected != 1 || commitedRow.CustomerName != customer_name1 {
		t.Fatalf("TestTxInsertRaw: 2 row inserted after rollback")
	}

}
