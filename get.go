package main

import (
	"errors"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

func main() {
	// TODO: Don't yet know how we easily get the port name for the device
	pli, err := NewPLI("/dev/tty.usbserial-A8008HlV")
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer pli.Close()

	err = pli.loopbackTest()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loopback test finished")

	// Now let's get the PL software version
	value, err := pli.readRAM(0)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("PL Software version", value)
}

func NewPLI(portName string) (PLI, error) {
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
	return PLI{Port: port}, err
}

// PLI is used to talk to a particular PLI
type PLI struct {
	Port io.ReadWriteCloser
}

func (pli *PLI) Close() error {
	return pli.Port.Close()
}

func (pli *PLI) readRAM(address byte) (byte, error) {
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
