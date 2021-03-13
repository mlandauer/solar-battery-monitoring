package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/influxdata/influxdb-client-go"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/mlandauer/solar-battery-monitoring/pkg/pli"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func captureAndRecord() {
	// Only try to read .env file if it exists
	_, err := os.Stat(".env")
	if err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Connect to influxdb
	influx, err := influxdb.New(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("INFLUXDB_TOKEN"),
		influxdb.WithHTTPClient(http.DefaultClient),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer influx.Close()

	// Connect to postgres
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DATABASE"),
	)

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.Fatal(err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

	systemVoltage.Set(float64(pli.Voltage))

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

		t := time.Now()

		measurementTime.SetToCurrentTime()
		batteryVoltage.Set(float64(v))
		batteryStateOfCharge.Set(float64(soc))
		inGauge.Set(float64(in))
		outGauge.Set(float64(out))
		chargeGauge.Set(float64(charge))
		loadGauge.Set(float64(load))

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
				t,
			),
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(
			"INSERT INTO measurements (time, battery_voltage, soc, in_value, out_value, charge, load, regulator_state) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			t, v, soc, in, out, charge, load, state,
		)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Sleeping for ten seconds...")
		time.Sleep(time.Second * 10)
	}
}

var (
	measurementTime = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "time",
		Help:      "Time",
	})
	batteryVoltage = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "battery_voltage",
		Help:      "Battery voltage in Volts",
	})
	batteryStateOfCharge = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "battery_state_of_charge_percentage",
		Help:      "Percentage full of the battery",
	})
	inGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "in_amp_hours",
		Help:      "Energy in since midnight measured in Amp Hours",
	})
	outGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "out_amp_hours",
		Help:      "Energy used since midnight measured in Amp Hours",
	})
	chargeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "charge_amps",
		Help:      "Current generated in Amps",
	})
	loadGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "load_amps",
		Help:      "Current used in Amps",
	})
	systemVoltage = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "solar",
		Name:      "system_voltage",
		Help:      "Voltage that overall system operates at",
	})
)

func main() {
	go func() {
		captureAndRecord()
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
