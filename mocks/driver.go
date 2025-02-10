package txmocks

import (
	"sync"
	"sync/atomic"
)

type driverAsserter interface {
	error(err error) error
	assert()
}

type Driver struct {
	asrt driverAsserter
}

func newDriver(t testReporter, asrt driverAsserter) *Driver {
	t.Cleanup(asrt.assert)

	return &Driver{asrt: asrt}
}

func (d *Driver) Error(err error) error {
	return d.asrt.error(err)
}

type DriverMock func(t testReporter) *Driver

func ExpectDriverError(
	errorMatches func(err, target error) bool,
	expectedErr, returnErr error,
) DriverMock {
	return func(t testReporter) *Driver {
		asrt := &driverError{
			t:            t,
			errorMatches: errorMatches,
			expectedErr:  expectedErr,
			returnErr:    returnErr,
		}

		return newDriver(t, asrt)
	}
}

type driverError struct {
	t      testReporter
	called atomic.Bool

	errorMatches func(err, target error) bool
	expectedErr  error
	returnErr    error
}

func (d *driverError) error(err error) error {
	swapped := d.called.CompareAndSwap(false, true)
	if !swapped {
		d.t.Fatal("unexpected call, driver.Error called more than once")
	}

	if !d.errorMatches(err, d.expectedErr) {
		d.t.Fatalf("unexpected error, expected %+v, actual %+v", d.expectedErr, err)
	}

	return d.returnErr
}

func (d *driverError) assert() {
	called := d.called.Load()

	if !called {
		d.t.Fatal("no calls occured to driver.Error")
	}
}

type nilDriver struct {
	t testReporter
}

func NilDriver(t testReporter) *Driver {
	asrt := &nilDriver{
		t: t,
	}

	return newDriver(t, asrt)
}

func (d nilDriver) error(err error) error {
	d.t.Fatal("unexpected call to driver.Error, expect no calls")

	return nil
}

func (_ nilDriver) assert() {}

func JoinDrivers(drivers ...DriverMock) DriverMock {
	return func(t testReporter) *Driver {
		switch len(drivers) {
		case 0:
			return NilDriver(t)
		case 1:
			return drivers[0](t)
		}

		asrts := make([]driverAsserter, len(drivers))

		for i := range drivers {
			index := len(drivers) - 1 - i

			prv := drivers[index](t)

			if prv.asrt == nil {
				t.Fatalf("invalid driver by index %d, Driver.asrt is nil", index)

				return nil
			}

			asrts[index] = prv.asrt
		}

		asrt := &driverAssertersJoin{
			t:     t,
			asrts: asrts,
		}

		return newDriver(t, asrt)
	}
}

type driverAssertersJoin struct {
	t            testReporter
	asrts        []driverAsserter
	currentIndex int
	mu           sync.Mutex
}

func (d *driverAssertersJoin) error(err error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	asrt, expected := d.currentAsserter()
	if !expected {
		d.t.Fatal("unexpected call to driver.Error, no calls left")

		return nil
	}

	asrtErr := asrt.error(err)

	d.currentIndex++

	return asrtErr
}

func (d *driverAssertersJoin) currentAsserter() (driverAsserter, bool) {
	if d.currentIndex > len(d.asrts)-1 {
		return nil, false
	}

	return d.asrts[d.currentIndex], true
}

func (d *driverAssertersJoin) assert() {
	for _, asrt := range d.asrts {
		asrt.assert()
	}
}
