package main

import (
  "os"
  "os/exec"
  "log"
  "net/http"
  "net/url"
  "encoding/json"
  "io/ioutil"
  "github.com/goji/httpauth"
  "github.com/gorilla/mux"
  "strings"
)

// Spec represents the Weather response object.
type Weather struct {
  City        string    `json:"city"`
  Country     string    `json:"country"`
  Temp        float64   `json:"temp"`
  Format      string    `json:"format"`
  ContainerId string    `json:"containerid"`
  InstanceId  string    `json:"instanceid"`
}

// Spec represents the Healthcheck response object.
type Healthcheck struct {
  Status      string    `json:"status"`
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/weather/{city}", WeatherHandler).Methods("GET")
  http.HandleFunc("/health", HealthHandler)
  http.Handle("/", (httpauth.SimpleBasicAuth(GetHttpUsername(), GetHttpPassord()))(r))
  log.Fatal(http.ListenAndServe(":80", nil))
}

// Handler to process the GET HTTP request. Calls the Yahoo Finance service,
// processes the csv response and sends back a JSON response to the client
func WeatherHandler(res http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  city := vars["city"]
  log.Printf("City is %s\n", city)

  v := url.Values{}
  v.Add("q", city)
  v.Add("units", "metric")

  // get the API Key from the environment variables
  if api_key := os.Getenv("API_KEY"); len(api_key) > 0 {
    v.Add("APPID", api_key)
  }

  s := "http://api.openweathermap.org/data/2.5/weather?" + v.Encode()
  log.Printf("Sending GET request %s\n", s)

  r := GetHttpResponse(s)
  city, country, temp := GetData(r)

  weather := Weather { City: city, Country: country, Temp: temp, Format: "Celsius",
    ContainerId: GetContainerId(), InstanceId: GetInstanceId() }
  if err := json.NewEncoder(res).Encode(weather); err != nil {
    log.Panic(err)
  }
}

// Handler to process the healthcheck
func HealthHandler(res http.ResponseWriter, req *http.Request) {
  hc := Healthcheck { Status: "OK" }
  if err := json.NewEncoder(res).Encode(hc); err != nil {
    log.Panic(err)
  }
}

// Call the remote web service and return the result as a byte array
func GetHttpResponse(url string) []byte {
  response, err := http.Get(url)
  defer response.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  contents, err := ioutil.ReadAll(response.Body)
  if err != nil {
    log.Fatal(err)
  }
  return contents
}

// set the HTTP BASIC AUTH username based on env variable WEATHER_USERNAME
// if not present return "admin"
func GetHttpUsername() string {
  if http_username := os.Getenv("WEATHER_USERNAME"); len(http_username) > 0 {
    return http_username
  }
  return "admin" // default username
}

// set the HTTP BASIC AUTH username based on env variable WEATHER_PASSWORD
// if not present return "admin"
func GetHttpPassord() string {
  if http_password := os.Getenv("WEATHER_PASSWORD"); len(http_password) > 0 {
    return http_password
  }
  return "password" // default password
}

// Get the Country and Temperature from the JSON repsonse object
func GetData(b []byte) (string, string, float64) {
  var f interface{}
  err := json.Unmarshal(b, &f)
  if err != nil {
    log.Fatal(err)
  }
  m := f.(map[string]interface{})

  city := m["name"].(string)
  country := m["sys"].(map[string]interface{})["country"].(string)
  log.Printf("Country is %s\n", country)
  temp := m["main"].(map[string]interface{})["temp"].(float64)
  log.Printf("Temp is %.2f\n", temp)
  return city, country, temp
}

// Get the ContainerID if exists
func GetContainerId() string {
  cmd := "cat /proc/self/cgroup | grep \"docker\" | sed s/\\\\//\\\\n/g | tail -1"
  out, err := exec.Command("bash", "-c", cmd).Output()
  if err != nil {
    log.Printf("Container Id err is %s\n", err)
		return ""
	}
	log.Printf("The container id is %s\n", out)
  return strings.TrimSpace(string(out))
}

// Get the Instance ID if exists
func GetInstanceId() string {
  cmd := "curl"
  cmdArgs := []string{"-s", "http://169.254.169.254/latest/meta-data/instance-id" }
  out, err := exec.Command(cmd, cmdArgs...).Output()
	if err != nil {
    log.Printf("Instance Id err is %s\n", err)
		return ""
	}
	log.Printf("The instance id is %s\n", out)
  return string(out)
}
