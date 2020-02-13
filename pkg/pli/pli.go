package pli

import (
	"errors"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

// PLI is used to talk to a particular PLI
type PLI struct {
	Port            io.ReadWriteCloser
	Prog            int // System program
	Voltage         int // Voltage of battery system
	Model           string
	SoftwareVersion int
}

func New(portName string, baudRate uint) (pli PLI, err error) {
	// TODO: Check that the baudRate is one of the speeds supported by the PLI
	// Set up options.
	// 8 bit, No parity, 1 stop bit is what the PLI expects
	// 9600 baud is the fastest speed the PLI can work at. That baud rate needs to be setup
	// with DIP switches on the PLI circuitboard itself. This is like a little glimpse into the past.
	options := serial.OpenOptions{
		PortName:   portName,
		BaudRate:   baudRate,
		DataBits:   8,
		StopBits:   1,
		ParityMode: serial.PARITY_NONE,
		// Will wait at most 1s for a new byte to arrive
		InterCharacterTimeout: 1000,
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
	if err != nil {
		return
	}
	pli.Prog = prog
	pli.Voltage = voltage

	// We need the model type for later so we're getting it now
	model, softwareVersion, err := pli.softwareVersion()
	if err != nil {
		return
	}
	pli.Model = model
	pli.SoftwareVersion = int(softwareVersion)

	return
}

func (pli *PLI) Close() error {
	return pli.Port.Close()
}

func extractNibbles(value byte) (msn byte, lsn byte) {
	msn = (value & 0xf0) >> 4 // most significant nibble
	lsn = value & 0xf         // least significant nibble
	return
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
