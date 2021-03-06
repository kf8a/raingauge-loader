package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

func fileToTail(fileName string) *tail.Tail {
	tail, err := tail.TailFile(fileName, tail.Config{
		Follow: true,
		ReOpen: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	return tail
}

func prepareData(fields []string, variates stringSlice) map[string]string {
	result := make(map[string]string)
	for k, v := range fields {
		result[v] = variates[k]
	}
	return result
}

func sendMessage(data map[string]string, fileName string) {
	data["filename"] = fileName
	message, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(message))
	sendMessageToRabbitMQ(message)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func sendMessageToRabbitMQ(message []byte) {
	amqpUrl := os.Getenv("AMQP_URL")
	if amqpUrl == "" {
		return
	}
	conn, err := amqp.Dial(amqpUrl)
	failOnError(err, "failed to open connection")

	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")

	// err = c.ExchangeDeclare("datalogger", "fanout", true, false, false, false, nil)
	// failOnError("failed to declare exchange")

	q, err := ch.QueueDeclare(
		"data", // name
		true,   // durable
		false,  // delete when usused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare queue datalogger")

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "text/json",
			Body:         message,
		})
	failOnError(err, "Failed to publish a message")
}

func loadData(logger logger) {

	tail := fileToTail(logger.FileName)

	var batteryFieldNumber int
	var variates stringSlice

	// Read title and variables and units
	skip := 0

	for line := range tail.Lines {
		if strings.Contains(line.Text, "TOA5") {
			skip = 4
		}
		if skip == 3 {
			// decode headers
			fields := readCSVLine(line.Text)
			variates = stringSlice(fields)
			batteryFieldNumber = variates.pos(logger.BatteryVariateName)
		}
		if skip > 0 {
			// skip the rest
			skip = skip - 1
			continue
		}
		// Read data
		fields := readCSVLine(line.Text)

		// Parse time field
		const timeFormat = "2006-01-02 15:04:05"
		datetime, err := time.Parse(timeFormat, fields[0])
		if err != nil {
			log.Fatal(err)
		}

		// Skip if older than 1 day
		ignore_before := time.Now().Add(-24 * time.Hour)
		if datetime.Before(ignore_before) {
			continue
		}

		if batteryFieldNumber > 0 {
			voltage, err := strconv.ParseFloat(fields[batteryFieldNumber], 64)
			if err != nil {
				log.Fatal(err)
			}
			// emit results
			log.Println(fmt.Sprintf("%v: %v", logger.Site, voltage))
			batteryVoltage.WithLabelValues(logger.Site).Set(voltage)
		}
		loadEvents.WithLabelValues(logger.Site).Inc()
		data := prepareData(variates, fields)
		sendMessage(data, logger.FileName)
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
