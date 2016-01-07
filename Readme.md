# Datalogger loader

A utility to trail the Campbell datalogger TOA5 files and make the number of rows and the battery voltage available for scraping into 
http://prometheus.io/, metrics are available on port 9094.

##Usage

    go build
    raingauge-loader < example.json

