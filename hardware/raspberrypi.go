package hardware

import log "github.com/sirupsen/logrus"

// WIP
type RaspberryPi struct {
	SenseHat *SenseHat
}

func NewRaspberryPiWithSenseHat() *RaspberryPi {
	senseHat, err := NewSenseHat()
	if err != nil {
		log.Fatalf("Could not instantiate new sense hat: %v", err)
	}
	return &RaspberryPi{
		SenseHat: senseHat,
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
