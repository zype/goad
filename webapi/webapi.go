package webapi

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gophergala2016/goad/sqsadaptor"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")
var upgrader = websocket.Upgrader{}

func serveResults(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/goad" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		http.Error(w, "Missing URL", 400)
		return
	}

	concurrencyStr := r.URL.Query().Get("c")
	concurrency, cerr := strconv.Atoi(concurrencyStr)
	if cerr != nil {
		http.Error(w, "Invalid concurrency", 400)
		return
	}

	totStr := r.URL.Query().Get("tot")
	tot, toterr := strconv.Atoi(totStr)
	if toterr != nil {
		http.Error(w, "Invalid total", 400)
		return
	}

	timeoutStr := r.URL.Query().Get("timeout")
	timeout, timeouterr := strconv.Atoi(timeoutStr)
	if timeouterr != nil {
		http.Error(w, "Invalid timeout", 400)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Websocket upgrade:", err)
		return
	}
	defer c.Close()

	resultChan := make(chan sqsadaptor.RegionsAggData)

	// go startTest(url, concurrency, tot, timeout)

	sqsadaptor.Aggregate(resultChan)

	for {
		result, more := <-resultChan
		if !more {
			break
		}
		fmt.Println(result) // stuff the results over the websocket
		err = c.WriteMessage(websocket.TextMessage, []byte("{\"hello\" : \"goodbye\"}"))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// Serve waits for connections and serves the results
func Serve() {
	http.HandleFunc("/goad", serveResults)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}