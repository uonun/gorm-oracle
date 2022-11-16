package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func initDB(t *testing.T) {
	initDDL := getDDL(t, "./assets/init.sql")
	db := getDb(t)
	tx := db.Exec(initDDL)

	if tx.Error != nil {
		t.Fatalf("TestCreateTable Error %s", tx.Error)
	}
}

func getDDL(t *testing.T, name string) string {
	fpath, _ := filepath.Abs(name)
	data, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func TestCreateTable(t *testing.T) {
	fmt.Printf("1")
	// db := getDb(t)
	// tx := db.Exec(`CREATE TABLE customers
	// ( customer_id number(10) NOT NULL,
	// 	customer_name varchar2(50) NOT NULL,
	// 	city varchar2(50)
	// );`)

	// if tx.Error != nil {
	// 	t.Fatalf("TestCreateTable Error %s", tx.Error)
	// }
}
