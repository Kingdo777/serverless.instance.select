/*
   Copyright 2017 The Knative Authors
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"github.com/Kingdo777/serverless.instance.select/pkg/hey"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	SloDefault time.Duration = 50 * time.Millisecond
)

func parseParam(r *http.Request, param string) (int, error) {
	return strconv.Atoi(r.URL.Query().Get(param))
}

func handler(w http.ResponseWriter, r *http.Request) {
	conc, _ := parseParam(r, "conc")
	runTime, _ := parseParam(r, "duration")
	url := getUrl()
	latency := hey.SendRequest(url, conc, runTime)
	//fmt.Println(latency)
	_, _ = fmt.Fprintf(w, "%f", latency)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world kingdo.\n")
}

func main() {
	validateToken := os.Getenv("VALIDATION")
	if validateToken != "" {
		http.HandleFunc("/"+validateToken+"/", replyWithToken(validateToken))
	}

	//这里由dockerfile给出
	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		listenPort = "8081"
	}
	http.HandleFunc("/start", handler)
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":"+listenPort, nil)
}

func getUrl() string {
	return "http://localhost:8080"
}

func replyWithToken(token string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, token)
	}
}
