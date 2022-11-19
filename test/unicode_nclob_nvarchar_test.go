package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	go_ora "github.com/sijms/go-ora/v2"
	"gorm.io/gorm"
)

func TestInsertUnicodeRaw(t *testing.T) {
	db := getDb(t)
	value := "😁🍎🇨🇳㊣①❷㏾罗🀄︎🌈🔥"

	insertWithCheck_Raw(t, db, value, go_ora.NVarChar(value))
	insertWithCheck_Raw(t, db, value, go_ora.NClob{String: value, Valid: true})

	// make sure more than 4k bytes.
	value = strings.Repeat(value, 100)
	insertWithCheck_Raw(t, db, value, go_ora.NVarChar(value))
	insertWithCheck_Raw(t, db, value, go_ora.NClob{String: value, Valid: true})
}

func insertWithCheck_Raw(t *testing.T, db *gorm.DB, value string, val interface{}) {
	var customer_name = fmt.Sprintf("insertWithCheck_Raw:%s", uuid.New().String())
	var newid int64
	checkTxError(t,
		db.Exec(
			`INSERT INTO CUSTOMERS (customer_id,customer_name,address) VALUES (customers_s.nextval,:1,:2) RETURNING customer_id INTO :3`, customer_name, val, &newid))

	var newValue string
	checkTxError(t, db.Raw("SELECT ADDRESS FROM CUSTOMERS WHERE customer_id = :1", newid).Scan(&newValue))

	if string(value) != newValue {
		t.Errorf("unicode not inserted")
	}
}

func TestInsertUnicodeModel(t *testing.T) {
	db := getDb(t)
	value := "😁🍎🇨🇳㊣①❷㏾罗🀄︎🌈🔥"
	cNVarChar := getCustomer("TestInsertUnicodeModel")
	cNClob := getCustomerOfNClob("TestInsertUnicodeModel")

	cNVarChar.Address = go_ora.NVarChar(value)
	cNClob.Address = &go_ora.NClob{String: value, Valid: true}
	insertWithCheck_NVarCharModel(t, db, value, cNVarChar)
	insertWithCheck_NClobModel(t, db, value, cNClob)

	// make sure more than 4k bytes.
	value = strings.Repeat(value, 100)
	cNVarChar.Address = go_ora.NVarChar(value)
	cNClob.Address = &go_ora.NClob{String: value, Valid: true}
	insertWithCheck_NVarCharModel(t, db, value, cNVarChar)
	insertWithCheck_NClobModel(t, db, value, cNClob)
}

func insertWithCheck_NVarCharModel(t *testing.T, db *gorm.DB, value string, c Customer) {
	checkTxError(t, checkTxError(t, db.Create(&c)))

	var newValue string
	checkTxError(t, db.Raw("SELECT ADDRESS FROM CUSTOMERS WHERE customer_id = :1", c.CustomerID).Scan(&newValue))

	if string(value) != newValue {
		t.Errorf("unicode not inserted")
	}
}
func insertWithCheck_NClobModel(t *testing.T, db *gorm.DB, value string, c CustomerOfNClob) {
	checkTxError(t, checkTxError(t, db.Create(&c)))

	var newValue string
	checkTxError(t, db.Raw("SELECT ADDRESS FROM CUSTOMERS WHERE customer_id = :1", c.CustomerID).Scan(&newValue))

	if string(value) != newValue {
		t.Errorf("unicode not inserted")
	}
}
