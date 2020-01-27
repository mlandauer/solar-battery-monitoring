package pli

import (
	"errors"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

// PLI is used to talk to a particular PLI
type PLI struct {
	Port    io.ReadWriteCloser
	Prog    int // System program
	Voltage int // Voltage of battery system
}

func New(portName string) (pli PLI, err error) {
	// Set up options.
	// 8 bit, No parity, 1 stop bit is what the PLI expects
	// 9600 baud is the fastest speed the PLI can work at. That baud rate needs to be setup
	// with DIP switches on the PLI circuitboard itself. This is like a little glimpse into the past.
	options := serial.OpenOptions{
		PortName:   portName,
		BaudRate:   9600,
		DataBits:   8,
		StopBits:   1,
		ParityMode: serial.PARITY_NONE,
		// Will wait at most 100ms for a new byte to arrive
		InterCharacterTimeout: 100,
	}

	// Open the port.
	port, err := serial.Open(options)
	pli.Port = port
	if err != nil {
		return
	}

	log.Println("Doing a loopback test to make sure that communication channels are all working...")
	err = pli.loopbackTest()
	if err != nil {
		return
	}

	// Now get the system voltage (because we need that later to scale some readings)
	log.Println("Getting the system voltage...")
	prog, voltage, err := pli.volt()
	pli.Prog = prog
	pli.Voltage = voltage

	return
}

func (pli *PLI) Close() error {
	return pli.Port.Close()
}

// Potentially useful addresses to read from in RAM
//
// 0 - Software version number.The following applies (subject to change without notice):Version      0-127 = PL20Version  128-191 = PL40Version  192-210 = PL60Version  215-255 = PL80
// batv - 50 - Battery voltage in 0.1V steps scaled relative to 12V.eg. 128=12.8V, for 24V system 128*2=25.6V, for 48V system128*4=51.2V
// solv - 53  - solar voltage msb
// volt - 93 - msn= Prog number (0-4), lsn=system voltage (0-4)System voltage... 0=12V, 1=24V, 2=32V, 3=36V, 4=48VEg. 00110001 = 24V system running Prog 3.
// bcap - 94 - battery capacity in 20/100 Ah chunks
// bminl - 124 - lower byte of battery min voltage scaled to 12V
// bmaxl - 125 - lower byte of battery max voltage scaled to 12V
// dtemp - 180 - current external temperature
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
// extf - 207 - external flag and scale fileBit 3, Enable of LEXT.Bit 2, Enable for CEXTBit 1, 1=1A/step for LEXT (times 10), 0=0.1A/step for LEXTBit 0, 1=1A/step for CEXT (times 10), 0=0.1A/step for CEXT
// vext - 208 - external voltage reading 0-255 volt 1V steps
// cint - 213 - Internal (solar) charge current:0.1A steps for PL20 (eg. 10=1.0 Amp solar charge)0.2A steps for PL40 (eg. 10=2.0 Amps solar charge)0.4A steps for PL60 (eg. 10=4.0 Amps solar charge)
// lint - 217 - Internal LOAD- current:0.1A steps for PL20/PL40 (eg. 10=1.0A), 0.2A steps for PL60 (eg.10=2.0A)
// batvl - 220 - battery voltage lsb
// vbat - 221 - battery voltage msb
// solvl - 232 - solar voltage lsb
// vsol - 233 - solar voltage msb

// Reading the solar voltage is complicated because the charging needs to be stopped and the
// display activated to get an accurate reading

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

func extractNibbles(value byte) (msn byte, lsn byte) {
	msn = (value & 0xf0) >> 4 // most significant nibble
	lsn = value & 0xf         // least significant nibble
	return
}

func (pli *PLI) BatteryVoltage() (float32, error) {
	b, err := pli.ReadRAM(50)
	value := float32(b) * 0.1 * float32(pli.Voltage) / 12
	return value, err
}

func (pli *PLI) ReadRAM(address byte) (byte, error) {
	err := commandReadRAM(pli.Port, address)
	if err != nil {
		return 0, err
	}
	return readResponse(pli.Port)
}

var ErrLoopbackResponse = errors.New("PLI Error: Loopback response code")

// All one byte responses we consider errors (even loopback response)
func readResponse(port io.Reader) (byte, error) {
	buf := make([]byte, 2)
	n, err := port.Read(buf)
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if buf[0] == 200 {
			// We expect another byte
			n, err = port.Read(buf)
			if err != nil {
				return 0, err
			}
			if n != 1 {
				return 0, errors.New("Expected another byte")
			}
			return buf[0], nil
		} else {
			switch buf[0] {
			case 5:
				return 0, errors.New("PLI Error: No comms or corrupt comms")
			case 128:
				return 0, ErrLoopbackResponse
			case 129:
				return 0, errors.New("PLI Error: Timeout Error")
			case 130:
				return 0, errors.New("PLI Error: Checksum error in PLI receive data")
			case 131:
				return 0, errors.New("PLI Error: Command received by PLI is not recognised")
			case 133:
				return 0, errors.New("PLI Error: Processor did not receive a reply to request")
			case 134:
				return 0, errors.New("PLI Error: Error in reply from PL")
			default:
				return 0, errors.New("PLI Error: Unknown error code")
			}
		}
	} else if n == 2 {
		if buf[0] != 200 {
			return 0, errors.New("Received one byte more than expected")
		}
		return buf[1], nil
	} else {
		return 0, errors.New("Unexpected number of bytes")
	}
}

func (pli *PLI) loopbackTest() error {
	err := commandLoopbackTest(pli.Port)
	if err != nil {
		return err
	}
	_, err = readResponse(pli.Port)
	if err == nil {
		return errors.New("Expected one byte response")
	}
	if err != ErrLoopbackResponse {
		return err
	}
	return nil
}

func commandReadRAM(port io.Writer, address byte) error {
	return command(port, 20, address, 0)
}
func commandLoopbackTest(port io.Writer) error {
	return command(port, 187, 0, 0)
}

func command(port io.Writer, command byte, address byte, value byte) error {
	_, err := port.Write([]byte{command, address, value, 255 - command})
	return err
}
