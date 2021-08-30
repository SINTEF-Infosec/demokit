// +build native_hardware

package hardware

import (
	_ "embed"
	"fmt"
	"github.com/DataDog/go-python3"
)

const (
	RaspberryModuleName     = "rpi"
	ReadTemperatureFunction = "read_temperature"
	ReadHumidityFunction    = "read_humidity"
	LightOn                 = "light_on"
	LightOff                = "light_off"
)

type SenseHat struct {
}

func NewSenseHat() (*SenseHat, error) {
	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		fmt.Println("Error initializing the python interpreter")
		return nil, fmt.Errorf("error initializing the interpreter")
	}
	return &SenseHat{}, nil
}

func (s *SenseHat) ReadHumidity() (float64, error) {
	return executePythonCodeWithNoArgumentsAndFloatReturn(RaspberryModuleName, ReadHumidityFunction)
}

func (s *SenseHat) ReadTemperature() (float64, error) {
	return executePythonCodeWithNoArgumentsAndFloatReturn(RaspberryModuleName, ReadTemperatureFunction)
}

func (s *SenseHat) LightOn() error {
	return executePythonCodeWithNoArgumentAndNoReturnValue(RaspberryModuleName, LightOn)
}

func (s *SenseHat) LightOff() error {
	return executePythonCodeWithNoArgumentAndNoReturnValue(RaspberryModuleName, LightOff)
}

func executePythonCodeWithNoArgumentsAndFloatReturn(moduleName, functionName string) (float64, error) {
	module := python3.PyImport_ImportModule(moduleName)
	if module == nil {
		return 0.0, fmt.Errorf("could not load module %s", moduleName)
	}

	function := module.GetAttrString(functionName)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			ret := function.Call(nilArgs, nil)
			float := python3.PyFloat_AsDouble(ret)
			return float, nil
		}
		return 0.0, fmt.Errorf("function %s is not callable", functionName)
	}
	return 0.0, fmt.Errorf("function %s not found", functionName)
}

func executePythonCodeWithNoArgumentAndNoReturnValue(moduleName, functionName string) error {
	module := python3.PyImport_ImportModule(moduleName)
	if module == nil {
		return fmt.Errorf("could not load module %s", moduleName)
	}

	function := module.GetAttrString(functionName)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			function.Call(nilArgs, nil)
			return nil
		}
		return fmt.Errorf("function %s is not callable", functionName)
	}
	return fmt.Errorf("function %s not found", functionName)
}
