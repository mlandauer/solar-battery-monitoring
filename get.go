package main

import (
	"errors"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

func main() {
	// Set up options.
	// 8 bit, No parity, 1 stop bit is what the PLI expects
	// 9600 baud is the fastest speed the PLI can work at. That baud rate needs to be setup
	// with DIP switches on the PLI circuitboard itself. This is like a little glimpse into the past.
	options := serial.OpenOptions{
		// TODO: Don't know how we easily get the device
		PortName:   "/dev/tty.usbserial-A8008HlV",
		BaudRate:   9600,
		DataBits:   8,
		StopBits:   1,
		ParityMode: serial.PARITY_NONE,
		// Will wait at most 100ms for a new byte to arrive
		InterCharacterTimeout: 100,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatal(err)
	}

	// Make sure to close it later.
	defer port.Close()

	err = loopbackTest(port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loopback test finished")

	// Now let's get the PL software version
	err = command(port, 20, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 2)
	n, err := port.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	if n == 1 {
		if buf[0] == 200 {
			// We expect another byte
			n, err = port.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			if n != 1 {
				log.Fatal("Expected another byte")
			}
			log.Println("PL Software version", buf[0])
		} else {
			log.Fatalf("Error code %v while getting PL software version", buf[0])
		}
	} else if n == 2 {
		if buf[0] != 200 {
			log.Fatal("Received one byte more than expected")
		}
		log.Println("PL Software version", buf[1])
	} else {
		log.Fatal("Unexpected number of bytes")
	}
}

func loopbackTest(port io.ReadWriter) error {
	err := commandLoopbackTest(port)
	if err != nil {
		return err
	}
	// Expect to receive one by with a value of 128
	buf := make([]byte, 2)
	n, err := port.Read(buf)
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("Only expected one byte")
	}
	if buf[0] != 128 {
		return errors.New("Expected return value of 128")
	}
	return nil
}

func commandLoopbackTest(port io.Writer) error {
	return command(port, 187, 0, 0)
}

func command(port io.Writer, command byte, address byte, value byte) error {
	_, err := port.Write([]byte{command, address, value, 255 - command})
	return err
}
