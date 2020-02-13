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

	model, version, err := pli.SoftwareVersion()
	log.Println("PL Model name", model)
	log.Println("PL Software version", version)

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

	h, m, s, err := pli.Time()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Time: %v:%v:%v", h, m, s)

	soc, err := pli.StateOfCharge()
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Add percent to number displayed
	log.Printf("State of charge: %v", soc)
}
