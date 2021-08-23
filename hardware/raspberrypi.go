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

func (r *RaspberryPi) ReadTemperature() (float64, error) {
	return r.SenseHat.ReadTemperature()
}

func (r *RaspberryPi) ReadHumidity() (float64, error) {
	return r.SenseHat.ReadHumidity()
}
