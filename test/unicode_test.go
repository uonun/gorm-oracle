package test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	go_ora "github.com/sijms/go-ora/v2"
)

func TestInsertUnicodeRaw(t *testing.T) {
	db := getDb(t)
	var customer_name = fmt.Sprintf("TestInsertUnicode:%s", uuid.New().String())
	value := fmt.Sprintf("unicode:😁🍎$➔®≧①◎◉§❤️🇨🇳:%s", uuid.New().String())

	var newid int64
	checkTxError(t,
		db.Exec(
			`INSERT INTO CUSTOMERS (customer_id,customer_name,address) VALUES (customers_s.nextval,:1,:2) RETURNING customer_id INTO :3`, customer_name, go_ora.NVarChar(value), &newid))

	var newValue string
	checkTxError(t, db.Raw("SELECT ADDRESS FROM CUSTOMERS WHERE customer_id = :1", newid).Scan(&newValue))

	if string(value) != newValue {
		t.Errorf("unicode not inserted")
	}
}

func TestInsertUnicodeModel(t *testing.T) {
	c := getRandomCustomerReturning("TestInsertUnicodeModel")
	value := fmt.Sprintf("unicode:😁🍎$➔®≧①◎◉§❤️🇨🇳:%s", uuid.New().String())
	c.Address = go_ora.NVarChar(value)
	db := getDb(t)
	checkTxError(t, checkTxError(t, db.Create(&c)))

	var newValue string
	checkTxError(t, db.Raw("SELECT ADDRESS FROM CUSTOMERS WHERE customer_id = :1", c.CustomerID).Scan(&newValue))

	if string(value) != newValue {
		t.Errorf("unicode not inserted")
	}
}
