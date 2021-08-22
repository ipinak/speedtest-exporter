package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func setTimeout() {
	if *timeoutOpt != 0 {
		timeout = *timeoutOpt
	}
}

func setPort() {
	if *portOpt != 0 {
		port = fmt.Sprintf(":%d", *portOpt)
	}
}

var (
	showList   = kingpin.Flag("list", "Show available speedtest.net servers").Short('l').Bool()
	serverIds  = kingpin.Flag("server", "Select server id to speedtest").Short('s').Ints()
	timeoutOpt = kingpin.Flag("timeout", "Define timeout seconds. Default: 10 sec").Short('t').Int()
	portOpt    = kingpin.Flag("port", "Select port to listen").Short('p').Int()
	timeout    = 10
	port       = ":8080"
)

var format string = `# HELP speed_test_dl_speed_total speedtest-exporter: Download speed in Mbit/s of internet connection.
# TYPE speed_test_dl_speed_total gauge
speed_test_dl_speed_total{hostname="%s"} %5.2f
# HELP speed_test_ul_speed_total speedtest-exporter: Upload speed in Mbit/s of internet connection.
# TYPE speed_test_ul_speed_total gauge
speed_test_ul_speed_total{hostname="%s"} %5.2f
# HELP speed_test_ping_seconds_total speedtest-exporter: Ping time of internet connection.
# TYPE speed_test_ping_seconds_total counter
speed_test_ping_seconds_total{hostname="%s"} %f`

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()

	setTimeout()
	setPort()
	hostname, _ := os.Hostname()

	user := fetchUserInfo()
	user.Show()

	list := fetchServerList(user)
	if *showList {
		list.Show()
		return
	}

	update := make(chan Result, 1)
	targets := list.FindServer(*serverIds)
	go func() {
		for {
			targets.StartTest()
			result := targets.GetResult()
			// send an update
			update <- *result

			fmt.Println("Waiting 20s...")
			time.Sleep(20 * time.Second)
		}
	}()

	// receive new data
	var result Result
	go func() {
		for {
			result = <-update
		}
	}()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, format,
			hostname, result.AvgDL,
			hostname, result.AvgUL,
			hostname, result.AvgPing.Seconds())
	})

	fmt.Println(http.ListenAndServe(port, nil))
}
