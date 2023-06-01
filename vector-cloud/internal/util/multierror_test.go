package util

import (
	"errors"
	"testing"
)

func TestMultierror(t *testing.T) {
	var e Errors
	if e.Error() != nil {
		t.Fatalf("default error should be nil")
	}
	e.Append(nil)
	if e.Error() != nil {
		t.Fatalf("appending nil should still have nil error")
	}
	e.AppendMulti(nil, nil)
	if e.Error() != nil {
		t.Fatalf("appending nils should still have nil error")
	}
	known := errors.New("known")
	e.Append(known)
	if e.Error() != known {
		t.Fatalf("single error should equal contained error")
	}
	if e.Error().Error() != "known" {
		t.Fatalf("single error should have equal string to original")
	}
	e.Append(known)
	expected := "[{known}, {known}]"
	if e.Error().Error() != expected {
		t.Fatalf("unexpected error format: got %s, expected %s", e.Error().Error(), expected)
	}
	e = Errors{}
	e.AppendMulti(known, known)
	if e.Error().Error() != expected {
		t.Fatalf("unexpected error format: got %s, expected %s", e.Error().Error(), expected)
	}
}
