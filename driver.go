package tx

type Driver interface {
	Error(err error) error
}

func getDriver(provider Provider) (Driver, bool) {
	driver, ok := provider.(interface{ Driver() Driver })
	if !ok {
		return nil, false
	}

	return driver.Driver(), true
}

type driverProvider struct {
	Provider
	driver Driver
}

func (d *driverProvider) Driver() Driver {
	return d.driver
}

func DriverProvider(provider Provider, driver Driver) Provider {
	return &driverProvider{
		Provider: provider,
		driver:   driver,
	}
}
