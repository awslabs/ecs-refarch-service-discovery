// Copyright 2016. Amazon Web Services, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Spec represents the ErrorResponse response object.
type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/book/{isbn}", BookHandler).Methods("GET")
	r.HandleFunc("/game/{name}", GameHandler).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(GetHtmlFileDir())))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":80", nil))
}

// Make API call to Goodreads microservice
func BookHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isbn := vars["isbn"]
	log.Printf("ISBN is %s\n", isbn)

	goodreads_endpoint := GetGoodreadsEndpoint()
	url := "http://" + goodreads_endpoint + "/book/" + isbn
	log.Printf("URL is %v\n", url)

	user, pass := GetGoodreadsCredentials()
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
		errorMsg := ErrorResponse{Error: response.Status}
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

// Make API call to Twitch microservice
func GameHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	log.Printf("Name is %s\n", name)

	twitch_endpoint := GetTwitchEndpoint()
	url := "http://" + twitch_endpoint + "/game/" + name
	log.Printf("URL is %v\n", url)

	user, pass := GetTwitchCredentials()
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
		errorMsg := ErrorResponse{Error: response.Status}
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

// GetGoodreadsEndpoint gets the Goodreads service endpoint from environment variable
func GetGoodreadsEndpoint() string {
	goodreadsapp_endpoint := os.Getenv("GOODREADSAPPENDPOINT")
	if len(goodreadsapp_endpoint) > 0 {
		return goodreadsapp_endpoint
	}
	log.Println("env variable GOODREADSAPPENDPOINT not found. Using default: goodreadsapp.ecs.internal:80")
	return "goodreadsapp.ecs.internal:80"
}

// GetTwitchEndpoint gets the twitch service endpoint from environment variable
func GetTwitchEndpoint() string {
	twitchapp_endpoint := os.Getenv("TWITCHAPPENDPOINT")
	if len(twitchapp_endpoint) > 0 {
		return twitchapp_endpoint
	}
	log.Println("env variable TWITCHAPPENDPOINT not found. Using default: twitchapp.ecs.internal:80")
	return "twitchapp.ecs.internal:80"
}

// GetGoodreadsCredentials gets the goodreads service credentials
func GetGoodreadsCredentials() (string, string) {
	goodreads_user := os.Getenv("GOODREADS_USERNAME")
	goodreads_pass := os.Getenv("GOODREADS_PASSWORD")
	if len(goodreads_user) > 0 && len(goodreads_pass) > 0 {
		return goodreads_user, goodreads_pass
	}
	log.Println("env variables GOODREADS_USERNAME & GOODREADS_PASSWORD not found. Returning default values: admin/password.")
	return "admin", "password"
}

// GetTwitchCredentials gets the twitch service credentials
func GetTwitchCredentials() (string, string) {
	twitch_user := os.Getenv("TWITCH_USERNAME")
	twitch_pass := os.Getenv("TWITCH_PASSWORD")
	if len(twitch_user) > 0 && len(twitch_pass) > 0 {
		return twitch_user, twitch_pass
	}
	log.Println("env variables TWITCH_USERNAME & TWITCH_PASSWORD not found. Returning default values: admin/password.")
	return "admin", "password"
}

// GetHtmlFileDir gets the html path
func GetHtmlFileDir() string {
	html_file_dir := os.Getenv("HTML_FILE_DIR")
	if len(html_file_dir) > 0 {
		return html_file_dir
	}
	log.Println("env variables HTML_FILE_DIR not found. Returning default values: ./public/")
	return "./public/"
}
