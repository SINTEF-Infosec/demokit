package hardware

import (
	log "github.com/sirupsen/logrus"
)

type VirtualHardwareLayer struct{}

func NewVirtualHardwareLayer() *VirtualHardwareLayer {
	return &VirtualHardwareLayer{}
}

func (v VirtualHardwareLayer) IsAvailable() bool {
	return false
}

func (v VirtualHardwareLayer) Init() {}

func (v VirtualHardwareLayer) SetEventHandler(_ func(interface{})) {}

func (v VirtualHardwareLayer) SetLogger(_ *log.Entry) {}
