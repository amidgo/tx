package tx

import "testing"

func Test_driverBeginner(t *testing.T) {
	beginner := &driverBeginner{}

	_, ok := getDriver(beginner)
	if !ok {
		t.Fatal("driver beginner doesn't implements Driver() Driver method")
	}
}

func Test_driverTx(t *testing.T) {
	tx := &driverTx{}

	_, ok := getDriver(tx)
	if !ok {
		t.Fatal("driver tx doesn't implements Driver() Driver method")
	}
}
