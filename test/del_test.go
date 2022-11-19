package test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestDelelteModel(t *testing.T) {
	var customer_name = fmt.Sprintf("TestDelelteModel:%s", uuid.New().String())

	db := getDb(t)

	var newid int64
	checkTxError(t,
		db.Exec(
			`INSERT INTO CUSTOMERS (customer_id,customer_name) VALUES (customers_s.nextval,:1) RETURNING customer_id INTO :2`, customer_name, &newid))

	c := CustomerWithSequenceButNotReturning{CustomerID: newid}

	ret := checkTxError(t, db.Where("customer_id = ? ", c.CustomerID).Delete(&c))
	if ret.RowsAffected != 1 {
		t.Fatalf("TestDelelteModel: %d rows deleted.", ret.RowsAffected)
	}
}
