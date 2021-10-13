package main

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"strings"
	"strconv"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil" )

type pirepData struct {
	Report	string
	Lng 	string
	Lat		string
}	
	
type weatherData struct {
	Lng			string
	Lat			string
	ICAO		string
	WindDir		string
	WindBarb	string
	WindSpeed 	string
	WindGust	string
	Metar		string
	Cond		string
	CondColor	string
	Precip		string
	Temperature string
	TAF			string
	UpWinds		string
	Lightning	string
}

type apData struct {
	airport []weatherData 
}

var WeatherData[] weatherData
var PirepData[] pirepData
var ap apData 

func tryRead(fname string) {
	fd, err := os.Open(fname)
	if err != nil {
		return
	}
	defer fd.Close()
	buf := make([]byte, 1048576*8)
	count, err := fd.Read(buf)
	if err != nil {
		return
	}
	var newWeatherData[] weatherData
	err = json.Unmarshal(buf[0:count], &newWeatherData)
	if err != nil {
		return
	}
	WeatherData = newWeatherData
}

func tryReadPirep(fname string) {
	fd, err := os.Open(fname)
	if err != nil {
		return
	}
	defer fd.Close()
	buf := make([]byte, 1048576*8)
	count, err := fd.Read(buf)
	if err != nil {
		return
	}
	var newPirepData[] pirepData
	err = json.Unmarshal(buf[0:count], &newPirepData)
	if err != nil {
		return
	}
	PirepData = newPirepData
}

func parsePireps(Lng1 float64, Lat1 float64, Lng2 float64, Lat2 float64) {
	var prList[] pirepData
	tryReadPirep("/disk/dev/mapsrv/pireps.txt")
	i := 0
	for {
		pr := PirepData[i]
		if i >= len(PirepData)-1 {
			break;
		}
		Lng,err1 := strconv.ParseFloat(pr.Lng, 64)
		Lat,err2 := strconv.ParseFloat(pr.Lat, 64)
//		fmt.Printf("i is %d len is %d lat is %s lng is %s\n",i,len(PirepData),pr.Lat,pr.Lng)
//		fmt.Printf("[ %f < %f < %f L2 %f L3 %f L4 %f ] ",Lng1,Lng, Lng2,Lat1,Lat2)
		if err1 == nil && err2 == nil {
			if (Lng > Lng1 && Lng < Lng2) && (Lat > Lat1 && Lat < Lat2) {
//				fmt.Printf("i is %d len is %d lat is %s lng is %s\n",i,len(PirepData),pr.Lat,pr.Lng)
				prList = append(prList, pr)
			}
			
		}
		i = i + 1
	}
	b, err := json.Marshal(prList)
	if err == nil {
		os.Stdout.Write(b)
	}	
}

func ParseAirports(Lng1 float64, Lat1 float64, Lng2 float64, Lat2 float64) {
	var apList []weatherData
	tryRead("/disk/dev/mapsrv/weather.txt")
	i := 0

	for {
		wx := WeatherData[i]
		if i >= len(WeatherData)-1 {
			break;
		}
		Lng,err1 := strconv.ParseFloat(wx.Lng, 64)
		Lat,err2 := strconv.ParseFloat(wx.Lat, 64)
		if err1 == nil && err2 == nil {
			if (Lng > Lng1 && Lng < Lng2) && (Lat > Lat1 && Lat < Lat2) {
//				fmt.Printf("<pre>Adding %s to list</pre>\n",wx.ICAO)			
				apList = append(apList, wx)
			}
			
		}
		i = i + 1
	}
	
	b, err := json.Marshal(apList)
	if err == nil {
		os.Stdout.Write(b)
	}
}

type voidCloser struct {
	io.Reader
}
func (voidCloser) Close() error { return nil }

func ModGoRequest() (r http.Request) {
	post, _ := ioutil.ReadAll(os.Stdin)
	r.URL, _ = url.ParseRequestURI( "http://"+os.Getenv("HTTP_HOST")+"?"+os.Getenv("QUERY_STRING") )
	r.Method = os.Getenv("REQUEST_METHOD");
	r.Header = map[string][]string{
		"Accept-Encoding": {os.Getenv("HTTP_ACCEPT_ENCODING")},
		"Accept-Language": {os.Getenv("HTTP_ACCEPT_LANGUAGE")},
		"Connection": {os.Getenv("HTTP_CONNECTION")},
		"Content-Type": {os.Getenv("CONTENT_TYPE")},
		"Content-Length": {os.Getenv("CONTENT_LENGTH")} }
	r.Body = voidCloser{bytes.NewBuffer( post ) }
	return r
}

func main() {
	var req http.Request = ModGoRequest()

	fmt.Print("Content-Type: text/html; charset=utf-8\r\n")
	fmt.Print("\r\n")
	if req.FormValue("req") == "airports" {
	
		var bounds = req.FormValue("bounds")
		var coords = strings.Split(bounds,",")
		
		Lon1,err1 := strconv.ParseFloat(coords[0], 64)
		Lat1,err2 := strconv.ParseFloat(coords[1], 64)
		Lon2,err3 := strconv.ParseFloat(coords[2], 64)
		Lat2,err4 := strconv.ParseFloat(coords[3], 64)
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			ParseAirports(Lon1, Lat1, Lon2, Lat2)
		}
		
	}
	if req.FormValue("req") == "pireps" {
	
		var bounds = req.FormValue("bounds")
		var coords = strings.Split(bounds,",")
		
		Lon1,err1 := strconv.ParseFloat(coords[0], 64)
		Lat1,err2 := strconv.ParseFloat(coords[1], 64)
		Lon2,err3 := strconv.ParseFloat(coords[2], 64)
		Lat2,err4 := strconv.ParseFloat(coords[3], 64)
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			parsePireps(Lon1, Lat1, Lon2, Lat2)
		}
		
	}	
}
