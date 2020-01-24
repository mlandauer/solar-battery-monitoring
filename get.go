package main

import (
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

	// Send loopback test command
	_, err = port.Write([]byte{187, 0, 0, 68})
	if err != nil {
		log.Fatal(err)
	}
	// Expect to receive one by with a value of 128
	buf := make([]byte, 2)
	n, err := port.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	if n != 1 {
		log.Fatal("Only expected one byte")
	}
	if buf[0] != 128 {
		log.Fatal("Expected return value of 128")
	}
	log.Println("Loopback test finished")
}
