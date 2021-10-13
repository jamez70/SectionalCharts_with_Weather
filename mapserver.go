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

var LatMin float64 = 36.0001576517236
var LatMax float64 = 48.4189882586259
var LngMin float64 = -93.0008962332189
var LngMax float64 = -83.7449437099231

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

func FindMeter(x string) int {
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
	if apt[:1] != "K" {
		return "K" + apt
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
		for i := range airports {
			fmt.Println("Airport "+airports[i].ICAO+" ["+airports[i].Lat+","+airports[i].Lng+"]")
		}
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

func scanMeters(fname string) {
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

	words := strings.Fields(metar)
	for i := 0; i < len(words); i++ {
		if strings.HasSuffix(words[i], "KT") {
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
		if strings.Contains(words[i], "RMK") {
			fmt.Printf("Contains RMK at %d - done\n", i)
			break
		}
	}
	return ""
}

func getLightning(metar string) int {
	words := strings.Fields(metar)
	for i := 1; i < len(words); i++ {
		if strings.Contains(words[i], "LTG") {
			return 1
		}
	}
	return 0
}

func generateFile(fname string) {

	w, err := os.Create(fname)
	check(err)

	message := `
	var map = null;
    var latlng_range = 0.01;	  
    var overlay;
	var inset_id = "";
	// Get url parameters
	var params = {};
	window.location.href.replace(/[?&]+([^=&]+)=([^&]*)/gi, function(m, key, value) {
	  params[key] = value;
	});

	if (params.layers) {
	  var activeLayers = params.layers.split(',').map(function(item) { // map function not supported in IE < 9
		return layers[item];
	  });
	}		  
    function tempTextWhite(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #FFFFFF; stroke: #FFFFFF;stroke-width:2.5;font-size: 10px;">' + txt + '</text></svg>';}
    function tempTextBlue(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #0000FF; stroke: #0000FF;font-size: 10px;">' + txt + '</text></svg>'; }
    function tempTextRed(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #FF0000; stroke: #FF0000;font-size: 10px;">' + txt + '</text></svg>'; }
    function nameText(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #222222; stroke: #222222;font-size: 10px;">' + txt + '</text></svg>'; }
    function nameTextWhite(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #FFFFFF; stroke: #FFFFFF;stroke-width: 2.5;font-size: 10px;">' + txt + '</text></svg>';}
    function wxText(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="60" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #cf0000; stroke: #cf0000;font-size: 10px;">' + txt + '</text></svg>';}
    function wxTextWhite(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="60" height="30"><text x="5" y="10" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #FFFFFF; stroke: #FFFFFF;stroke-width:2.5;font-size: 10px;">' + txt + '</text></svg>';}
    function gustText(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" transform="rotate(30 0,0)" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #FF0000; stroke: #FF0000;font-size: 11px;">' + txt + '</text></svg>'; }
    function gustTextWhite(txt) { return '<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="30" height="30"><text x="5" y="10" transform="rotate(30 0,0)" style="pointer-events: none; font-family: Arial; text-anchor: start;fill: #FFFFFF; stroke: #FFFFFF;stroke-width:2.5;font-size: 11px;">' + txt + '</text></svg>'; }

function FindReference(lat, lng)
{
	for (var index in airports)
	{
		val = airports[index];
		lat1 = val[0];
		lng1 = val[1];
		if ( (lat > (lat1 - latlng_range)) && (lat < (lat1 + latlng_range)) && (lng > (lng1 - latlng_range)) && (lng < (lng1 + latlng_range)))
		{
			return index;
		}
		
	}
	return "";
}

function showCoordinates (e) {
	alert(e.latlng);
}

function centerMap (e) {
	map.panTo(e.latlng);
}

function showWeather(e) {
	f = FindReference(e.latlng.lat,e.latlng.lng);
	if (f)
	{
		str = "https://forecast.weather.gov/MapClick.php?lon="+e.latlng.lng+"&lat="+e.latlng.lat;
		window.open(str);
	}
}
	
   function init() {
			  var mapBounds = new L.LatLngBounds(
`
	message = message + fmt.Sprintf("new L.LatLng(%f, %f),\nnew L.LatLng(%f, %f));\n", LatMinMap, LngMinMap, LatMaxMap, LngMaxMap)
	message = message + `                  
			  var mapMinZoom = 6;
			  var mapMaxZoom = 11;
			map = L.map('map', {
			contextmenu: true,
						contextmenuItems: [{
							text: 'Show coordinates',
							callback: showCoordinates
						}, {
							text: 'Center map here',
							callback: centerMap
						}, {
							text: 'fuck off'
						}]			
			});
			//.setView([42.7611667,-87.8139167], 9);
			var myRenderer = L.canvas({ padding: 0.5 });

/*		
		map.on('contextmenu', function(e) {
			//alert(FindReference(e.latlng.lat,e.latlng.lng));
			f = FindReference(e.latlng.lat,e.latlng.lng);
			if (f)
			{
				str = "https://forecast.weather.gov/MapClick.php?lon="+e.latlng.lng+"&lat="+e.latlng.lat;
				window.open(str);
			}
		});	
*/		
		  var json_obj = JSON.parse(Get("http://www.linair.net/cgi-bin/test.cgi?req=airports&bounds=-88,41,-87,41.5"));
		  console.log(json_obj[1].Lng);
		  
		  var airplaneIcon = L.icon({iconUrl: 'images/airplane.png',iconSize: [24, 24],iconAnchor: [0, 0], popupAnchor: [-3, -5]});
		  var lightningIcon = L.icon({iconUrl: 'images/lightning.png',iconSize: [24, 24],iconAnchor: [0, 0], popupAnchor: [-3, -5]});
		  var tempMarkers =   new L.FeatureGroup();
		  var nameMarkers =   new L.FeatureGroup();
		  var gustMarkers =   new L.FeatureGroup();
		  var precipMarkers = new L.FeatureGroup();
		  var barbMarkers =   new L.FeatureGroup();
		  var lightningMarkers = new L.FeatureGroup();
		  var dotMarkers =    new L.FeatureGroup();`
	// Add our code here
	message = message + `
		  var latlngRAC = [42.7611667,-87.8139167];
		  var latlngENW = [42.5956944,-87.9278056];
		  var latlngMKE = [42.9469444,-87.8970556];
		  var latlngUGN = [42.4221492,-87.8679192];
		  var directions = [0, 90];
		  
		  var ccline_latlng = [latlngRAC,latlngUGN,latlngENW];

		  overlay = L.tileLayer('http://www.linair.net/map/tiles/{z}/{x}/{y}.png', {
				minZoom: mapMinZoom, maxZoom: mapMaxZoom,
				bounds: mapBounds,
				attribution: 'Chicago Sectional',opacity: 0.99,
				tms: true,
				preferCanvas: true,
			  }).addTo(map);
				L.control.radar({}).addTo(map);		
		// If passed on the command line, set the view to what the command line requested
		if (params.station)
		{
			station = params.station;
			statLocation = airports[station.toUpperCase()];
			if (statLocation)
			{
				z = 9;
				if (params.zoom)
					z = params.zoom;
				map.setView(statLocation, z);
			}
			else
			{
				alert("Location " + station + " not found");
			}
			
		} else
		if ((params.lat) && (params.lng) && (params.zoom))
		{

			lat = params.lat;
			lng = params.lng;
			z = params.zoom;
			map.setView([lat, lng], z);
		}
		else
		{
			if (!map.restoreView()) {
				map.setView([42.7611667,-87.8139167], 9);
		
		}	 
	}		
	
	function getStationLatLng(station)
	{
		var ret = {};
		ret['lat'] = 42.7;
		ret['lng'] = -88;
		return ret;
	}
	function scaleIconZoom(zoomLevel) {
		var zValue;
		
		zValue = zoomLevel;
		z = map.getZoom();
		if (z < 5)
		{
			zValue = zValue / 4;
		}
		else if (z < 6)
		{
			zValue = zValue / 3;
		}
		else if (z < 10)
		{
			zValue = zValue / 2;
		}
		return zValue;
	}
	

	function addStation(idval, lat, lng, stationcolor, winddir, barbvel, windspeed, windgust, temperature, stationname, precip, METAR, ltg) {
	
		metarStr = "<small>"+METAR+"</SMALL>";
		vp = L.viewpoint([lat,lng], { id:idval,radius: 8, contextmenu: true, color: 'black', weight: 1, fillColor: stationcolor, 
		                 fillOpacity: 0.9, direction: winddir, 
						    barb: { width: 5, height: 10, offset: 5, stroke: true, color: 'black', weight: 3, opacity: 1,
						    fill: false,fillColor: 'black',fillOpacity: 1, velocity: barbvel },
						contextmenuItems: [{
							text: 'Weather for '+stationname ,
							callback: showWeather
						}, {
							text: 'Airnav for '+stationname,
						}]
		}).bindPopup(metarStr).addTo(map);
		barbMarkers.addLayer(vp);	
		
		if ( windspeed != windgust) {
			var thisgustText = "G" + windgust;
			
			imgGA = 'data:image/svg+xml,' + encodeURIComponent(gustTextWhite(thisgustText))
			icon = L.icon({ iconUrl: imgGA, iconSize: [30, 30],iconAnchor: [0, 0]});
			
			var marker1 = L.marker([lat,lng], {icon: icon, interactive: false}).addTo(map);
			icon = L.icon({ iconUrl: imgGA, iconSize: [30, 30], iconAnchor: [0, 0] });
			
			var gustmarker = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
			gustMarkers.addLayer(marker1);
			gustMarkers.addLayer(gustmarker);
			
			imgGB = 'data:image/svg+xml,' + encodeURIComponent(gustText(thisgustText))
			icon = L.icon({ iconUrl: imgGB,iconSize: [30, 30],iconAnchor: [0, 0]});
			var gustmarker2 = L.marker([lat,lng], {icon: icon, interactive: false, renderer: myRenderer}).addTo(map);
			gustMarkers.addLayer(gustmarker2);
		
		}
		
		// Temperature
		if (temperature != -999) 
		{
			imgTA = 'data:image/svg+xml,' + encodeURIComponent(tempTextWhite(temperature))
			icon = L.icon({ iconUrl: imgTA,iconSize: [30, 30],iconAnchor: [12, 20]});
			var tempmarker = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
			tempMarkers.addLayer(tempmarker);
			if (temperature > 32)
				imgTB = 'data:image/svg+xml,' + encodeURIComponent(tempTextRed(temperature))
			else
				imgTB = 'data:image/svg+xml,' + encodeURIComponent(tempTextBlue(temperature))
			
			icon = L.icon({ iconUrl: imgTB,iconSize: [30, 30],iconAnchor: [12, 20]});
			var tempmarker2 = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
			tempMarkers.addLayer(tempmarker2);
		}
		// Name text
		imgWA = 'data:image/svg+xml,' + encodeURIComponent(nameTextWhite(stationname))
		icon = L.icon({ iconUrl: imgWA,iconSize: [30, 30],iconAnchor: [33, 6]});
		
		var namemarker = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
		nameMarkers.addLayer(namemarker);
		imgWB = 'data:image/svg+xml,' + encodeURIComponent(nameText(stationname))
		icon = L.icon({ iconUrl: imgWB,iconSize: [30, 30],iconAnchor: [33, 6]});
		var namemarker = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
		nameMarkers.addLayer(namemarker);
		
		
		// Precip if any
		if ( precip != '' ) {
			imgBc = 'data:image/svg+xml,' + encodeURIComponent(wxTextWhite(precip))
			icon = L.icon({ iconUrl: imgBc,iconSize: [60, 30],iconAnchor: [-2, 6]});
			var pmarker = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
			precipMarkers.addLayer(pmarker);
			imgWB = 'data:image/svg+xml,' + encodeURIComponent(wxText(precip))
			icon = L.icon({ iconUrl: imgWB,iconSize: [60, 30],iconAnchor: [-2, 6]});
			var pmarker = L.marker([lat,lng], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);
			precipMarkers.addLayer(pmarker);

		}
		// Lightning if any
		if (ltg != 0)
		{
			imgLtg = 'images/lightning.png';
			zv = 24; //scaleIconZoom(24);
			var licon = L.icon({ iconUrl: imgLtg, iconSize: [zv,zv], iconAnchor: [0,30]});
			var ltgMarker = L.marker([lat,lng], {icon: licon, interactive: false,renderer: myRenderer}).addTo(map);
			lightningMarkers.addLayer(ltgMarker);
		}
		
	}
	
	`
	val := 0

	//overlay = L.tileLayer('/map/tileserver.php?/index.json?/1/{z}/{x}/{y}.jpg', {
	for i := range airports {
		//		ApICAO := airports[i].ICAO
		//		ApLat := airports[i].Lat
		//		ApLng := airports[i].Lng
		MetarIndex := FindMeter(airports[i].ICAO)
		TafIndex := FindTaf(airports[i].ICAO)
		WindIndex := FindWinds(airports[i].ICAO)
		// Do we have a metar?
		if MetarIndex != -1 {
			MetarString := metars[MetarIndex].METAR
			wdir, wspeed, wgust := getWinds(MetarString)
			fmt.Printf(" - Winds %d %d gust %d metar %s\n", wdir, wspeed, wgust, MetarString)
			cond := getCondition(metars[MetarIndex].COND)
			precip := getPrecip(MetarString)
			mstr := fmt.Sprintf("<small>%s</small>", MetarString)
			pstr := mstr
			if TafIndex != -1 {
				TafString := tafs[TafIndex].TAF
				var taftmp string
				if strings.Contains(TafString, "<br>") {
					taftmp = strings.Replace(TafString, " FM", "<b> FM</b>", -1)
				} else {
					taftmp = strings.Replace(TafString, " FM", "<br><b>FM</b>", -1)
				}

				fmt.Printf("Has TAF %s\n", taftmp)
				pstr = pstr + fmt.Sprintf("<br><small>%s</small>", taftmp)
			}
			if WindIndex != -1 {
				WindString := winds[WindIndex].Winds
				fmt.Printf("Has Winds %s\n", WindString)
				pstr = pstr + fmt.Sprintf("<small><br>%s</small>", WindString)

			}
			val = val + 1

			if 1 == 1 {
				tt := getTemperature(MetarString)
				if len(tt) == 0 {
					tt = "-999"
				}
				sk := stripK(airports[i].ICAO)
				barbvel := wspeed
				barbvel = barbvel / 5
				if wspeed == 0 {
					barbvel = -1
				}
				ltg := getLightning(MetarString)
				
				message = message + fmt.Sprintf("   addStation(%d, %s, %s, '%s', %d, %d, %d, %d, %s, '%s', '%s', '%s', %d);\n", val, airports[i].Lat, airports[i].Lng, cond, wdir+180, barbvel, wspeed, wgust, tt, sk, precip, pstr, ltg)
				fmt.Println("Airport " + airports[i].ICAO + " [" + airports[i].Lat + "," + airports[i].Lng + "]")
				
			} else { 
				message = message + fmt.Sprintf("     vp = L.viewpoint([%s,%s], { id:%s,radius: 8, color: 'black', weight: 1, fillColor: '%s', fillOpacity: 0.9, direction: %d, ",
					airports[i].Lat, airports[i].Lng, strconv.Itoa(val), cond, wdir+180)
				windspeed := wspeed
				windspeed = windspeed / 5
				if wspeed == 0 {
					windspeed = -1
				}
				message = message + fmt.Sprintf("\nbarb: { width: 5, height: 10, offset: 5, stroke: true, color: 'black', weight: 3, opacity: 1,fill: false,fillColor: 'black',fillOpacity: 1, velocity: %d, } }).bindPopup(\"%s\").addTo(map);\n", windspeed, pstr)
	            message = message + fmt.Sprintf(" barbMarkers.addLayer(vp);\n")

//				message = message + fmt.Sprintf("     lmkr = L.marker([%s,%s], { icon: icon })\n", airports[i].Lat, airports[i].Lng);
//				message = message + ` dotMarkers.addLayer(lmkr);
//				`
				
				
				if wgust != wspeed {
					message = message + fmt.Sprintf("imggw%d = 'data:image/svg+xml,' + encodeURIComponent(gustTextWhite(\"G%d\"))\n", i, wgust)
					message = message + fmt.Sprintf("icon = L.icon({ iconUrl: imggw%d, iconSize: [30, 30], iconAnchor: [0, 0] });\n", i)
					message = message + fmt.Sprintf("var marker1 = L.marker([%s,%s], {icon: icon, interactive: false}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
					message = message + `
						  icon = L.icon({ iconUrl: imggw` + strconv.Itoa(i) + `, iconSize: [30, 30], iconAnchor: [0, 0] });
						  var gustmarker = L.marker([`
					message = message + airports[i].Lat + "," + airports[i].Lng
					message = message + `], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);	
					gustMarkers.addLayer(marker1);
					gustMarkers.addLayer(gustmarker);
					`

					message = message + "img" + strconv.Itoa(i) + " = 'data:image/svg+xml,' + encodeURIComponent(gustText(\"G" + strconv.Itoa(wgust) + "\"))"
					message = message + `
						  icon = L.icon({iconUrl: img` + strconv.Itoa(i) + `,iconSize: [30, 30],iconAnchor: [0, 0] });
						  var gustmarker2 = L.marker([`
					message = message + airports[i].Lat + "," + airports[i].Lng
					message = message + `], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);	
					gustMarkers.addLayer(gustmarker2);
					`

				}

				tt := getTemperature(MetarString)
				// Add temperature
				message = message + fmt.Sprintf("imgTA%d = 'data:image/svg+xml,' + encodeURIComponent(tempTextWhite(\"%s\"))\n", i, tt)
				message = message + fmt.Sprintf("icon = L.icon({ iconUrl: imgTA%d,iconSize: [30, 30],iconAnchor: [12, 20]});\n", i)
				message = message + fmt.Sprintf("var tempmarker = L.marker([%s,%s], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
				ttval, err := strconv.Atoi(tt)
				if err != nil {
				}
				if ttval <= 32 {
					message = message + fmt.Sprintf("imgT%d = 'data:image/svg+xml,' + encodeURIComponent(tempTextBlue(\"%s\"))\n", i, tt)
				} else {
					message = message + fmt.Sprintf("imgT%d = 'data:image/svg+xml,' + encodeURIComponent(tempTextRed(\"%s\"))\n", i, tt)
				}
				message = message + fmt.Sprintf("      tempMarkers.addLayer(tempmarker);\n")
				message = message + fmt.Sprintf("      icon = L.icon({iconUrl: imgT%d,iconSize: [30, 30],iconAnchor: [12, 20] });\n", i)
				message = message + fmt.Sprintf("      var tempmarker2 = L.marker([%s,%s], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
				message = message + fmt.Sprintf("      tempMarkers.addLayer(tempmarker2);\n")
                   
				// Add Names

				message = message + fmt.Sprintf("imgWA%d = 'data:image/svg+xml,' + encodeURIComponent(nameTextWhite(\"%s\"))\n", i, stripK(airports[i].ICAO))
				message = message + fmt.Sprintf("icon = L.icon({ iconUrl: imgWA%d,iconSize: [30, 30],iconAnchor: [33, 6]});\n", i)
				message = message + fmt.Sprintf("var namemarker = L.marker([%s,%s], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
				message = message + fmt.Sprintf("      nameMarkers.addLayer(namemarker);\n")

				message = message + fmt.Sprintf("imgA%d = 'data:image/svg+xml,' + encodeURIComponent(nameText(\"%s\"))\n", i, stripK(airports[i].ICAO))
				message = message + fmt.Sprintf("icon = L.icon({ iconUrl: imgA%d,iconSize: [30, 30],iconAnchor: [33, 6]});\n", i)
				message = message + fmt.Sprintf("var namemarker = L.marker([%s,%s], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
				message = message + fmt.Sprintf("      nameMarkers.addLayer(namemarker);\n")

				// Add precip
				if precip != "" {

					message = message + fmt.Sprintf("imgBc%d = 'data:image/svg+xml,' + encodeURIComponent(wxTextWhite(\"%s\"))\n", i, precip)
					message = message + fmt.Sprintf("icon = L.icon({ iconUrl: imgBc%d,iconSize: [60, 30],iconAnchor: [-2, 6]});\n", i)
					message = message + fmt.Sprintf("var pmarker = L.marker([%s,%s], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
    				message = message + fmt.Sprintf("      precipMarkers.addLayer(pmarker);\n")

					message = message + fmt.Sprintf("imgBc%d = 'data:image/svg+xml,' + encodeURIComponent(wxText(\"%s\"))\n", i, precip)
					message = message + fmt.Sprintf("icon = L.icon({ iconUrl: imgBc%d,iconSize: [60, 30],iconAnchor: [-2, 6]});\n", i)
					message = message + fmt.Sprintf("var pmarker = L.marker([%s,%s], {icon: icon, interactive: false,renderer: myRenderer}).addTo(map);\n", airports[i].Lat, airports[i].Lng)
    				message = message + fmt.Sprintf("      precipMarkers.addLayer(pmarker);\n")
				}
				message = message + "\n"
				fmt.Println("Airport " + airports[i].ICAO + " [" + airports[i].Lat + "," + airports[i].Lng + "]")
			}
		}
	}
	for i := range pireps {
		message = message + fmt.Sprintf("L.marker([%s,%s], {icon: airplaneIcon, title: \"%s\",renderer: myRenderer}).bindPopup('%s').addTo(map);\n", pireps[i].Lat, pireps[i].Lng, pireps[i].Report, pireps[i].Report)
	}
	// close init function
//	var polyline = L.polyline(ccline_latlng, {color: 'red'}).addTo(map);	
	message = message + `
			map.on('zoomend', function() {
				map.removeLayer(lightningMarkers);
				map.addLayer(lightningMarkers);
				if (map.getZoom() <7){
						map.removeLayer(nameMarkers);
				}
				else {
						map.addLayer(nameMarkers);
					}
				if (map.getZoom() <8){
						map.removeLayer(barbMarkers);
						map.addLayer(dotMarkers);
				}
				else {
						map.addLayer(barbMarkers);
						map.removeLayer(dotMarkers);
					}
				
				if (map.getZoom() <9){
						map.removeLayer(precipMarkers);
						map.removeLayer(tempMarkers);
						map.removeLayer(gustMarkers);
				}
				else {
						map.addLayer(precipMarkers);
						map.addLayer(tempMarkers);
						map.addLayer(gustMarkers);
					}
			});    
	  }
	  `
//	  var airports = {};
//	  `
	  
//	for j := range airports {
//		message = message + fmt.Sprintf("   airports['%s'] = [%s,%s];\n",airports[j].ICAO,airports[j].Lat,airports[j].Lng);
//	}
	w.Write([]byte(message))
	w.Close()
	//	  fmt.Println((message))
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
		scanMeters("metars.csv")
		scanTafs("tafs.csv")
		scanPireps("pireps.csv")
	} else {
		scanUatReportFile("dump.txt")
	}
	generateFile("/var/www/html/map/code.js")

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
