package test

import (
	"testing"
)

func TestQueryRaw(t *testing.T) {
	db := getDb(t)

	var count int
	tx := db.Raw("SELECT COUNT(*) FROM CUSTOMERS").Scan(&count)

	err := tx.Error
	if err != nil {
		t.Errorf("tx.Error %s", err)
	}
}
