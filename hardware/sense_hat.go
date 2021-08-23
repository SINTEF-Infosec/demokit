package hardware

import (
	_ "embed"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	RaspberryModuleName     = "rpi"
	ReadTemperatureFunction = "read_temperature"
	ReadHumidityFunction    = "read_humidity"
)

type SenseHat struct {
}

func NewSenseHat() *SenseHat {
	return &SenseHat{}
}

func (s *SenseHat) ReadHumidity() (float64, error) {
	return executePythonCodeWithNoArgumentsAndFloatReturn(RaspberryModuleName, ReadHumidityFunction)
}

func (s *SenseHat) ReadTemperature() (float64, error) {
	return executePythonCodeWithNoArgumentsAndFloatReturn(RaspberryModuleName, ReadTemperatureFunction)
}

// Should be replace by real python code execution (ToDo: check timing)
func executePythonCodeWithNoArgumentsAndFloatReturn(moduleName, functionName string) (float64, error) {
	pythonCode := fmt.Sprintf("import %s; print(%s.%s())", moduleName, moduleName, functionName)
	cmd := exec.Command("/usr/bin/python3", "-c", pythonCode)
	result, err := cmd.CombinedOutput()
	if err != nil {
		return 0.0, fmt.Errorf("could not execute python code: %v", err)
	}
	floatRes, err := strconv.ParseFloat(strings.Trim(string(result), "\n"), 64)
	if err != nil {
		return 0.0, fmt.Errorf("could not parse float result: %v", err)
	}
	return floatRes, nil
}
