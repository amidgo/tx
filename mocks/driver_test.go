package txmocks_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/amidgo/tx"
	txmocks "github.com/amidgo/tx/mocks"
)

func Test_Driver_Error_Valid_equal(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	expectedErr := io.ErrUnexpectedEOF
	returnErr := io.EOF

	driver := txmocks.ExpectDriverError(
		func(err, target error) bool { return err == target },
		expectedErr,
		returnErr,
	)(testReporter)

	err := driver.Error(expectedErr)
	if err != returnErr {
		t.Fatalf("unexpected error, expected %+v, actual %+v", returnErr, err)
	}
}

func Test_Driver_Error_Valid_errorsIs(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	expectedErr := fmt.Errorf("error: %w", io.ErrUnexpectedEOF)
	returnErr := io.EOF

	driver := txmocks.ExpectDriverError(
		errors.Is,
		expectedErr,
		returnErr,
	)(testReporter)

	err := driver.Error(expectedErr)
	if err != returnErr {
		t.Fatalf("unexpected error, expected %+v, actual %+v", returnErr, err)
	}
}

func Test_Driver_Error_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, driver.Error called more than once")

	expectedErr := fmt.Errorf("error: %w", io.ErrUnexpectedEOF)
	returnErr := io.EOF

	driver := txmocks.ExpectDriverError(
		errors.Is,
		expectedErr,
		returnErr,
	)(testReporter)

	err := driver.Error(expectedErr)
	if err != returnErr {
		t.Fatalf("unexpected error, expected %+v, actual %+v", returnErr, err)
	}

	err = driver.Error(expectedErr)
	if err != returnErr {
		t.Fatalf("unexpected error, expected %+v, actual %+v", returnErr, err)
	}
}

func Test_Driver_Error_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "no calls occured to driver.Error")

	expectedErr := fmt.Errorf("error: %w", io.ErrUnexpectedEOF)
	returnErr := io.EOF

	txmocks.ExpectDriverError(
		errors.Is,
		expectedErr,
		returnErr,
	)(testReporter)
}

func Test_Driver_Error_WrongError(t *testing.T) {
	expectedErr := fmt.Errorf("error: %w", io.ErrUnexpectedEOF)
	actualErr := io.ErrUnexpectedEOF
	returnErr := io.EOF

	tFatalMessage := fmt.Sprintf("unexpected error, expected %+v, actual %+v", expectedErr, actualErr)

	testReporter := newMockTestReporter(t, tFatalMessage)

	driver := txmocks.ExpectDriverError(
		func(err, target error) bool { return err == target },
		expectedErr,
		returnErr,
	)(testReporter)

	err := driver.Error(actualErr)
	if err != returnErr {
		t.Fatalf("unexpected error, expected %+v, actual %+v", returnErr, err)
	}
}

type joinDriversTest struct {
	Name          string
	DriverMocks   []txmocks.DriverMock
	F             func(t *testing.T, driver tx.Driver)
	TFatalMessage string
}

func (j *joinDriversTest) Test(t *testing.T) {
	testReporter := newMockTestReporter(t, j.TFatalMessage)

	driver := txmocks.JoinDrivers(j.DriverMocks...)(testReporter)

	if j.F != nil {
		j.F(t, driver)
	}
}

func Test_JoinDrivers(t *testing.T) {

	errorf := fmt.Errorf("error: %w", io.ErrUnexpectedEOF)

	tests := []*joinDriversTest{
		{
			Name:          "zero drivers",
			DriverMocks:   []txmocks.DriverMock{},
			F:             func(t *testing.T, driver tx.Driver) {},
			TFatalMessage: "",
		},
		{
			Name: "single driver mock",
			DriverMocks: []txmocks.DriverMock{
				txmocks.ExpectDriverError(errors.Is, io.ErrUnexpectedEOF, io.EOF),
			},
			F: func(t *testing.T, driver tx.Driver) {
				err := driver.Error(io.ErrUnexpectedEOF)
				if err != io.EOF {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.EOF, err)
				}
			},
		},
		{
			Name: "two driver mocks",
			DriverMocks: []txmocks.DriverMock{
				txmocks.ExpectDriverError(errors.Is, io.ErrUnexpectedEOF, io.EOF),
				txmocks.ExpectDriverError(
					func(err, target error) bool { return err == target },
					io.ErrNoProgress,
					io.ErrShortBuffer,
				),
			},
			F: func(t *testing.T, driver tx.Driver) {
				err := driver.Error(io.ErrUnexpectedEOF)
				if err != io.EOF {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.EOF, err)
				}

				err = driver.Error(io.ErrNoProgress)
				if err != io.ErrShortBuffer {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.ErrNoProgress, err)
				}
			},
		},
		{
			Name: "two driver mocks, but called 3 times",
			DriverMocks: []txmocks.DriverMock{
				txmocks.ExpectDriverError(errors.Is, io.ErrUnexpectedEOF, io.EOF),
				txmocks.ExpectDriverError(
					func(err, target error) bool { return err == target },
					io.ErrNoProgress,
					io.ErrShortBuffer,
				),
			},
			F: func(t *testing.T, driver tx.Driver) {
				err := driver.Error(io.ErrUnexpectedEOF)
				if err != io.EOF {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.EOF, err)
				}

				err = driver.Error(io.ErrNoProgress)
				if err != io.ErrShortBuffer {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.ErrNoProgress, err)
				}

				err = driver.Error(io.ErrNoProgress)
				if err != nil {
					t.Fatalf("expected no error, actual %+v", err)
				}
			},
			TFatalMessage: "unexpected call to driver.Error, no calls left",
		},
		{
			Name: "two driver mocks, second call with wrong error",
			DriverMocks: []txmocks.DriverMock{
				txmocks.ExpectDriverError(errors.Is, io.ErrUnexpectedEOF, io.EOF),
				txmocks.ExpectDriverError(
					func(err, target error) bool { return err == target },
					io.ErrNoProgress,
					io.ErrShortBuffer,
				),
			},
			F: func(t *testing.T, driver tx.Driver) {
				err := driver.Error(io.ErrUnexpectedEOF)
				if err != io.EOF {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.EOF, err)
				}

				err = driver.Error(errorf)
				if err != io.ErrShortBuffer {
					t.Fatalf("unexpected err, expected %+v, actual %+v", io.ErrNoProgress, err)
				}
			},
			TFatalMessage: fmt.Sprintf("unexpected error, expected %+v, actual %+v", io.ErrNoProgress, errorf),
		},
	}

	for _, tst := range tests {
		t.Run(tst.Name, tst.Test)
	}
}
