package mocks_test

import (
	"errors"
	"reflect"
	"testing"
)

func requireEqual(t *testing.T, expected, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expected %+v not equal %+v", expected, actual)
	}
}

func requireTrue(t *testing.T, v bool) {
	if !v {
		t.Fatal("expected TRUE, actual FALSE")
	}
}

func requireFalse(t *testing.T, v bool) {
	if v {
		t.Fatal("expected FALSE, actual TRUE")
	}
}

func requireNil(t *testing.T, v any) {
	if !isNil(v) {
		t.Fatalf("expected Nil, actual %+v", v)
	}
}

func requireNotNil(t *testing.T, v any) {
	if isNil(v) {
		t.Fatal("expected not nil, actual nil", v)
	}
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	switch value.Kind() {
	case
		reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice, reflect.UnsafePointer:

		return value.IsNil()
	}

	return false
}

func requireErrorIs(t *testing.T, err error, target error) {
	if !errors.Is(err, target) {
		t.Fatalf("expected %s actual %s", target, err)
	}
}

func requireNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("expected no error, actual %s", err)
	}
}
