package store

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func jsonEquivalent(left, right json.RawMessage) (bool, error) {
	var leftValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		return false, fmt.Errorf("decode left JSON: %w", err)
	}

	var rightValue any
	if err := json.Unmarshal(right, &rightValue); err != nil {
		return false, fmt.Errorf("decode right JSON: %w", err)
	}

	return reflect.DeepEqual(leftValue, rightValue), nil
}

func TestJSONEquivalentIgnoresWhitespaceAndObjectKeyOrder(t *testing.T) {
	left := json.RawMessage(`{"version":1,"metadata":{"enabled":true,"labels":["a","b"]}}`)
	right := json.RawMessage(`{
		"metadata": {"labels": ["a", "b"], "enabled": true},
		"version": 1
	}`)

	equal, err := jsonEquivalent(left, right)
	if err != nil {
		t.Fatal(err)
	}
	if !equal {
		t.Fatalf("expected equivalent JSON: left %s right %s", left, right)
	}
}

func TestJSONEquivalentRejectsDifferentValues(t *testing.T) {
	left := json.RawMessage(`{"version":1,"enabled":true}`)
	right := json.RawMessage(`{"enabled":false,"version":1}`)

	equal, err := jsonEquivalent(left, right)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Fatalf("expected different JSON: left %s right %s", left, right)
	}
}
