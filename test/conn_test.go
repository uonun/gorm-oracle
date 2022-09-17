package test

import (
	"testing"
)

func TestOpen(t *testing.T) {
	db := getDb(t)
	if db != nil {
		t.Log("Test Open OK")
	}
}
