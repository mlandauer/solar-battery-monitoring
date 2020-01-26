package main

import (
	"log"

	"github.com/mlandauer/solar-battery-monitoring/pkg/pli"
)

func main() {
	// TODO: Don't yet know how we easily get the port name for the device
	pli, err := pli.New("/dev/tty.usbserial-A8008HlV", 24)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer pli.Close()

	err = pli.LoopbackTest()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loopback test finished")

	// Now let's get the PL software version
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
}
