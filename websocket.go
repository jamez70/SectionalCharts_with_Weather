// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"log"
	"strings"
	"net/url"
	"os"
	"os/signal"
	"time"
    "encoding/json"
	"github.com/gorilla/websocket"
)

type WeatherReports struct {
   Location          string 
   Time              string 
   Metar              string 
   Pirep			string
   TAF				string
   Winds			string
}

type WeatherMessage struct {
   Type              string `json:"Type"`
   Location          string `json:"Location"`
   Time              string `json:"Time"`
   Data              string `json:"Data"`
   LocaltimeReceived time.Time `json:"LocaltimeReceiver"`
}

var addr = flag.String("addr", "192.168.1.8", "http service address")

var Rpts map[string]WeatherReports

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func main() {
    
	Rpts = make(map[string]WeatherReports)
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if Exists("dump.txt") {
		f2, err2 := os.Open("dump.txt")
		if err2 != nil {
			panic(err2)
		}
		readBuf := make([]byte, 1048576)
		count, err := f2.Read(readBuf)
		if err != nil {
			panic(err)
		}
		errRead := json.Unmarshal(readBuf[0:count],&Rpts)
		if errRead != nil {
			panic(errRead)
		}
		f2.Sync()
		f2.Close()
	}
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/weather"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	log.Println("Connected!")
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
//			log.Printf("%s", message)
			var d = new(WeatherMessage)
			err2 := json.Unmarshal(message, &d)
			if err2 != nil {
				log.Println("Error unmarshal ",err2)
			}
			ourLocation := d.Location
			switch d.Type {
			case "PIREP", "WINDS":
				ourLocation = "K"+ourLocation
			}
			log.Printf("Type %s\nLocation %s\nData %s\n\n",d.Type,ourLocation,d.Data)
			// If not in the database then add it
			if _, ok := Rpts[ourLocation]; !ok {
				log.Println("Adding new report for "+ourLocation)
				var newrpt WeatherReports
				newrpt.Location = ourLocation
			}
			rpt := Rpts[ourLocation]
			rpt.Time = d.Time;
			rpt.Location = ourLocation
			fmtData := strings.Replace(d.Data, "\n","<br>",-1)
			switch d.Type {
			case "METAR","SPECI":
				rpt.Metar = fmtData
			case "TAF", "TAF.AMD":
				rpt.TAF = fmtData
			case "PIREP":
				rpt.Pirep = fmtData
			case "WINDS":
				rpt.Winds = fmtData
			default:
				log.Println("Unhandled type "+d.Type)
			}
			Rpts[ourLocation] = rpt
		}
	}()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			log.Println("Save File ")
			
			f2, err2 := os.Create("dump.txt")
			if err2 != nil {
				panic(err2)
			}
			buf, err := json.Marshal(&Rpts)
			if err != nil {
				panic(err)
			}
			f2.Write(buf)
			f2.Sync()
			f2.Close()
			
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			return
		}
	}
}