package main

import (
	"encoding/csv"
	"github.com/ActiveState/tail"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	loadEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datalogger",
			Name:      "rows_loaded",
			Help:      "the number of rows loaded into the database",
		},
		[]string{"site"},
	)
	batteryVoltage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "datalogger_main_battery_voltage",
		Help: "The current main battery voltage",
	},
		[]string{"site"},
	)
)

func init() {
	prometheus.MustRegister(loadEvents)
	prometheus.MustRegister(batteryVoltage)
}

type stringSlice []string

func (slice stringSlice) pos(value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func readCSVLine(text string) []string {
	reader := csv.NewReader(strings.NewReader(text))
	fields, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}
	return fields
}

func loadData(fileName string) {

	t, err := tail.TailFile(fileName, tail.Config{
		Follow: true,
		ReOpen: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	battery := 1

	// Read title and variables and units
	skip := 0
	for line := range t.Lines {
		if strings.Contains(line.Text, "TOA5") {
			skip = 4
		}
		if skip == 3 {
			// decode headers
			fields := readCSVLine(line.Text)
			variates := stringSlice(fields)
			battery = variates.pos("GageMinV")
		}
		if skip > 0 {
			// skip the rest
			skip = skip - 1
			continue
		}
		// Read data
		fields := readCSVLine(line.Text)
		voltage, err := strconv.ParseFloat(fields[battery], 64)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(voltage)
		batteryVoltage.WithLabelValues("luxarbor").Set(voltage)
		loadEvents.WithLabelValues("luxarbor").Inc()
	}
}

func main() {

	go loadData("raingauge_Table1.dat")

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":9094", nil)

}
