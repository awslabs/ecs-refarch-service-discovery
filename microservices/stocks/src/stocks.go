package main

import (
  "os"
  "os/exec"
  "log"
  "net/http"
  "net/url"
  "encoding/json"
  "encoding/csv"
  "io/ioutil"
  "strings"
  "strconv"
  "github.com/goji/httpauth"
  "github.com/gorilla/mux"
)

// Spec represents the Stock response object.
type Stock struct {
  Name        string    `json:"name"`
  Symbol      string    `json:"symbol"`
  Price       float64   `json:"price"`
  ContainerId string    `json:"containerid"`
  InstanceId  string    `json:"instanceid"`
}

// Spec represents the Healthcheck response object.
type Healthcheck struct {
  Status      string    `json:"status"`
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/stocks/{symbol}", StockHandler).Methods("GET")
  http.HandleFunc("/health", HealthHandler)
  http.Handle("/", (httpauth.SimpleBasicAuth(GetHttpUsername(), GetHttpPassord()))(r))
  log.Fatal(http.ListenAndServe(":80", nil))
}

// Handler to process the healthcheck
func HealthHandler(res http.ResponseWriter, req *http.Request) {
  hc := Healthcheck { Status: "OK" }
  if err := json.NewEncoder(res).Encode(hc); err != nil {
    log.Panic(err)
  }
}

// Handler to process the GET HTTP request. Calls the Yahoo Finance service,
// processes the csv response and sends back a JSON response to the client
func StockHandler(res http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  symbol := vars["symbol"]
  log.Printf("Stock symbol is %s\n", symbol)

  v := url.Values{}
  v.Add("s", symbol)
  v.Add("f", "nl1r")

  s := "http://download.finance.yahoo.com/d/quotes.csv?" + v.Encode()
  log.Printf("Sending GET request %s\n", s)
  r := GetHttpResponse(s)
  log.Printf("Got response %s\n", r)
  name, price := GetData(r)

  stock := Stock { Name: name, Symbol: symbol, Price: price,
    ContainerId: GetContainerId(), InstanceId: GetInstanceId() }
  if err := json.NewEncoder(res).Encode(stock); err != nil {
    log.Panic(err)
  }
}

// Call the remote web service and return the result as a string
func GetHttpResponse(url string) string {
  response, err := http.Get(url)
  defer response.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  contents, err := ioutil.ReadAll(response.Body)
  if err != nil {
    log.Fatal(err)
  }
  return string(contents)
}

// set the HTTP BASIC AUTH username based on env variable STOCKS_USERNAME
// if not present return "admin"
func GetHttpUsername() string {
  if http_username := os.Getenv("STOCKS_USERNAME"); len(http_username) > 0 {
    return http_username
  }
  return "admin" // default username
}

// set the HTTP BASIC AUTH username based on env variable STOCKS_USERNAME
// if not present return "admin"
func GetHttpPassord() string {
  if http_password := os.Getenv("STOCKS_PASSWORD"); len(http_password) > 0 {
    return http_password
  }
  return "password" // default password
}

// Get the Stock Name and Price from the comma delimited string
func GetData(csv_response string) (string, float64) {
  r := csv.NewReader(strings.NewReader(csv_response))
  line, err := r.Read()
  if err != nil {
    log.Fatal(err)
  }

  price, err := strconv.ParseFloat(line[1],64)
  if err != nil {
    log.Fatal(err)
  }

  return line[0], price
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
