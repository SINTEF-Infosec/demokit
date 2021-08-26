package hardware

// Hal (Hardware Abstraction Layer) provides access to hardware functionalities
type Hal interface {
	EnvironmentReader
	Light
	IsAvailable() bool
}

type EnvironmentReader interface {
	TemperatureReader
	HumidityReader
}

type TemperatureReader interface {
	ReadTemperature() (float64, error)
}

type HumidityReader interface {
	ReadHumidity() (float64, error)
}

type Light interface {
	LightOn() error
	LightOff() error
}
