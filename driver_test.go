package tx

import "testing"

func Test_driverBeginner(t *testing.T) {
	beginner := &driverBeginner{}

	_, ok := getDriver(beginner)
	if !ok {
		t.Fatal("driver beginner doesn't implements Driver() Driver method")
	}
}
