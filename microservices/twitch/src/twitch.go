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
	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type Game struct {
	Name        string `json:"name"`
	Popularity  int64  `json:"popularity"`
	Box         string `json:"box"`
	ContainerId string `json:"containerid"`
	InstanceId  string `json:"instanceid"`
}

type NoGame struct {
	Error       string `json:"error"`
	ContainerId string `json:"containerid"`
	InstanceId  string `json:"instanceid"`
}

// Spec represents the Healthcheck response object.
type Healthcheck struct {
	Status string `json:"status"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/game/{name}", GameHandler).Methods("GET")
	http.HandleFunc("/health", HealthHandler)
	http.Handle("/", (httpauth.SimpleBasicAuth(GetHttpUsername(), GetHttpPassord()))(r))
	log.Fatal(http.ListenAndServe(":80", nil))
}

// Handler to process the healthcheck
func HealthHandler(res http.ResponseWriter, req *http.Request) {
	hc := Healthcheck{Status: "OK"}
	if err := json.NewEncoder(res).Encode(hc); err != nil {
		log.Panic(err)
	}
}

// Handler to process the GET HTTP request. Calls the Twitch API,
// processes the JSON response and sends back a JSON response to the client
func GameHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	name := vars["name"]
	log.Printf("Name is %s\n", name)

	v := url.Values{}
	v.Add("query", name)
	v.Add("type", "suggest")

	s := "https://api.twitch.tv/kraken/search/games?" + v.Encode()
	log.Printf("Sending GET request %s\n", s)
	r := GetHttpResponse(s)
	log.Printf("Got response %s\n", r)

	fullname, popularity, box := GetData(r)

	if popularity == -99 {
		nogame := NoGame{Error: "error - Game not found",
			ContainerId: GetContainerId(), InstanceId: GetInstanceId()}
		if err := json.NewEncoder(res).Encode(nogame); err != nil {
			log.Panic(err)
		}
	} else {
		game := Game{Name: fullname, Popularity: popularity, Box: box,
			ContainerId: GetContainerId(), InstanceId: GetInstanceId()}
		if err := json.NewEncoder(res).Encode(game); err != nil {
			log.Panic(err)
		}
	}
}

// Call the remote web service and return the result as a string
func GetHttpResponse(url string) string {
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		log.Panic(err)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Panic(err)
	}
	return string(contents)
}

// set the HTTP BASIC AUTH username based on env variable TWITCH_USERNAME
// if not present return "admin"
func GetHttpUsername() string {
	if http_username := os.Getenv("TWITCH_USERNAME"); len(http_username) > 0 {
		return http_username
	}
	return "admin" // default username
}

// set the HTTP BASIC AUTH username based on env variable TWITCH_USERNAME
// if not present return "admin"
func GetHttpPassord() string {
	if http_password := os.Getenv("TWITCH_PASSWORD"); len(http_password) > 0 {
		return http_password
	}
	return "password" // default password
}

// Unmsarshal json response
func GetData(jsonResponse string) (string, int64, string) {

	log.Printf("jsonResponse string is %s\n", jsonResponse)

	type Out struct {
		Links struct {
			Self string `json:"self"`
		} `json:"_links"`
		Games []struct {
			ID    int64 `json:"_id"`
			Links struct {
			} `json:"_links"`
			Box struct {
				Large    string `json:"large"`
				Medium   string `json:"medium"`
				Small    string `json:"small"`
				Template string `json:"template"`
			} `json:"box"`
			GiantbombID int64 `json:"giantbomb_id"`
			Logo        struct {
				Large    string `json:"large"`
				Medium   string `json:"medium"`
				Small    string `json:"small"`
				Template string `json:"template"`
			} `json:"logo"`
			Name       string `json:"name"`
			Popularity int64  `json:"popularity"`
		} `json:"games"`
	}

	game := Out{}

	err := json.Unmarshal([]byte(jsonResponse), &game)
	if err != nil {
		log.Printf("Game not found: %s", err)
		return "", -99, ""
	}

	if len(game.Games) == 0 {
		log.Printf("Game not found: %s", jsonResponse)
		return "", -99, ""
	}

	log.Printf("%s, %d, %s\n", game.Games[0].Name,
		game.Games[0].Popularity,
		game.Games[0].Box.Small)

	return game.Games[0].Name,
		game.Games[0].Popularity,
		game.Games[0].Box.Small
}

// GetContainerId gets the ContainerID if exists
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

// GetInstanceId gets the Instance ID if exists
func GetInstanceId() string {
	cmd := "curl"
	cmdArgs := []string{"-s", "http://169.254.169.254/latest/meta-data/instance-id"}
	out, err := exec.Command(cmd, cmdArgs...).Output()
	if err != nil {
		log.Printf("Instance Id err is %s\n", err)
		return ""
	}
	log.Printf("The instance id is %s\n", out)
	return string(out)
}
