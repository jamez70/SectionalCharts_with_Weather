package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Pirep struct {
	Report string
	Lat    string
	Lng    string
}

type Airport struct {
	ICAO string
	Lat  string
	Lng  string
	Alt  string
}

type Metar struct {
	ICAO  string
	METAR string
	COND  string
	LONG  string
	LAT   string
}

type Taf struct {
	ICAO string
	TAF  string
}

type WindUL struct {
	ICAO  string
	Winds string
}

type Winds struct {
	Direction int
	Speed     int
	Gust      int
}

type WeatherReports struct {
	Location string
	Time     string
	Metar    string
	Pirep    string
	TAF      string
	Winds    string
}


var Rpts map[string]WeatherReports
var airports []Airport
var metars []Metar
var pireps []Pirep
var tafs []Taf
var winds []WindUL
var useWx bool

var LatMin float64 = 20.0001576517236
var LatMax float64 = 55.4189882586259
var LngMin float64 = -179.0008962332189
var LngMax float64 = -53.7449437099231

var LatMinMap float64 = 20.0001576517236
var LatMaxMap float64 = 55.4189882586259
var LngMinMap float64 = -179.0008962332189
var LngMaxMap float64 = -53.7449437099231

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func FindMetar(x string) int {
	for i := range metars {
		if x == metars[i].ICAO {
			return i
		}
	}
	return -1
}

func FindWinds(x string) int {
	for i := range winds {
		if x == winds[i].ICAO {
			return i
		}
	}
	return -1
}

func FindTaf(x string) int {
	for i := range tafs {
		if x == tafs[i].ICAO {
			return i
		}
	}
	return -1
}

func stripK(apt string) string {
	if apt[:1] == "K" {
		return apt[1:]
	}
	return apt
}

func makeAirportName(apt string) string {
	if apt[:1] != "C" {
		if apt[:1] != "K" {
			return "K" + apt
		}
	}
	return apt
}

func readAirports(fname string) {
	f, err := os.Open(fname)
	check(err)
	fmt.Println("Opened " + fname)
	reader := csv.NewReader(bufio.NewReader(f))

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		AptLat, err := strconv.ParseFloat(strings.TrimSpace(line[1]), 64)
		check(err)
		AptLng, _ := strconv.ParseFloat(strings.TrimSpace(line[2]), 64)
		if AptLat >= LatMin && AptLat <= LatMax {
			if AptLng >= LngMin && AptLng <= LngMax {
				airports = append(airports, Airport{
					ICAO: makeAirportName(line[0]),
					Lat:  strings.TrimSpace(line[1]),
					Lng:  strings.TrimSpace(line[2]),
					Alt:  line[3],
				})
			}
		}
	}
	defer f.Close()
//		fmt.Println("$data = json_decode('[")
//		for i := range airports {
//			fmt.Println(" {\"loc\":["+airports[i].Lat+","+airports[i].Lng+"], \"title\": \""+airports[i].ICAO+"\"}," )
//		}
//		fmt.Println("]',true);")
}

//				  new L.LatLng(40.0003047916915, -93.0008962332189),
//				  new L.LatLng(44.2728613107929, -84.644232216245));

func scanPireps(fname string) {
	f, err := os.Open(fname)
	check(err)
	fmt.Println("Opened " + fname)
	reader := csv.NewReader(bufio.NewReader(f))
	reader.FieldsPerRecord = 45
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			if strings.Contains(error.Error(), "wrong number of fields in line") {
				// ignore copyright notice
				fmt.Printf("%#v\n", err)
				continue
			}
		}

		if len(line) > 1 {
			//		fmt.Println("Adding "+line[0]+" - "+line[1])
			PirepLat, _ := strconv.ParseFloat(strings.TrimSpace(line[9]), 64)
			//			check(err)
			PirepLng, _ := strconv.ParseFloat(strings.TrimSpace(line[10]), 64)
			if PirepLat >= LatMin && PirepLat <= LatMax {
				if PirepLng >= LngMin && PirepLng <= LngMax {
					//					for j := 0; j < len(line); j++ {
					//						fmt.Printf("[%d] - %s\n", j, line[j])
					//					}
					pireps = append(pireps, Pirep{
						Report: line[43],
						Lat:    line[9],
						Lng:    line[10],
					})
				}
			}
		}
	}
	defer f.Close()
}

func scanMetars(fname string) {
	f, err := os.Open(fname)
	check(err)
	fmt.Println("Opened " + fname)
	reader := csv.NewReader(bufio.NewReader(f))
	reader.FieldsPerRecord = 44
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			if strings.Contains(error.Error(), "wrong number of fields in line") {
				// ignore copyright notice
				fmt.Printf("%#v\n", err)
				continue
			}
		}

		if len(line) > 1 {
			//		fmt.Println("Adding "+line[0]+" - "+line[1])
			metars = append(metars, Metar{
				ICAO:  line[1],
				METAR: line[0],
				COND:  line[30],
				LONG:  line[4],
				LAT:   line[3],
			})
		}
	}
	defer f.Close()
}

func scanUatReportFile(fname string) {
	f2, err2 := os.Open(fname)
	if err2 != nil {
		panic(err2)
	}
	readBuf := make([]byte, 1048576)
	count, err := f2.Read(readBuf)
	if err != nil {
		panic(err)
	}
	errRead := json.Unmarshal(readBuf[0:count], &Rpts)
	if errRead != nil {
		panic(errRead)
	}
	f2.Sync()
	f2.Close()
	fmt.Printf("Read in %d reports\n", len(Rpts))
	// Read METARS into data
	for _, rpt := range Rpts {
		if len(rpt.Metar) > 0 {
			fmt.Printf("%s has METAR %s\n", rpt.Location, rpt.Metar)
			metars = append(metars, Metar{
				ICAO:  rpt.Location,
				METAR: rpt.Location + " " + rpt.Metar,
				COND:  "VFR",
			})
		}
		if len(rpt.TAF) > 0 {
			fmt.Printf("%s has TAF %s\n", rpt.Location, rpt.TAF)
			tafs = append(tafs, Taf{
				ICAO: rpt.Location,
				TAF:  rpt.Location + " " + rpt.TAF,
			})
		}
		if len(rpt.Winds) > 0 {
			fmt.Printf("%s has Winds %s\n", rpt.Location, rpt.Winds)
			winds = append(winds, WindUL{
				ICAO:  rpt.Location,
				Winds: rpt.Location + " " + rpt.Winds,
			})
		}
	}
}

func scanTafs(fname string) {
	f, err := os.Open(fname)
	check(err)
	fmt.Println("Opened " + fname)
	reader := csv.NewReader(bufio.NewReader(f))
	reader.FieldsPerRecord = 44
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			if strings.Contains(error.Error(), "wrong number of fields in line") {
				// ignore copyright notice
				fmt.Printf("%#v\n", err)
				continue
			}
		}

		if len(line) > 1 {
			//		for j:=0; j < len(line); j++ {
			//			fmt.Printf("[%d] - %s\n",j,line[j])
			//		}
			tafs = append(tafs, Taf{
				ICAO: line[1],
				TAF:  line[0],
			})
		}
	}
	defer f.Close()
}
func isOnMap(lat float64, lng float64) {
}

func getCondition(qual string) string {
	if qual == "VFR" {
		return "#60FF60"
	}
	if qual == "MVFR" {
		return "#4040FF"
	}
	if qual == "IFR" {
		return "#FF3030"
	}
	if qual == "LIFR" {
		return "#FF60FF"
	}
	return "white"
}

func getWinds(metar string) (int, int, int) {
	var val int
	var err error
	var wgust int

//	fmt.Printf("Get winds from %s\n",metar)
	words := strings.Fields(metar)
	for i := 0; i < len(words); i++ {
		if strings.HasSuffix(words[i], "KT") && len(words[i])>2 {
			windval := words[i]
			// Split it up here
			val, err = strconv.Atoi(windval[0:3])
			if err != nil {
				return 0, 0, 0
			}
			wdir := val
			val, err = strconv.Atoi(windval[3:5])
			if err != nil {
				return 0, 0, 0
			}
			wspd := val
			if windval[5] == 'G' {
				val, err = strconv.Atoi(windval[6:8])
				if err != nil {
					return 0, 0, 0
				}
				wgust = val
			} else {
				wgust = wspd
			}
			return wdir, wspd, wgust
		}
	}
	return 0, 0, 0
}

func getPrecip(metar string) string {
	words := strings.Fields(metar)
	for i := 1; i < len(words); i++ {
		if strings.Contains(words[i], "SN") {
			return words[i]
		}
		if strings.Contains(words[i], "TS") {
			return words[i]
		}
		if strings.Contains(words[i], "RA") {
			return words[i]
		}
		if strings.Contains(words[i], "BR") {
			return words[i]
		}
		if strings.Contains(words[i], "FG") {
			return words[i]
		}
		// Ignore everything after 'RMK'
		if strings.Contains(words[i], "RMK") {
			return ""
		}
	}
	return ""
}

func getTemperature(metar string) string {
	words := strings.Fields(metar)

	for i := 1; i < len(words); i++ {
		if strings.Contains(words[i], "/") && (strings.HasSuffix(words[i], "SM") == false) {
			w := strings.Replace(words[i], "M", "-", -1)
			slashpos := strings.IndexAny(w, "/")
			tpart := w[:slashpos]
			//			fmt.Printf("getTemperature: %s part is %s\n",w,tpart)
			val, err := strconv.Atoi(tpart)
			if err != nil {
			}
			if val > 1000 {
				val = val - 1000
				val = -val
			}
			val = val * 9
			val = val / 5
			val += 32
			return strconv.Itoa(val)
		}
	}
	return ""
}

type weatherData struct {
	Lng			string
	Lat			string
	ICAO		string
	WindDir		string
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

var WeatherData[] weatherData

//type Pirep struct {
//	Report string
//	Lat    string
//	Lng    string
//}

func getLightning(metar string) int {
	words := strings.Fields(metar)
	for i := 1; i < len(words); i++ {
		if strings.Contains(words[i], "LTG") {
			return 1
		}
	}
	return 0
}

func generatePireps(fname string) {
	w, err := os.Create(fname)
	check(err)
	message := "[ "
	for i:= range pireps {
		message = message + fmt.Sprintf(" { \"Report\": \"%s\", \"Lng\": \"%s\", \"Lat\": \"%s\" },\n",pireps[i].Report,pireps[i].Lng,pireps[i].Lat)
	}
	message = message + fmt.Sprintf("{} ]\n")
	w.Write([]byte(message))
	w.Close()
}

func generateFile(fname string) {

	w, err := os.Create(fname)
	check(err)
	message := "[ "
	
	for i := range metars {
		//		ApICAO := airports[i].ICAO
		//		ApLat := airports[i].Lat
		//		ApLng := airports[i].Lng
		MetarIndex := i // FindMetar(airports[i].ICAO)
		TafIndex := FindTaf(metars[i].ICAO)
		WindIndex := FindWinds(metars[i].ICAO)
		// Do we have a metar?
		if MetarIndex != -1 {
			message = message + fmt.Sprintf("{ \"Lng\": \"%s\", \"Lat\": \"%s\", \"ICAO\": \"%s\", ",metars[i].LONG, metars[i].LAT, metars[i].ICAO)
			MetarString := metars[MetarIndex].METAR
			wdir, wspeed, wgust := getWinds(MetarString)
			wbarb := wspeed
			wbarb = wbarb / 5
			if wspeed == 0 {
				wbarb = -1
			}
			
			message = message + fmt.Sprintf("\"WindDir\": \"%d\", \"WindSpeed\": \"%d\", \"WindBarb\": \"%d\", \"WindGust\": \"%d\", \"Metar\": \"%s\"", wdir, wspeed, wbarb, wgust, MetarString)
			cond := getCondition(metars[MetarIndex].COND)
			precip := getPrecip(MetarString)
			tt := getTemperature(MetarString)
			message = message + fmt.Sprintf(", \"Cond\": \"%s\", \"CondColor\": \"%s\", \"Precip\": \"%s\", \"Temperature\": \"%s\"", metars[MetarIndex].COND, cond, precip, tt)
			if TafIndex != -1 {
				TafString := tafs[TafIndex].TAF
				var taftmp string
				if strings.Contains(TafString, "<br>") {
					taftmp = strings.Replace(TafString, " FM", "<b> FM</b>", -1)
				} else {
					taftmp = strings.Replace(TafString, " FM", "<br><b>FM</b>", -1)
				}
				
				message = message + fmt.Sprintf(", \"TAF\": \"<br><small>%s</small>\"", taftmp)
	
			}
			if WindIndex != -1 {
				WindString := winds[WindIndex].Winds
				message = message + fmt.Sprintf(", \"UpWinds\": \"%s\"", WindString)

			}
			message = message + fmt.Sprintf(", \"Lightning\": \"%d\"",getLightning(MetarString))
			message = message + fmt.Sprintf(" },\n")
		}
	}
//	for i := range pireps {
//		message = message + fmt.Sprintf("L.marker([%s,%s], {icon: airplaneIcon, title: \"%s\",renderer: myRenderer}).bindPopup('%s').addTo(map);\n", pireps[i].Lat, pireps[i].Lng, pireps[i].Report, pireps[i].Report)
//	}
	// close init function
	message = message + fmt.Sprintf("{} ]\n")
	w.Write([]byte(message))
	w.Close()
}

func tryRead(fname string) {
	fd, err := os.Open(fname)
	if err != nil {
		fmt.Println("Cannot read file")
		return
	}
	defer fd.Close()
	buf := make([]byte, 1048576*8)
	count, err := fd.Read(buf)
	if err != nil {
		fmt.Println("Trouble reading file")
		return
	}
	var newWeatherData[] weatherData
	err = json.Unmarshal(buf[0:count], &newWeatherData)
	if err != nil {
		fmt.Println("Trouble parsing file")
		return
	}
	WeatherData = newWeatherData
	fmt.Println("Read in weather data!")
	fmt.Printf("Size is %d\n",len(WeatherData))
}

var useFlag string

func main() {
	useWx = false
	args := os.Args[1:]
	if len(args) != 0 && args[0] == "-w" {
		useWx = true
	}
	if useWx == true {
		useFlag = "On"
	} else {
		useFlag = "Off"
	}
	fmt.Println("Launching the program. useWx flag is " + useFlag)
	readAirports("airports.txt")
	if useWx == true {
		fmt.Println("Downloading metars.csv")
		err := DownloadFile("metars.csv", "https://aviationweather.gov/adds/dataserver_current/current/metars.cache.csv")
		check(err)
		fmt.Println("Downloading tafs.csv")
		err2 := DownloadFile("tafs.csv", "https://aviationweather.gov/adds/dataserver_current/current/tafs.cache.csv")
		check(err2)
		fmt.Println("Downloading reports.csv")
		err3 := DownloadFile("reports.csv", "https://aviationweather.gov/adds/dataserver_current/current/aircraftreports.cache.csv")
		check(err3)
		fmt.Println("Downloading pireps.csv")
		err4 := DownloadFile("pireps.csv", "https://aviationweather.gov/adds/dataserver_current/current/pireps.cache.csv")
		check(err4)
		scanMetars("metars.csv")
		scanTafs("tafs.csv")
		scanPireps("pireps.csv")
		generateFile("./weather.txt")
		generatePireps("./pireps.txt")
	} else {
//		scanUatReportFile("dump.txt")
	}
	tryRead("./weather.txt")
}

/*
	realtime = L.realtime(function(success, error) {
		fetch('https://wanderdrone.appspot.com/')
		.then(function(response) { return response.json(); })
		.then(function(data) {
			var trailCoords = trail.geometry.coordinates;
			trailCoords.push(data.geometry.coordinates);
			trailCoords.splice(0, Math.max(0, trailCoords.length - 5));
			success({
				type: 'FeatureCollection',
				features: [data, trail]
			});
		})
		.catch(error);
	}, {
		interval: 250
	}).addTo(map);
*/
