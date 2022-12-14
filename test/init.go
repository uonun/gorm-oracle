package test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	go_ora "github.com/sijms/go-ora/v2"
	oracle "github.com/uonun/gorm-oracle"
	"gorm.io/gorm"
)

var dsn string

func init() {
	initDSN()
	// initDB(&testing.T{})
}

func initDSN() {
	// see: https://github.com/sijms/go-ora
	// CONN=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=***)(PORT=***))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=***)))
	// USER= ""
	// PASSWORD = ""

	// set connection time for 3 second
	urlOptions := map[string]string{
		"Connection Timeout": "60",
		"Pooling":            "true",
		"Max Pool Size":      "40",
		"Min Pool Size":      "2",
		"Incr Pool Size":     "5",
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

	return db.Debug()
}

func getCustomerWithSequenceButNotReturning(name string) CustomerWithSequenceButNotReturning {
	return CustomerWithSequenceButNotReturning{
		CustomerName: fmt.Sprintf("%s:%s", name, uuid.New().String()),
		Address:      go_ora.NVarChar(fmt.Sprintf("Address:%s", uuid.New().String())),
		City:         fmt.Sprintf("City:%s", uuid.New().String()),
		State:        fmt.Sprintf("State:%d", rand.Int31()),
		ZipCode:      fmt.Sprintf("Z:%d", rand.Intn(9999)),
		CreatedTime:  time.Now(),
		Age:          rand.Int31(),
	}
}

func getCustomer(name string) Customer {
	return Customer{
		CustomerName: fmt.Sprintf("%s:%s", name, uuid.New().String()),
		Address:      go_ora.NVarChar(fmt.Sprintf("Address:%s", uuid.New().String())),
		City:         fmt.Sprintf("City:%s", uuid.New().String()),
		State:        fmt.Sprintf("State:%d", rand.Int31()),
		ZipCode:      fmt.Sprintf("Z:%d", rand.Intn(9999)),
		CreatedTime:  time.Now(),
		Age:          rand.Int31(),
	}
}

func getCustomerOfUDT(name string) CustomerOfUDT {
	return CustomerOfUDT{
		CustomerName: fmt.Sprintf("%s:%s", name, uuid.New().String()),
		// AddressNClob: &ContentNClob{String: fmt.Sprintf("Address:%s", uuid.New().String())},
		Address:     &NClobContent{String: go_ora.NClob{String: fmt.Sprintf("Address:%s", uuid.New().String()), Valid: true}},
		City:        fmt.Sprintf("City:%s", uuid.New().String()),
		State:       fmt.Sprintf("State:%d", rand.Int31()),
		ZipCode:     fmt.Sprintf("Z:%d", rand.Intn(9999)),
		CreatedTime: time.Now(),
		Age:         rand.Int31(),
	}
}

func getCustomerOfNClob(name string) CustomerOfNClob {
	return CustomerOfNClob{
		CustomerName: fmt.Sprintf("%s:%s", name, uuid.New().String()),
		Address:      &go_ora.NClob{String: fmt.Sprintf("Address:%s", uuid.New().String()), Valid: true},
		City:         fmt.Sprintf("City:%s", uuid.New().String()),
		State:        fmt.Sprintf("State:%d", rand.Int31()),
		ZipCode:      fmt.Sprintf("Z:%d", rand.Intn(9999)),
		CreatedTime:  time.Now(),
		Age:          rand.Int31(),
	}
}

func getCustomerWithPrimaryKey(name string) CustomerWithPrimaryKey {
	return CustomerWithPrimaryKey{
		CustomerName: fmt.Sprintf("%s:%s", name, uuid.New().String()),
		Address:      go_ora.NVarChar(fmt.Sprintf("Address:%s", uuid.New().String())),
		City:         fmt.Sprintf("City:%s", uuid.New().String()),
		State:        fmt.Sprintf("State:%d", rand.Int31()),
		ZipCode:      fmt.Sprintf("Z:%d", rand.Intn(9999)),
		CreatedTime:  time.Now(),
		Age:          rand.Int31(),
	}
}

func getCustomerWithHook(name string) CustomerWithHook {
	return CustomerWithHook{
		CustomerName: fmt.Sprintf("%s:%s", name, uuid.New().String()),
		Address:      go_ora.NVarChar(fmt.Sprintf("Address:%s", uuid.New().String())),
		City:         fmt.Sprintf("City:%s", uuid.New().String()),
		State:        fmt.Sprintf("State:%d", rand.Int31()),
		ZipCode:      fmt.Sprintf("Z:%d", rand.Intn(9999)),
		CreatedTime:  time.Now(),
		Age:          rand.Int31(),
	}
}

func checkTxError(t *testing.T, tx *gorm.DB) *gorm.DB {
	if tx.Error != nil {
		t.Errorf("%s: tx.Error %s", t.Name(), tx.Error)
	}
	return tx
}
