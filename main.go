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

var format string = `# speed_test_dl_speed
speed_test_dl_speed %5.2f
# speed_test_ul_speed
speed_test_ul_speed %5.2f
# speed_test_ping_ms
speed_test_ping_ms %d`

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()

	setTimeout()
	setPort()

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
		fmt.Fprintf(w, format, result.AvgDL, result.AvgUL, result.AvgPing)
	})

	fmt.Println(http.ListenAndServe(port, nil))
}
