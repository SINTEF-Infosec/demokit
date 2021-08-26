package hardware

import "fmt"

const UnavailableHardware = "hardware unavailable, this is a virtual hardware layer"

type VirtualHardwareLayer struct{}

func NewVirtualHardwareLayer() *VirtualHardwareLayer {
	return &VirtualHardwareLayer{}
}

func (v VirtualHardwareLayer) IsAvailable() bool {
	return false
}

func (v VirtualHardwareLayer) ReadTemperature() (float64, error) {
	return 0.0, fmt.Errorf(UnavailableHardware)
}

func (v VirtualHardwareLayer) ReadHumidity() (float64, error) {
	return 0.0, fmt.Errorf(UnavailableHardware)
}

func (v VirtualHardwareLayer) LightOn() error {
	return fmt.Errorf(UnavailableHardware)
}

func (v VirtualHardwareLayer) LightOff() error {
	return fmt.Errorf(UnavailableHardware)
}
