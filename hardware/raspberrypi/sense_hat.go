// +build hardware

package raspberrypi

import (
	"context"
	"fmt"
	"github.com/DataDog/go-python3"
	"strconv"
	"strings"
	"time"
)

const (
	SenseHatModuleName = "sense_hat"

	// LED Matrix
	SetRotation = "set_rotation"
	FlipH       = "flip_h"
	FlipV       = "flip_v"
	SetPixels   = "set_pixels"
	GetPixels   = "get_pixels"
	SetPixel    = "set_pixel"
	GetPixel    = "get_pixel"
	LoadImage   = "load_image"
	Clear       = "clear"
	ShowMessage = "show_message"
	ShowLetter  = "show_letter"
	LowLight    = "low_light"

	// Environmental sensors
	GetTemperature             = "get_temperature"
	GetTemperatureFromHumidity = "get_temperature_from_humidity"
	GetTemperatureFromPressure = "get_temperature_from_pressure"
	GetHumidity                = "get_humidity"
	GetPressure                = "get_pressure"

	// IMU Sensor
	SetIMUConfig          = "set_imu_config"
	GetOrientationRadians = "get_orientation_radians"
	GetOrientationDegrees = "get_orientation_degrees"
	GetOrientation        = "get_orientation"
	GetCompass            = "get_compass"
	GetCompassRaw         = "get_compass_raw"
	GetGyroscope          = "get_gyroscope"
	GetGyroscopeRaw       = "get_gyroscope_raw"
	GetAccelerometer      = "get_accelerometer"
	GetAccelerometerRaw   = "get_accelerometer_raw"

	// Joystick
	WaitForEvent = "wait_for_event"
	GetEvents    = "get_events"
)

// SenseHat provides a direct interface to the Sense HAT API
// see https://pythonhosted.org/sense-hat/api/
type SenseHat struct {
	senseHatObject *python3.PyObject
}

func NewSenseHat() (*SenseHat, error) {
	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		fmt.Println("Error initializing the python interpreter")
		return nil, fmt.Errorf("error initializing the interpreter")
	}

	// Loading sense_hat senseHatModule
	senseHatModule := python3.PyImport_ImportModule(SenseHatModuleName)
	if senseHatModule == nil {
		return nil, fmt.Errorf("could not load sense_hat module")
	}

	// Getting the constructor
	senseHatConstructor := senseHatModule.GetAttrString("SenseHat")
	if senseHatConstructor == nil {
		return nil, fmt.Errorf("could not create a SenseHat object")
	}

	if !python3.PyCallable_Check(senseHatConstructor) {
		return nil, fmt.Errorf("sense hat constructor not callable")
	}

	nilArgs := python3.PyTuple_New(0)
	senseHat := senseHatConstructor.Call(nilArgs, nil)

	return &SenseHat{
		senseHatObject: senseHat,
	}, nil
}

// **********
// LED Matrix
// **********

// Rotation represents the possible image rotations
// Rotation0 corresponds to the RPi HDMI Port facing downwards
type Rotation int

const (
	Rotation0   Rotation = 0
	Rotation90           = 90
	Rotation180          = 180
	Rotation270          = 270
)

// PixelColor represents the RGB color of a Pixel
type PixelColor struct {
	R uint8
	G uint8
	B uint8
}

// SetRotation calls set_rotation which corrects the orientation of the image being shown.
func (s *SenseHat) SetRotation(r Rotation, redraw bool) error {
	function := s.senseHatObject.GetAttrString(SetRotation)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(2)

			pyR := python3.PyLong_FromGoInt(int(r))
			python3.PyTuple_SetItem(args, 0, pyR)

			pyRedraw := python3.Py_True
			if !redraw {
				pyRedraw = python3.Py_False
			}
			python3.PyTuple_SetItem(args, 1, pyRedraw)

			function.Call(args, nil)
		}
		return fmt.Errorf("function %s is not callable", SetRotation)
	}
	return fmt.Errorf("could not get %s function", SetRotation)
}

// FlipH calls flip_h which flips the image on the LED matrix horizontally.
// Redraw controls whether or not to redraw what is already displayed on the LED Matrix.
func (s *SenseHat) FlipH(redraw bool) error {
	return s.flipImage(FlipH, redraw)
}

// FlipV calls flip_h which flips the image on the LED matrix horizontally.
// Redraw controls whether or not to redraw what is already displayed on the LED Matrix.
func (s *SenseHat) FlipV(redraw bool) error {
	return s.flipImage(FlipV, redraw)
}

// SetPixels calls set_pixels which updates the entire LED matrix based on a 64 length list of pixel values.
func (s *SenseHat) SetPixels(pixelsList []PixelColor) error {
	if len(pixelsList) != 64 {
		return fmt.Errorf("the pixels' list must contain 64 pixels, found %d", len(pixelsList))
	}

	function := s.senseHatObject.GetAttrString(SetPixels)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(1)
			pyPixelsList := fromPixelsListToPyPixelsList(pixelsList)
			python3.PyTuple_SetItem(args, 0, pyPixelsList)
			function.Call(args, nil)
			return nil
		}
		return fmt.Errorf("function %s is not callable", SetPixels)
	}
	return fmt.Errorf("could not get %s function", SetPixels)
}

// GetPixels calls get_pixels which returns the entire LED matrix list of pixel values.
func (s *SenseHat) GetPixels() ([]PixelColor, error) {
	function := s.senseHatObject.GetAttrString(GetPixels)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			pyPixels := function.Call(nilArgs, nil)
			return fromPyPixelsListToPixelList(pyPixels)
		}
		return nil, fmt.Errorf("function %s is not callable", GetPixels)
	}
	return nil, fmt.Errorf("could not get %s function", GetPixels)
}

// SetPixel calls set_pixel which sets an individual LED matrix pixel at the specified
// X-Y coordinate to the specified colour.
func (s *SenseHat) SetPixel(x, y uint8, p PixelColor) error {
	function := s.senseHatObject.GetAttrString(SetPixel)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(3)
			pyX := python3.PyLong_FromLong(int(x))
			pyY := python3.PyLong_FromLong(int(y))
			pyPixel := fromPixelToPyPixel(p)

			python3.PyTuple_SetItem(args, 0, pyX)
			python3.PyTuple_SetItem(args, 1, pyY)
			python3.PyTuple_SetItem(args, 2, pyPixel)

			function.Call(args, nil)
			return nil
		}
		return fmt.Errorf("function %s is not callable", SetPixel)
	}
	return fmt.Errorf("could not get %s function", SetPixel)
}

// GetPixel calls get_pixel which returns the PixelColor of the pixel at coordinates X-Y
func (s *SenseHat) GetPixel(x, y uint8) (PixelColor, error) {
	function := s.senseHatObject.GetAttrString(GetPixel)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(2)
			pyX := python3.PyLong_FromLong(int(x))
			pyY := python3.PyLong_FromLong(int(y))

			python3.PyTuple_SetItem(args, 0, pyX)
			python3.PyTuple_SetItem(args, 1, pyY)

			pyPixel := function.Call(args, nil)
			return fromPyPixelToPixel(pyPixel)
		}
		return PixelColor{}, fmt.Errorf("function %s is not callable", GetPixel)
	}
	return PixelColor{}, fmt.Errorf("could not get %s function", GetPixel)
}

// LoadImage calls load_image which loads an image file, converts it to RGB format
// and displays it on the LED matrix. The image must be 8 x 8 pixels in size.
func (s *SenseHat) LoadImage(path string, redraw bool) error {
	function := s.senseHatObject.GetAttrString(LoadImage)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(2)
			pyPath := python3.PyUnicode_FromString(path)

			pyRedraw := python3.Py_True
			if !redraw {
				pyRedraw = python3.Py_False
			}

			python3.PyTuple_SetItem(args, 0, pyPath)
			python3.PyTuple_SetItem(args, 1, pyRedraw)

			function.Call(args, nil)
			return nil
		}
		return  fmt.Errorf("function %s is not callable", LoadImage)
	}
	return fmt.Errorf("could not get %s function", LoadImage)
}

// Clear calls clear which sets the entire LED matrix to a single colour.
func (s *SenseHat) Clear(color PixelColor) error {
	function := s.senseHatObject.GetAttrString(Clear)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(1)

			pyColor := fromPixelToPyPixel(color)
			python3.PyTuple_SetItem(args, 0, pyColor)

			function.Call(args, nil)
			return nil
		}
		return  fmt.Errorf("function %s is not callable", Clear)
	}
	return fmt.Errorf("could not get %s function", Clear)
}

// ShowMessage calls show_message which scrolls a text message from right to left across the LED matrix
// and at the specified speed, in the specified colour and background colour.
// scrollSpeed is the pause between two shifts
func (s *SenseHat) ShowMessage(textString string, scrollSpeed float64, textColor PixelColor, backColor PixelColor) error {
	function := s.senseHatObject.GetAttrString(ShowMessage)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(4)

			pyText := python3.PyUnicode_FromString(textString)
			pyScrollSpeed := python3.PyFloat_FromDouble(scrollSpeed)

			pyTextColor := fromPixelToPyPixel(textColor)
			pyBackColor := fromPixelToPyPixel(backColor)

			python3.PyTuple_SetItem(args, 0, pyText)
			python3.PyTuple_SetItem(args, 1, pyScrollSpeed)
			python3.PyTuple_SetItem(args, 2, pyTextColor)
			python3.PyTuple_SetItem(args, 3, pyBackColor)

			function.Call(args, nil)
			return nil
		}
		return  fmt.Errorf("function %s is not callable", ShowMessage)
	}
	return fmt.Errorf("could not get %s function", ShowMessage)
}

// ShowLetter calls show_letter which displays a single text character on the LED matrix.
func (s *SenseHat) ShowLetter(letter string, textColor PixelColor, backColor PixelColor) error {
	function := s.senseHatObject.GetAttrString(ShowLetter)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(3)

			pyText := python3.PyUnicode_FromString(letter)
			pyTextColor := fromPixelToPyPixel(textColor)
			pyBackColor := fromPixelToPyPixel(backColor)

			python3.PyTuple_SetItem(args, 0, pyText)
			python3.PyTuple_SetItem(args, 1, pyTextColor)
			python3.PyTuple_SetItem(args, 2, pyBackColor)

			function.Call(args, nil)
			return nil
		}
		return  fmt.Errorf("function %s is not callable", ShowLetter)
	}
	return fmt.Errorf("could not get %s function", ShowLetter)
}

// LowLight toggles the LED Matrix low light mode, which is useful when the Sense HAT is used
// in dark environment.
func (s *SenseHat) LowLight(enabled bool) {
	pyEnabled := python3.Py_True
	if !enabled {
		pyEnabled = python3.Py_False
	}
	s.senseHatObject.SetAttrString(LowLight, pyEnabled)
}

// *********************
// Environmental sensors
// *********************

// GetTemperature calls get_temperature (which calls get_temperature_from_humidity)
// The temperature is in degrees Celsius.
func (s *SenseHat) GetTemperature() (float64, error) {
	return s.readFloatNoArg(GetTemperature)
}

// GetTemperatureFromHumidity calls get_temperature_from_humidity, which reads temperature
// from the humidity sensor. The temperature is in degrees Celsius.
func (s *SenseHat) GetTemperatureFromHumidity() (float64, error) {
	return s.readFloatNoArg(GetTemperatureFromHumidity)
}

// GetTemperatureFromPressure calls get_temperature_from_pressure, which reads temperature
// from the pressure sensor. The temperature is in degrees Celsius.
func (s *SenseHat) GetTemperatureFromPressure() (float64, error) {
	return s.readFloatNoArg(GetTemperatureFromPressure)
}

// GetHumidity calls get_humidity, which reads the relative humidity from the humidity sensor.
// The humidity is a percentage.
func (s *SenseHat) GetHumidity() (float64, error) {
	return s.readFloatNoArg(GetHumidity)
}

// GetPressure calls get_pressure, which reads the current pressure in Millibars from the pressure sensor.
func (s *SenseHat) GetPressure() (float64, error) {
	return s.readFloatNoArg(GetPressure)
}

// *******************************
// Inertial Measurement Unit (IMU)
// *******************************

type Orientation struct {
	Pitch float64
	Roll  float64
	Yaw   float64
}

type Point3D struct {
	X float64
	Y float64
	Z float64
}

// SetImuConfig calls set_imu_config which enables and disables the gyroscope, accelerometer
// and/or magnetometer contribution to the get orientation functions below.
// BUG(guillaumebour) Calling this function seems to fail with "function not callable", but
// getting the orientation data still works fine without calling it beforehand.
func (s *SenseHat) SetImuConfig(compassEnabled, gyroEnabled, accelEnabled bool) error {
	function := s.senseHatObject.GetAttrString(SetIMUConfig)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(3)
			cEnabled := python3.Py_True
			if !compassEnabled {
				cEnabled = python3.Py_False
			}
			python3.PyTuple_SetItem(args, 0, cEnabled)

			gEnabled := python3.Py_True
			if !gyroEnabled {
				gEnabled = python3.Py_False
			}
			python3.PyTuple_SetItem(args, 1, gEnabled)

			aEnabled := python3.Py_True
			if !accelEnabled {
				aEnabled = python3.Py_False
			}
			python3.PyTuple_SetItem(args, 2, aEnabled)

			function.Call(args, nil)
		}
		return fmt.Errorf("function %s is not callable", SetIMUConfig)
	}
	return fmt.Errorf("could not get %s function", SetIMUConfig)
}

// GetOrientationRadians calls get_orientation_radians which gets the current orientation
// in radians using the aircraft principal axes of pitch, roll and yaw.
func (s *SenseHat) GetOrientationRadians() (Orientation, error) {
	return s.readOrientation(GetOrientationRadians)
}

// GetOrientationDegrees calls get_orientation_degrees which gets the current orientation
// in degrees using the aircraft principal axes of pitch, roll and yaw.
func (s *SenseHat) GetOrientationDegrees() (Orientation, error) {
	return s.readOrientation(GetOrientationDegrees)
}

// GetOrientation calls get_orientation which calls get_orientation_degrees
func (s *SenseHat) GetOrientation() (Orientation, error) {
	return s.readOrientation(GetOrientation)
}

// GetCompass calls get_compass which calls set_imu_config to
// disable the gyroscope and accelerometer then gets the direction of North
// from the magnetometer in degrees.
func (s *SenseHat) GetCompass() (float64, error) {
	return s.readFloatNoArg(GetCompass)
}

// GetCompassRaw calls get_compass_raw which gets the raw x, y and z axis magnetometer data.
func (s *SenseHat) GetCompassRaw() (Point3D, error) {
	return s.readPoint3DData(GetCompassRaw)
}

// GetGyroscope calls get_gyroscope which calls set_imu_config to disable the magnetometer
// and accelerometer then gets the current orientation from the gyroscope only.
func (s *SenseHat) GetGyroscope() (Orientation, error) {
	return s.readOrientation(GetGyroscope)
}

// GetGyroscopeRaw calls get_gyroscope_raw which gets the raw x, y and z axis gyroscope data.
func (s *SenseHat) GetGyroscopeRaw() (Point3D, error) {
	return s.readPoint3DData(GetGyroscopeRaw)
}

// GetAccelerometer calls get_accelerometer which Calls set_imu_config to disable the magnetometer
// and gyroscope then gets the current orientation from the accelerometer only.
func (s *SenseHat) GetAccelerometer() (Orientation, error) {
	return s.readOrientation(GetAccelerometer)
}

// GetAccelerometerRaw calls get_accelerometer_raw which gets the raw x, y and z axis accelerometer data.
func (s *SenseHat) GetAccelerometerRaw() (Point3D, error) {
	return s.readPoint3DData(GetAccelerometerRaw)
}

// ********
// Joystick
// ********

type InputEvent struct {
	Timestamp time.Time
	Direction string
	Action    string
}

// WaitForEvent calls wait_for_event which blocks execution until a joystick event occurs,
// then returns an InputEvent representing the event that occurred.
func (s *SenseHat) WaitForEvent(emptyBuffer bool) (InputEvent, error) {
	stick := s.senseHatObject.GetAttrString("stick")
	if stick == nil {
		return InputEvent{}, fmt.Errorf("could not get stick")
	}

	function := stick.GetAttrString(WaitForEvent)
	if function != nil {
		if python3.PyCallable_Check(function) {
			var pyInputEvent *python3.PyObject
			if emptyBuffer {
				args := python3.PyTuple_New(1)
				python3.PyTuple_SetItem(args, 0, python3.Py_True)
				pyInputEvent = function.Call(args, nil)
			} else {
				nilArgs := python3.PyTuple_New(0)
				pyInputEvent = function.Call(nilArgs, nil)
			}
			return fromPyInputEventToInputEvent(pyInputEvent)
		}
		return InputEvent{}, fmt.Errorf("function %s is not callable", WaitForEvent)
	}
	return InputEvent{}, fmt.Errorf("could not get %s function", WaitForEvent)
}

// GetEvents calls get_events which returns a list of InputEvent
// representing all events that have occurred since the last call to get_events or wait_for_event.
func (s *SenseHat) GetEvents() ([]InputEvent, error) {
	stick := s.senseHatObject.GetAttrString("stick")
	if stick == nil {
		return nil, fmt.Errorf("could not get stick")
	}

	function := stick.GetAttrString(GetEvents)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			pyInputEvents := function.Call(nilArgs, nil)
			lenPyInputEvents := python3.PyList_Size(pyInputEvents)
			events := make([]InputEvent, 0)
			if lenPyInputEvents > 0 {
				for k := 0; k < lenPyInputEvents; k++ {
					pyInputEvent := python3.PyList_GetItem(pyInputEvents, k)
					inputEvent, err := fromPyInputEventToInputEvent(pyInputEvent)
					if err != nil {
						return nil, fmt.Errorf("could not get event: %v", err)
					}
					events = append(events, inputEvent)
				}
			}
			return events, nil
		}
		return nil, fmt.Errorf("function %s is not callable", WaitForEvent)
	}
	return nil, fmt.Errorf("could not get %s function", WaitForEvent)
}

type InputEventHandler func(inputEvent InputEvent)

// StartListeningForJoystickEvents listens for joystick input events and pass them to the provided InputEventHandler
// If blocking is true, the call to the handler will block, otherwise, it will be done asynchronously.
func (s *SenseHat) StartListeningForJoystickEvents(handler InputEventHandler, blocking bool) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				nextEvent, err := s.WaitForEvent(false)
				if err != nil {
					fmt.Printf("error while getting input event: %v", err)
				}
				if !blocking {
					go handler(nextEvent)
				} else {
					handler(nextEvent)
				}

			}
		}
	}()
	return cancel
}

// ****************
// Helper functions
// ****************

func (s *SenseHat) readFloatNoArg(functionName string) (float64, error) {
	function := s.senseHatObject.GetAttrString(functionName)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			retValue := function.Call(nilArgs, nil)
			temperature := python3.PyFloat_AsDouble(retValue)
			return temperature, nil
		}
		return 0.0, fmt.Errorf("function %s is not callable", functionName)
	}
	return 0.0, fmt.Errorf("could not get %s function", functionName)
}

func (s *SenseHat) readOrientation(functionName string) (Orientation, error) {
	function := s.senseHatObject.GetAttrString(functionName)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			orientationDict := function.Call(nilArgs, nil)

			pyPitch := python3.PyDict_GetItemString(orientationDict, "pitch")
			pyRoll := python3.PyDict_GetItemString(orientationDict, "roll")
			pyYaw := python3.PyDict_GetItemString(orientationDict, "yaw")

			pitch := python3.PyFloat_AsDouble(pyPitch)
			roll := python3.PyFloat_AsDouble(pyRoll)
			yaw := python3.PyFloat_AsDouble(pyYaw)

			return Orientation{
				Pitch: pitch,
				Roll:  roll,
				Yaw:   yaw,
			}, nil
		}
		return Orientation{}, fmt.Errorf("function %s is not callable", functionName)
	}
	return Orientation{}, fmt.Errorf("could not get %s function", functionName)
}

func (s *SenseHat) readPoint3DData(functionName string) (Point3D, error) {
	function := s.senseHatObject.GetAttrString(functionName)
	if function != nil {
		if python3.PyCallable_Check(function) {
			nilArgs := python3.PyTuple_New(0)
			dataDict := function.Call(nilArgs, nil)

			pyX := python3.PyDict_GetItemString(dataDict, "x")
			pyY := python3.PyDict_GetItemString(dataDict, "y")
			pyZ := python3.PyDict_GetItemString(dataDict, "z")

			pitch := python3.PyFloat_AsDouble(pyX)
			roll := python3.PyFloat_AsDouble(pyY)
			yaw := python3.PyFloat_AsDouble(pyZ)

			return Point3D{
				X: pitch,
				Y: roll,
				Z: yaw,
			}, nil
		}
		return Point3D{}, fmt.Errorf("function %s is not callable", functionName)
	}
	return Point3D{}, fmt.Errorf("could not get %s function", functionName)
}

func (s *SenseHat) flipImage(functionName string, redraw bool) error {
	function := s.senseHatObject.GetAttrString(functionName)
	if function != nil {
		if python3.PyCallable_Check(function) {
			args := python3.PyTuple_New(1)

			pyRedraw := python3.Py_True
			if !redraw {
				pyRedraw = python3.Py_False
			}
			python3.PyTuple_SetItem(args, 0, pyRedraw)

			function.Call(args, nil)
			return nil
		}
		return fmt.Errorf("function %s is not callable", functionName)
	}
	return fmt.Errorf("could not get %s function", functionName)
}

func pythonRepr(o *python3.PyObject) (string, error) {
	if o == nil {
		return "", fmt.Errorf("object is nil")
	}

	s := o.Repr()
	if s == nil {
		python3.PyErr_Clear()
		return "", fmt.Errorf("failed to call Repr object method")
	}
	defer s.DecRef()

	return python3.PyUnicode_AsUTF8(s), nil
}

func fromPyInputEventToInputEvent(pyInputEvent *python3.PyObject) (InputEvent, error) {
	timestampStr, err := pythonRepr(pyInputEvent.GetAttrString("timestamp"))
	if err != nil {
		return InputEvent{}, fmt.Errorf("could not read timestamp: %v", err)
	}
	timestampParts := strings.Split(timestampStr, ".")
	var timestampNSec int64 = 0

	timestampSec, err := strconv.ParseInt(timestampParts[0], 10, 64)
	if err != nil {
		return InputEvent{}, fmt.Errorf("could not parse timestamp: %v", err)
	}

	if len(timestampParts) > 1 {
		timestampNSec, err = strconv.ParseInt(timestampParts[1], 10, 64)
		if err != nil {
			return InputEvent{}, fmt.Errorf("could not parse timestamp: %v", err)
		}
	}

	direction, err := pythonRepr(pyInputEvent.GetAttrString("direction"))
	if err != nil {
		return InputEvent{}, fmt.Errorf("could not read direction: %v", err)
	}

	action, err := pythonRepr(pyInputEvent.GetAttrString("action"))
	if err != nil {
		return InputEvent{}, fmt.Errorf("could not read action: %v", err)
	}

	event := InputEvent{
		Timestamp: time.Unix(timestampSec, timestampNSec),
		Direction: strings.Replace(direction, "'", "", -1),
		Action:    strings.Replace(action, "'", "", -1),
	}
	return event, nil
}

func fromPixelsListToPyPixelsList(pixelsList []PixelColor) *python3.PyObject {
	pyPixelsList := python3.PyList_New(len(pixelsList))
	for k, pixel := range pixelsList {
		python3.PyList_SetItem(pyPixelsList, k, fromPixelToPyPixel(pixel))
	}
	return pyPixelsList
}

func fromPyPixelsListToPixelList(pyPixelsList *python3.PyObject) ([]PixelColor, error) {
	pixelsList := make([]PixelColor, 0)
	for k := 0; k < python3.PyList_Size(pyPixelsList); k++ {
		pixel, err := fromPyPixelToPixel(python3.PyList_GetItem(pyPixelsList, k))
		if err != nil {
			return nil, fmt.Errorf("could not get pixel at index %d: %v", k, err)
		}
		pixelsList = append(pixelsList, pixel)
	}
	return pixelsList, nil
}

func fromPixelToPyPixel(pixel PixelColor) *python3.PyObject {
	pyPixel := python3.PyList_New(3)

	python3.PyList_SetItem(pyPixel, 0, python3.PyLong_FromGoInt(int(pixel.R)))
	python3.PyList_SetItem(pyPixel, 1, python3.PyLong_FromGoInt(int(pixel.G)))
	python3.PyList_SetItem(pyPixel, 2, python3.PyLong_FromGoInt(int(pixel.B)))

	return pyPixel
}

func fromPyPixelToPixel(pyPixel *python3.PyObject) (PixelColor, error) {
	if python3.PyList_Size(pyPixel) != 3 {
		return PixelColor{}, fmt.Errorf("incorrect pixel size: expected 3, got %d", python3.PyList_Size(pyPixel))
	}

	R := uint8(python3.PyLong_AsLong(python3.PyList_GetItem(pyPixel, 0)))
	G := uint8(python3.PyLong_AsLong(python3.PyList_GetItem(pyPixel, 1)))
	B := uint8(python3.PyLong_AsLong(python3.PyList_GetItem(pyPixel, 2)))

	return PixelColor{R, G, B}, nil
}
