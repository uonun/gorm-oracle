package test

import (
	"testing"
)

func TestQueryRawCount(t *testing.T) {
	db := getDb(t)

	var count int
	tx := db.Raw("SELECT COUNT(*) FROM CUSTOMERS").Scan(&count)

	err := tx.Error
	if err != nil {
		t.Errorf("tx.Error %s", err)
	}
}

func TestQueryRawModel(t *testing.T) {
	db := getDb(t)

	var row = CustomerWithSequenceButNotReturning{}
	checkTxError(t, db.Raw("SELECT customer_id,customer_name FROM CUSTOMERS").Scan(&row))

	if row.CustomerID == 0 {
		t.Errorf("TestQueryRawModel no row return")
	}
}

func TestQueryModel(t *testing.T) {
	db := getDb(t)

	var row CustomerWithSequenceButNotReturning
	checkTxError(t, db.Find(&row))

	if row.CustomerID == 0 {
		t.Errorf("TestQueryRawModel no row return")
	}
}
