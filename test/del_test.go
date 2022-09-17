package test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestDelelteModel(t *testing.T) {
	var customer_name = fmt.Sprintf("TestDelelteModel:%s", uuid.New().String())

	db := getDb(t)

	checkTxError(t,
		db.Exec("INSERT INTO CUSTOMERS (customer_id,customer_name) VALUES (customers_s.nextval,:1)", customer_name))

	c := Customer{CustomerName: customer_name}

	ret := checkTxError(t, db.Where("customer_name = ? ", customer_name).Delete(&c))
	if ret.RowsAffected != 1 {
		t.Fatalf("TestDelelteModel: %d rows deleted.", ret.RowsAffected)
	}
}
