package test

import (
	"fmt"
	"os"
	"testing"

	go_ora "github.com/sijms/go-ora/v2"
	oracle "github.com/udnz/gorm-oracle"
	"gorm.io/gorm"
)

var dsn string

func init() {
	initDSN()
}

func initDSN() {
	// see: https://github.com/sijms/go-ora
	// CONN=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=***)(PORT=***))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=***)))
	// USER= ""
	// PASSWORD = ""

	// set connection time for 3 second
	urlOptions := map[string]string{
		"CONNECTION TIMEOUT": "3",
	}
	dsn = go_ora.BuildJDBC(os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("CONN"), urlOptions)
	fmt.Printf("==> DSN: %s\n", dsn)
}

func getDb(t *testing.T) *gorm.DB {
	dialector := oracle.Open(dsn)
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open Error %s", err)
	}
	err = db.Error
	if err != nil {
		t.Fatalf("getDb Error %s", err)
	}

	return db
}

func checkTxError(t *testing.T, tx *gorm.DB) *gorm.DB {
	if tx.Error != nil {
		t.Errorf("%s: tx.Error %s", t.Name(), tx.Error)
	}
	return tx
}
