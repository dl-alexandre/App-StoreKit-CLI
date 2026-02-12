package validate

import "testing"

func TestNormalizeOne(t *testing.T) {
	value, err := NormalizeOne("transaction.history.version", "V2")
	if err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	if value != "v2" {
		t.Fatalf("expected v2, got %s", value)
	}
}

func TestNormalizeMany(t *testing.T) {
	values, err := NormalizeMany("transaction.history.productType", []string{"consumable", "NON_RENEWABLE"})
	if err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	if values[0] != "CONSUMABLE" || values[1] != "NON_RENEWABLE" {
		t.Fatalf("unexpected values: %v", values)
	}
}
