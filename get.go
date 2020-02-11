package main

import (
	"log"

	"github.com/mlandauer/solar-battery-monitoring/pkg/pli"
)

func main() {
	// TODO: Don't yet know how we easily get the port name for the device
	pli, err := pli.New("/dev/tty.usbserial-AM009SBW")
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer pli.Close()

	log.Println("System program number", pli.Prog)
	log.Println("System voltage, ", pli.Voltage)

	// Now let's get the PL software version
	// TODO: Get PL model number out of this and extract method
	value, err := pli.ReadRAM(0)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("PL Software version", value)

	v, err := pli.BatteryVoltage()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Battery voltage", v)

	bc, err := pli.BatteryCapacity()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Batter capacity: %v Ah", bc)
}
