package localkv

import (
	"testing"
)

var databasePath = "../../user_data"

func TestLocalKV(t *testing.T) {
	kv, err := NewLocalKV(&databasePath)
	if err != nil {
		t.Fatal(err)
	}

	if kv.Set("hello", "world"); err != nil {
		t.Fatal(err)
	}

	val, err := kv.Get("hello")
	if err != nil {
		t.Fatal(err)
	}

	if val != "world" {
		t.Fatal("expected value doesn't match")
	}

	if kv.Close(); err != nil {
		t.Fatal(err)
	}

	if err = kv.RemoveDB(); err != nil {
		t.Fatal(err)
	}
}
