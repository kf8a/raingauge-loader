package main

import (
	"encoding/csv"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strconv"
	"strings"
	// "regexp"
	// "time"
)

var (
	loadEvents = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "datalogger_rows_loaded",
		Help: "the number of rows loaded into the database",
	})
	batteryVoltage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "datalogger_main_battery_voltage",
		Help: "The current main battery voltage",
	})
)

func init() {
	prometheus.MustRegister(loadEvents)
	prometheus.MustRegister(batteryVoltage)
}

func loadData(fileName string) {

	t, err := tail.TailFile(fileName, tail.Config{
		Follow: true,
		ReOpen: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Read title and variables and units
	skip := 0
	for line := range t.Lines {
		if strings.Contains(line.Text, "TOA5") {
			skip = 4
		}
		if skip != 0 {
			fmt.Println(line.Text)
			skip = skip - 1
			continue
		}
		// Read data
		reader := csv.NewReader(strings.NewReader(line.Text))
		fields, err := reader.Read()
		if err != nil {
			log.Fatal(err)
		}
		voltage, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(voltage)
		batteryVoltage.Set(voltage)
	}
}

func main() {

	go loadData("raingauge_Table1.dat")

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":9094", nil)

}
