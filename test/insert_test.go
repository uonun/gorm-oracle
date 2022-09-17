package test

import (
	"testing"
)

func TestInsertRaw(t *testing.T) {

	var customer_name = "customer_name2"
	var address = "address1"
	var city = "city1"
	var state = "state1"
	var zip = "zip1"
	var age = 10

	db := getDb(t)

	// NOTE: Anonymous parameters only, passed by order
	tx := db.Exec("INSERT INTO CUSTOMERS VALUES (customers_s.nextval,:1,:2,:3,:4,:5,:6)", customer_name, address, city, state, zip, age)

	err := tx.Error
	if err != nil {
		t.Errorf("tx.Error %s", err)
	}
}

func TestInsertModel(t *testing.T) {
	c := Customer{
		CustomerName: "cname1112222",
		Address:      "address2222",
		City:         "city2222",
		State:        "state2222",
		ZipCode:      "zipcode",
		Age:          324234,
	}

	db := getDb(t)
	tx := db.Create(&c)
	err := tx.Error
	if err != nil {
		t.Errorf("tx.Error %s", err)
	}
}
