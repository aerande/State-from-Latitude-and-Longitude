package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
)

var borderData []BorderData

func init() {
	borderData = getData()
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/stateserver/", getState)
	http.ListenAndServe(":8080", nil)
}

// home page
func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello! Please visit localhost:8080/stateserver/?longitude=<longitude>&latitutde=<latitutde>"))
}

// function to take inputs and find state from border data
func getState(w http.ResponseWriter, r *http.Request) {
	longitude := r.URL.Query().Get("longitude")
	latitude := r.URL.Query().Get("latitude")
	longf, _ := strconv.ParseFloat(longitude, 64)
	latf, _ := strconv.ParseFloat(latitude, 64)
	name, _ := findState(borderData, RoundPlus(longf, 6), RoundPlus(latf, 6))
	w.Write([]byte(name))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// function to clean and load .json data in BorderData format
func getData() []BorderData {
	raw, _ := ioutil.ReadFile("./states.json")
	data := string(raw[:])
	temp := strings.Split(data, "\n")
	var total = make([]BorderData, 0, 0)
	for _, p := range temp {
		var stateData BorderData
		json.Unmarshal([]byte(p), &stateData)
		if stateData.State != "" || len(stateData.Border) != 0 {
			stateData.Border = stateData.Border[:len(stateData.Border)-1]
			for i, _ := range stateData.Border {
				stateData.Border[i][0] = RoundPlus(stateData.Border[i][0], 6)
				stateData.Border[i][1] = RoundPlus(stateData.Border[i][1], 6)
			}
			total = append(total, stateData)
		}
	}
	fmt.Println("Data Cleaned and Loaded")
	return total
}

// function to iterate through all the states to find state where the point is inside.
func findState(borderData []BorderData, lon float64, lat float64) (string, error) {
	for _, state := range borderData {
		if isInside(state.Border, len(state.Border), lon, lat) {
			fmt.Println("result" + state.State)
			return state.State, nil
		}
	}
	return "Not Found", nil
}

// Border data struct
type BorderData struct {
	State  string      `json:"state"`
	Border [][]float64 `json:"border"`
}

/*
	To find the state for a given latitude and longitude, I have a created polygon from the border's latitude and longitude.
	From the latitude and longitude,I extend a horizontal line and check the number of crossingd to see if the point lies inside
	the polygon,if the point crosses odd number of times then the point is inside the state's border and that is the state for the given
	latitude and longitude.

*/

// function to check if the end is between start and latitute and longitude point
func onSegment(start []float64, end []float64, lon float64, lat float64) bool {
	if (end[0] <= math.Max(start[0], lon)) && (end[0] >= math.Min(start[0], lon)) &&
		(end[1] <= math.Max(start[1], lat)) && (end[1] >= math.Min(start[1], lat)) {
		return true
	}
	return false
}

// function to check order of triplet start,end,(lon,lat) 0--colinear,1--clockwise,2--anticlockwise
func orientation(start []float64, end []float64, lon float64, lat float64) int64 {
	val := (end[1]-start[1])*(lon-end[0]) -
		(end[0]-start[0])*(lat-end[1])
	if val == 0 {
		return 0
	}
	if val > 0 {
		return 1
	} else {
		return 2
	}
}

//function to check if line segment p1q1 intersects with p2q2
func doIntersect(p1 []float64, q1 []float64, p2 []float64, q2 []float64) bool {
	o1 := orientation(p1, q1, p2[0], p2[1])
	o2 := orientation(p1, q1, q2[0], q2[1])
	o3 := orientation(p2, q2, p1[0], p1[1])
	o4 := orientation(p2, q2, q1[0], q1[1])
	if o1 != o2 && o3 != o4 {
		return true
	}
	if o1 == 0 && onSegment(p1, p2, q1[0], q1[1]) {
		return true
	}
	if o2 == 0 && onSegment(p1, q2, q1[0], q1[1]) {
		return true
	}
	if o3 == 0 && onSegment(p2, p1, q2[0], q2[1]) {
		return true
	}
	if o4 == 0 && onSegment(p2, q1, q2[0], q2[1]) {
		return true
	}

	return false
}

// Function returns true if (lon,lat) is inside the polygon
func isInside(Point [][]float64, n int, lon float64, lat float64) bool {
	if n < 3 {
		return false
	}

	var p = make([]float64, 2, 2)
	var extreme = make([]float64, 2, 2)
	extreme[0] = math.Inf(1)
	extreme[1] = lat
	p[0] = lon
	p[1] = lat

	count := 0

	if doIntersect(Point[0], Point[1], p, extreme) {
		if orientation(Point[0], p, Point[1][0], Point[1][1]) == 0 {
			return onSegment(Point[0], p, Point[1][0], Point[1][1])
		}

		count = count + 1
	}

	i := 1
	for i != 0 {
		next := (i + 1) % n
		if doIntersect(Point[i], Point[next], p, extreme) {
			if orientation(Point[i], p, Point[next][0], Point[next][1]) == 0 {
				return onSegment(Point[i], p, Point[next][0], Point[next][1])
			}

			count = count + 1
		}
		i = next
	}
	if count%2 == 0 {
		return false
	} else {
		return true
	}
}

// function for rounding of the float64
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
