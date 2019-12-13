package main

import (
	"fmt"
	"github.com/kingdo777/hey/requester"
	"math"
	"net/http"
	gourl "net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

const (
	headerRegexp = `^([\w-]+):\s*(.+)`
	authRegexp   = `^(.+):([^\s].+)`
	heyUA        = "hey/0.0.1"
)

var (
	m                  = "GET"
	headers            = ""
	body               = ""
	bodyFile           = ""
	accept             = ""
	contentType        = "text/html"
	authHeader         = ""
	hostHeader         = ""
	output             = ""
	c                  = 50
	n                  = 200
	q                  = 0.0
	t                  = 20
	z                  = 0
	h2                 = false
	cpus               = runtime.GOMAXPROCS(-1)
	disableCompression = false
	disableKeepAlives  = false
	disableRedirects   = false
	proxyAddr          = ""
)

func hey(url string, conc int, d string, ch chan float64) {
	runtime.GOMAXPROCS(cpus)
	num := n
	//conc := c
	//dur := z

	//flag.Duration("z", 0, "")
	//dur := new(time.Duration)
	dur, err := time.ParseDuration(d)
	if err != nil {
		crashAndExit(err.Error())
	}

	if dur > 0 {
		num = math.MaxInt32
		if conc <= 0 {
			crashAndExit("-c cannot be smaller than 1.")
		}
	} else {
		if num <= 0 || conc <= 0 {
			crashAndExit("-n and -c cannot be smaller than 1.")
		}

		if num < conc {
			crashAndExit("-n cannot be less than -c.")
		}
	}

	method := strings.ToUpper(m)

	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", contentType)
	// set basic auth if set
	var username, password string
	var bodyAll []byte
	var proxyURL *gourl.URL

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		crashAndExit(err.Error())
	}
	req.ContentLength = int64(len(bodyAll))
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
	ua := req.UserAgent()
	if ua == "" {
		ua = heyUA
	} else {
		ua += " " + heyUA
	}
	header.Set("User-Agent", ua)
	req.Header = header

	w := &requester.Work{
		Request:            req,
		RequestBody:        bodyAll,
		N:                  num,
		C:                  conc,
		QPS:                q,
		Timeout:            t,
		DisableCompression: disableCompression,
		DisableKeepAlives:  disableKeepAlives,
		DisableRedirects:   disableRedirects,
		H2:                 h2,
		ProxyAddr:          proxyURL,
		Output:             output,
	}
	w.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		w.Stop()
	}()
	if dur > 0 {
		go func() {
			time.Sleep(dur)
			w.Stop()
		}()
	}
	w.Run()
	ch <- w.Report.GetLatency()
}

func crashAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
