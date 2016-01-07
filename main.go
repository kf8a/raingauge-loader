package main

import (
	"encoding/csv"
	"encoding/json"
	"github.com/ActiveState/tail"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
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
		Name: "datalogger_battery_voltage",
		Help: "The current battery voltage",
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

type logger struct {
	FileName           string `json:"file-name"`
	Site               string `json:"site"`
	BatteryVariateName string `json:"battery-variate-name"`
}

func readCSVLine(text string) []string {
	reader := csv.NewReader(strings.NewReader(text))
	fields, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}
	return fields
}

func loadData(logger logger) {

	t, err := tail.TailFile(logger.FileName, tail.Config{
		Follow: true,
		ReOpen: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	var batteryFieldNumber int

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
			batteryFieldNumber = variates.pos(logger.BatteryVariateName)
		}
		if skip > 0 {
			// skip the rest
			skip = skip - 1
			continue
		}
		// Read data
		fields := readCSVLine(line.Text)
		voltage, err := strconv.ParseFloat(fields[batteryFieldNumber], 64)
		if err != nil {
			log.Fatal(err)
		}
		// log.Println(voltage)
		batteryVoltage.WithLabelValues(logger.Site).Set(voltage)
		loadEvents.WithLabelValues(logger.Site).Inc()
	}
}

func main() {

	dec := json.NewDecoder(os.Stdin)

	var loggers []logger
	if err := dec.Decode(&loggers); err != nil {
		log.Println(err)
		return
	}
	for _, lg := range loggers {
		go loadData(lg)
	}

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(":9094", nil)

}
