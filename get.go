package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/influxdata/influxdb-client-go"
	"github.com/joho/godotenv"
	"github.com/mlandauer/solar-battery-monitoring/pkg/pli"
)

func main() {
	// Only try to read .env file if it exists
	_, err := os.Stat(".env")
	if err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}

	influx, err := influxdb.New(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("INFLUXDB_TOKEN"),
		influxdb.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer influx.Close()

	var device string
	switch runtime.GOOS {
	case "darwin":
		device = "/dev/tty.usbserial-AM009SBW"
	case "linux":
		device = "/dev/ttyUSB0"
	default:
		log.Fatal("Unsupported operation system")
	}

	// TODO: Don't yet know how we easily get the port name for the device
	log.Println("Setting up communication with the PLI...")
	pli, err := pli.New(device, 9600)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer pli.Close()

	log.Printf("System program number: %v", pli.Prog)
	log.Printf("System voltage: %v V", pli.Voltage)
	log.Printf("PL Model name: %v", pli.Model)
	log.Printf("PL Software version: %v", pli.SoftwareVersion)

	// h, m, s, err := pli.CheckTime()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("Time: %v:%v:%v", h, m, s)

	for {
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

		state, err := pli.RegulatorState()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Regulator State: %v", state)

		_, err = influx.Write(
			context.Background(), os.Getenv("INFLUXDB_BUCKET"), os.Getenv("INFLUXDB_ORG"),
			influxdb.NewRowMetric(
				map[string]interface{}{
					"battery_voltage": v,
					"soc":             soc,
					"in":              in,
					"out":             out,
					"charge":          charge,
					"load":            load,
					"regulator_state": state,
				},
				"solar",
				map[string]string{},
				time.Now(),
			),
		)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Sleeping for ten seconds...")
		time.Sleep(time.Second * 10)
	}
}
