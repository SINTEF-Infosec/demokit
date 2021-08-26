package hardware

// WIP
type RaspberryPi struct {
	SenseHat *SenseHat
}

func NewRaspberryPiWithSenseHat() *RaspberryPi {
	return &RaspberryPi{
		SenseHat: NewSenseHat(),
	}
}

func (r *RaspberryPi) IsAvailable() bool {
	return true
}

func (r *RaspberryPi) ReadTemperature() (float64, error) {
	return r.SenseHat.ReadTemperature()
}

func (r *RaspberryPi) ReadHumidity() (float64, error) {
	return r.SenseHat.ReadHumidity()
}

func (r *RaspberryPi) LightOn() error {
	return r.SenseHat.LightOn()
}

func (r *RaspberryPi) LightOff() error {
	return r.SenseHat.LightOff()
}
