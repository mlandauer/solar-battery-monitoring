package main

import (
	"log"

	"github.com/mlandauer/solar-battery-monitoring/pkg/pli"
)

func main() {
	// TODO: Don't yet know how we easily get the port name for the device
	log.Println("Setting up communication with the PLI...")
	pli, err := pli.New("/dev/tty.usbserial-AM009SBW", 9600)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer pli.Close()

	log.Printf("System program number: %v", pli.Prog)
	log.Printf("System voltage: %v V", pli.Voltage)
	log.Printf("PL Model name: %v", pli.Model)
	log.Printf("PL Software version: %v", pli.SoftwareVersion)

	v, err := pli.BatteryVoltage()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Battery voltage: %v V", v)

	bc, err := pli.BatteryCapacity()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Battery capacity: %v Ah", bc)

	h, m, s, err := pli.Time()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Time: %v:%v:%v", h, m, s)

	soc, err := pli.StateOfCharge()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("State of charge: %v%%", soc)

	in, err := pli.In()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("In: %v Ah", in)

	out, err := pli.Out()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Out: %v Ah", out)

	charge, err := pli.Charge()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Charge: %v A", charge)

	load, err := pli.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Load: %v A", load)
}
