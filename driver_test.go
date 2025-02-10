package tx

import "testing"

func Test_driverProvider(t *testing.T) {
	provider := &driverProvider{}

	_, ok := getDriver(provider)
	if !ok {
		t.Fatal("driver provider doesn't implements Driver() Driver method")
	}
}
