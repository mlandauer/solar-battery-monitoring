package pli

import (
	"errors"
	"time"
)

// Methods for reading values (at a high level) from the PLI

// Potentially useful addresses to read from in RAM
// Done:
// 0 - Software version number.The following applies (subject to change without notice):Version      0-127 = PL20Version  128-191 = PL40Version  192-210 = PL60Version  215-255 = PL80
// sec - 46 - 2 seconds file, inc at 2 sec intervals
// min - 47 - Minutes file  (Value range = 0-5). Used for 6 minute timer
// hour - 48 - Current time (in 0.1 hrs steps ie. 6 minute intervals) Values = 0-23.9 Eg. 0=midnight, 100=10.0 hrs (10am), 145=14.5 hrs (2:30pm)
// batv - 50 - Battery voltage in 0.1V steps scaled relative to 12V.eg. 128=12.8V, for 24V system 128*2=25.6V, for 48V system128*4=51.2V
// volt - 93 - msn= Prog number (0-4), lsn=system voltage (0-4)System voltage... 0=12V, 1=24V, 2=32V, 3=36V, 4=48VEg. 00110001 = 24V system running Prog 3.
// bcap - 94 - battery capacity in 20/100 Ah chunks
// dsoc- 181 - SOC (day data state of charge)
// ciahl - 188 - Internal charge ah low byte
// ciahh - 189 - Internal charge ah high byte
// ceahl - 193 - External charge ah low byte
// ceahh - 194 - External charge ah high byte
// liahl - 198 - Internal load ah low byte
// liahh - 199 - Internal load ah high byte
// leahl - 203 - External load ah low byte
// leahh - 204 - External load ah high byte
// cext - 205 - external charge input (NOTE: First read ‘extf’ to check validity andscaling)
// lext - 206 - external load input (NOTE: First read ‘extf’ to check validity andscaling)
// extf - 207 - external flag and scale file - Bit 3, Enable of LEXT. - Bit 2, Enable for CEXT - Bit 1, 1=1A/step for LEXT (times 10), 0=0.1A/step for LEXT - Bit 0, 1=1A/step for CEXT (times 10), 0=0.1A/step for CEXT
// cint - 213 - Internal (solar) charge current:0.1A steps for PL20 (eg. 10=1.0 Amp solar charge)0.2A steps for PL40 (eg. 10=2.0 Amps solar charge)0.4A steps for PL60 (eg. 10=4.0 Amps solar charge)
// lint - 217 - Internal LOAD- current:0.1A steps for PL20/PL40 (eg. 10=1.0A), 0.2A steps for PL60 (eg.10=2.0A)
//
// TODO:
// solv - 53  - solar voltage msb
// bminl - 124 - lower byte of battery min voltage scaled to 12V
// bmaxl - 125 - lower byte of battery max voltage scaled to 12V
// vext - 208 - external voltage reading 0-255 volt 1V steps
// batvl - 220 - battery voltage lsb
// vbat - 221 - battery voltage msb
// solvl - 232 - solar voltage lsb
// vsol - 233 - solar voltage msb

// Reading the solar voltage is complicated because the charging needs to be stopped and the
// display activated to get an accurate reading

const PL20 = "PL20"
const PL40 = "PL40"
const PL60 = "PL60"
const PL80 = "PL80"

func (pli *PLI) softwareVersion() (string, byte, error) {
	value, err := pli.ReadRAM(0)
	if err != nil {
		return "", value, err
	}
	if value < 128 {
		return PL20, value, nil
	} else if value < 192 {
		return PL40, value, nil
	} else if value < 211 {
		return PL60, value, nil
	} else {
		return PL80, value, nil
	}
}

// Time returns the time (to the nearest 2 seconds) as stored in the PLI. This is used internally to
// total things over the day. So, it's fairly important that it's roughly correct.
func (pli *PLI) Time() (hour int, min int, sec int, err error) {
	a, err := pli.ReadRAM(46) // 2 second chunks
	if err != nil {
		return
	}
	if a > 29 {
		err = errors.New("Expected 'seconds' byte to be in the range 0-29")
		return
	}
	b, err := pli.ReadRAM(47) // minute chunks (0-5)
	if err != nil {
		return
	}
	if b > 5 {
		err = errors.New("Expected 'minute' byte to be in the range 0-5")
		return
	}
	c, err := pli.ReadRAM(48) // 6 minute chunks
	if err != nil {
		return
	}
	if c > 239 {
		err = errors.New("Expected 'hour' byte to be in the range 0-239")
	}
	sec = int(c)*6*60 + int(b)*60 + int(a)*2
	min = sec / 60
	sec = sec % 60
	hour = min / 60
	min = min % 60
	return
}

// CheckTime gets the time as stored in the PLI but will also error if it's too
// different from the "real" time as known by the computer
func (pli *PLI) CheckTime() (hour int, min int, sec int, err error) {
	hour, min, sec, err = pli.Time()
	if err != nil {
		return
	}
	// Now compare the time to the real time and error if it is 15 minutes or more out
	realTime := time.Now()
	systemTime := time.Date(
		realTime.Year(), realTime.Month(), realTime.Day(),
		hour, min, sec, 0,
		realTime.Location(),
	)
	if systemTime.Sub(realTime).Minutes() >= 15 {
		err = errors.New("PL system time is too different from the real time")
	}
	return
}

func (pli *PLI) byteToVoltage(b byte) float32 {
	return float32(b) * float32(pli.Voltage) / 10 / 12
}

// TODO: This gives a slightly different reading to what the PL80 is showing (out by 0.1V)
func (pli *PLI) BatteryVoltage() (float32, error) {
	b, err := pli.ReadRAM(50)
	return pli.byteToVoltage(b), err
}

// BatterCapacity returns the capacity of the battery measured in Ah
func (pli *PLI) BatteryCapacity() (int, error) {
	b, err := pli.ReadRAM(94)
	if err != nil {
		return 0, err
	}
	// Battery capacity setting 20A/100A per step - 20A steps until 1000Ah, 100Ah steps >1000Ah
	var value int
	value = int(b) * 20
	if value > 1000 {
		value = 1000 + (value-1000)/20*100
	}
	return value, nil
}

// Gets the overall PL program number and the system voltage
func (pli *PLI) volt() (prog int, voltage int, err error) {
	v, err := pli.ReadRAM(93)
	progByte, voltByte := extractNibbles(v)
	if progByte > 4 {
		err = errors.New("Expected program number to be in the range 0-4")
		return
	}
	prog = int(progByte)
	switch voltByte {
	case 0:
		voltage = 12
	case 1:
		voltage = 24
	case 2:
		voltage = 32
	case 3:
		voltage = 36
	case 4:
		voltage = 48
	default:
		err = errors.New("Unexpected voltage value returned by PLI")
	}
	return
}

const RegulatorStateBoost = "boost"
const RegulatorStateEqualise = "equalise"
const RegulatorStateAbsorption = "absorption"
const RegulatorStateFloat = "float"

func (pli *PLI) RegulatorState() (string, error) {
	b, err := pli.ReadRAM(101)
	if err != nil {
		return "", err
	}
	switch b & 0x3 {
	case 0:
		return RegulatorStateBoost, nil
	case 1:
		return RegulatorStateEqualise, nil
	case 2:
		return RegulatorStateAbsorption, nil
	case 3:
		return RegulatorStateFloat, nil
	default:
		return "", errors.New("Unexpected value")
	}
}

// StateOfCharge returns a number between 0 and 100 which is very very roughly a measure of
// how full the battery is. There are many ways this number can be misleading. So be careful.
func (pli *PLI) StateOfCharge() (int, error) {
	b, err := pli.ReadRAM(181)
	return int(b), err
}

func twoBytes(h byte, l byte) int {
	return (int(h) << 8) | int(l)
}

func (pli *PLI) readRAMTwoBytes(la byte, ha byte) (int, error) {
	l, err := pli.ReadRAM(la)
	if err != nil {
		return 0, err
	}
	h, err := pli.ReadRAM(ha)
	if err != nil {
		return 0, err
	}
	return twoBytes(h, l), nil
}

// InternalIn returns value as Ah
func (pli *PLI) InternalIn() (int, error) {
	return pli.readRAMTwoBytes(188, 189)
}

// ExternalIn returns value as Ah
func (pli *PLI) ExternalIn() (int, error) {
	return pli.readRAMTwoBytes(193, 194)
}

// In returns value as Ah
func (pli *PLI) In() (int, error) {
	internal, err := pli.InternalIn()
	if err != nil {
		return 0, err
	}
	external, err := pli.ExternalIn()
	if err != nil {
		return 0, err
	}
	return internal + external, nil
}

// InternalOut returns value as Ah
func (pli *PLI) InternalOut() (int, error) {
	return pli.readRAMTwoBytes(198, 199)
}

// ExternalOut returns value as Ah
func (pli *PLI) ExternalOut() (int, error) {
	return pli.readRAMTwoBytes(203, 204)
}

// Out returns value as Ah
func (pli *PLI) Out() (int, error) {
	internal, err := pli.InternalOut()
	if err != nil {
		return 0, err
	}
	external, err := pli.ExternalOut()
	if err != nil {
		return 0, err
	}
	return internal + external, nil
}

// ExternalCharge returns value in A
func (pli *PLI) ExternalCharge() (float32, error) {
	extf, err := pli.ReadRAM(207)
	if err != nil {
		return 0, err
	}
	// If bit 2 is not set it is disabled so just return zero
	if extf&0x4 == 0 {
		return 0, nil
	}
	current, err := pli.ReadRAM(205)
	if err != nil {
		return 0, err
	}
	value := float32(current)
	if extf&0x1 == 0 {
		value = value / 10
	}
	return value, nil
}

// ExternalLoad returns value in A
func (pli *PLI) ExternalLoad() (float32, error) {
	extf, err := pli.ReadRAM(207)
	if err != nil {
		return 0, err
	}
	// If bit 3 is not set it is disabled so just return zero
	if extf&0x8 == 0 {
		return 0, nil
	}
	current, err := pli.ReadRAM(206)
	if err != nil {
		return 0, err
	}
	value := float32(current)
	if extf&0x2 == 0 {
		value = value / 10
	}
	return value, nil
}

func (pli *PLI) InternalCharge() (float32, error) {
	v, err := pli.ReadRAM(213)
	if err != nil {
		return 0, err
	}
	value := float32(v)
	switch pli.Model {
	case PL20:
		value = value / 10
	case PL40:
		value = value / 5
	// Guessing what it is for PL80 - undocumented
	case PL60, PL80:
		value = value * 2 / 5
	}
	return value, nil
}

func (pli *PLI) InternalLoad() (float32, error) {
	v, err := pli.ReadRAM(217)
	if err != nil {
		return 0, err
	}
	value := float32(v)
	switch pli.Model {
	case PL20, PL40:
		value = value / 10
	// Guessing what it is for PL80 - undocumented
	case PL60, PL80:
		value = value / 5
	}
	return value, nil
}

func (pli *PLI) Charge() (float32, error) {
	internal, err := pli.InternalCharge()
	if err != nil {
		return 0, err
	}
	external, err := pli.ExternalCharge()
	if err != nil {
		return 0, err
	}
	return internal + external, nil
}

func (pli *PLI) Load() (float32, error) {
	internal, err := pli.InternalLoad()
	if err != nil {
		return 0, err
	}
	external, err := pli.ExternalLoad()
	if err != nil {
		return 0, err
	}
	return internal + external, nil
}
