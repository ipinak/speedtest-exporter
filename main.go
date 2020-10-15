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

var (
	showList   = kingpin.Flag("list", "Show available speedtest.net servers").Short('l').Bool()
	serverIds  = kingpin.Flag("server", "Select server id to speedtest").Short('s').Ints()
	timeoutOpt = kingpin.Flag("timeout", "Define timeout seconds. Default: 10 sec").Short('t').Int()
	timeout    = 10
)

var format string = `# speed_test_dl_speed
speed_test_dl_speed %f
# speed_test_ul_speed
speed_test_ul_speed %f
`

func main() {
	kingpin.Version("1.0.3")
	kingpin.Parse()

	setTimeout()

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
		targets.StartTest()
		for {
			result := targets.GetResult()
			// send an update
			update <- *result
			time.Sleep(10 * time.Second)
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
		fmt.Fprintf(w, format, result.AvgDL, result.AvgUL)
	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}
