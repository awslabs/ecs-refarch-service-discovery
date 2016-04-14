package main

import (
  "os"
  "log"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "github.com/gorilla/mux"
)

// Spec represents the ErrorResponse response object.
type ErrorResponse struct {
  Error   string    `json:"error"`
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/stock/{symbol}", StockHandler).Methods("GET")
  r.HandleFunc("/weather/{city}", WeatherHandler).Methods("GET")
  r.PathPrefix("/").Handler(http.FileServer(http.Dir(GetHtmlFileDir())))
  http.Handle("/", r)
  log.Fatal(http.ListenAndServe(":80", nil))
}

// Handler to process the GET HTTP request. Calls the Yahoo Finance service,
// processes the csv response and sends back a JSON response to the client
func StockHandler(res http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  symbol := vars["symbol"]
  log.Printf("Stock symbol is %s\n", symbol)

  stock_endpoint := GetStocksEndpoint()
  url := "http://" + stock_endpoint + "/stocks/" + symbol
  log.Printf("URL is %v\n", url)

  user, pass := GetStocksCredentials()
  log.Printf("Got credentials username= %s, password=%s\n", user, pass)
  req, err := http.NewRequest("GET", url, nil)
  if len(user) > 0 && len(pass) > 0 {
    req.SetBasicAuth(user, pass)
  }
  cli := &http.Client{}
  response, err := cli.Do(req)
  defer response.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  if response.StatusCode != 200 {
    log.Printf("Got Error response: %s\n", response.Status)
    errorMsg := ErrorResponse { Error: response.Status }
    if err := json.NewEncoder(res).Encode(errorMsg); err != nil {
      log.Panic(err)
    }
  } else {
    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
      log.Fatal(err)
    }
    log.Printf("Got Successful response: %s\n", string(contents))
    res.Write(contents)
  }
}

// Handler to process the GET HTTP request. Calls the Yahoo Finance service,
// processes the csv response and sends back a JSON response to the client
func WeatherHandler(res http.ResponseWriter, req *http.Request) {
  vars := mux.Vars(req)
  city := vars["city"]
  log.Printf("City is %s\n", city)

  weather_endpoint := GetWeatherEndpoint()
  url := "http://" + weather_endpoint + "/weather/" + city
  log.Printf("URL is %v\n", url)

  user, pass := GetWeatherCredentials()
  log.Printf("Got credentials username= %s, password=%s\n", user, pass)
  req, err := http.NewRequest("GET", url, nil)

  if len(user) > 0 && len(pass) > 0 {
    req.SetBasicAuth(user, pass)
  }
  cli := &http.Client{}
  response, err := cli.Do(req)
  defer response.Body.Close()
  if err != nil {
    log.Fatal(err)
  }

  if response.StatusCode != 200 {
    log.Printf("Got Error response: %s\n", response.Status)
    errorMsg := ErrorResponse { Error: response.Status }
    if err := json.NewEncoder(res).Encode(errorMsg); err != nil {
      log.Panic(err)
    }
  } else {
    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
      log.Fatal(err)
    }

    log.Printf("Got Successful response: %s\n", string(contents))
    res.Write(contents)
  }
}

// get the stock service endpoint from environment variable
func GetStocksEndpoint() string {
  stocksapp_endpoint := os.Getenv("STOCKSAPPENDPOINT");
  if len(stocksapp_endpoint) > 0 {
    return stocksapp_endpoint
  }
  log.Println("env variable STOCKSAPPENDPOINT not found. Using default: stocksapp.ecs.internal:80")
  return "stocksapp.ecs.internal:80"
}

// get the weather service endpoint from environment variable
func GetWeatherEndpoint() string {
  weatherapp_endpoint := os.Getenv("WEATHERAPPENDPOINT");
  if len(weatherapp_endpoint) > 0 {
    return weatherapp_endpoint
  }
  log.Println("env variable WEATHERAPPENDPOINT not found. Using default: weatherapp.ecs.internal:80")
  return "weatherapp.ecs.internal:80"
}

// get the stock service credentials
func GetStocksCredentials() (string, string) {
  stock_user := os.Getenv("STOCKS_USERNAME");
  stock_pass := os.Getenv("STOCKS_PASSWORD");
  if len(stock_user) > 0 && len(stock_pass) > 0 {
      return stock_user, stock_pass
  }
  log.Println("env variables STOCKS_USERNAME & STOCKS_PASSWORD not found. Returning default values: admin/password.")
  return "admin", "password"
}

// get the weather service credentials
func GetWeatherCredentials() (string, string) {
  weather_user := os.Getenv("WEATHER_USERNAME");
  weather_pass := os.Getenv("WEATHER_PASSWORD");
  if len(weather_user) > 0 && len(weather_pass) > 0 {
    return weather_user, weather_pass
  }
  log.Println("env variables WEATHER_USERNAME & WEATHER_PASSWORD not found. Returning default values: admin/password.")
  return "admin", "password"
}

// get the html path
func GetHtmlFileDir() string {
  html_file_dir := os.Getenv("HTML_FILE_DIR");
  if len(html_file_dir) > 0 {
      return html_file_dir
  }
  log.Println("env variables HTML_FILE_DIR not found. Returning default values: ./public/")
  return "./public/"
}
