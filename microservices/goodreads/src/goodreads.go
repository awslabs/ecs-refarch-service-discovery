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

type Book struct {
	Isbn          string  `json:"isbn"`
	RatingsCount  int64   `json:"ratingscount"`
	ReviewsCount  int64   `json:"reviewscount"`
	AverageRating float64 `json:"averagerating"`
	ContainerId   string  `json:"containerid"`
	InstanceId    string  `json:"instanceid"`
}

type NoBook struct {
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
	r.HandleFunc("/book/{isbn}", BookHandler).Methods("GET")
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

// Handler to process the GET HTTP request. Calls the Goodreads API,
// processes the JSON response and sends back a JSON response to the client
func BookHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isbn := vars["isbn"]
	log.Printf("ISBN is %s\n", isbn)

	v := url.Values{}
	v.Add("isbns", isbn)

	s := "http://www.goodreads.com/book/review_counts.json?" + v.Encode()
	log.Printf("Sending GET request %s\n", s)
	r := GetHttpResponse(s)
	log.Printf("Got response %s\n", r)

	ratingscount, reviewscount, averagerating := GetData(r)

	if ratingscount == -99 {
		nobook := NoBook{Error: "error - ISBN not found",
			ContainerId: GetContainerId(), InstanceId: GetInstanceId()}
		if err := json.NewEncoder(res).Encode(nobook); err != nil {
			log.Panic(err)
		}
	} else {
		book := Book{Isbn: isbn, RatingsCount: ratingscount, ReviewsCount: reviewscount,
			AverageRating: averagerating, ContainerId: GetContainerId(),
			InstanceId: GetInstanceId()}
		if err := json.NewEncoder(res).Encode(book); err != nil {
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

// set the HTTP BASIC AUTH username based on env variable GOODREADS_USERNAME
// if not present return "admin"
func GetHttpUsername() string {
	if http_username := os.Getenv("GOODREADS_USERNAME"); len(http_username) > 0 {
		return http_username
	}
	return "admin" // default username
}

// set the HTTP BASIC AUTH username based on env variable GOODREADS_USERNAME
// if not present return "admin"
func GetHttpPassord() string {
	if http_password := os.Getenv("GOODREADS_PASSWORD"); len(http_password) > 0 {
		return http_password
	}
	return "password" // default password
}

// unmarshal json response
func GetData(jsonResponse string) (int64, int64, float64) {

	log.Printf("jsonResponse string is %s\n", jsonResponse)

	type ReviewCountsJson struct {
		Books []struct {
			ID                   int64   `json:"id"`
			AverageRating        float64 `json:"average_rating,string"`
			Isbn                 int64   `json:"isbn,string"`
			Isbn13               int64   `json:"isbn13,string"`
			RatingsCount         int64   `json:"ratings_count"`
			ReviewsCount         int64   `json:"reviews_count"`
			TextReviewsCount     int64   `json:"text_reviews_count"`
			WorkRatingsCount     int64   `json:"work_ratings_count"`
			WorkReviewsCount     int64   `json:"work_reviews_count"`
			WorkTextReviewsCount int64   `json:"work_text_reviews_count"`
		} `json:"books"`
	}

	book := ReviewCountsJson{}

	err := json.Unmarshal([]byte(jsonResponse), &book)
	if err != nil {
		log.Printf("ISBN not found: %s", err)
		return -99, -99, 0.00
	}

	log.Printf("%d, %d, %f\n", book.Books[0].WorkRatingsCount,
		book.Books[0].WorkTextReviewsCount,
		book.Books[0].AverageRating)

	return book.Books[0].WorkRatingsCount,
		book.Books[0].WorkTextReviewsCount,
		book.Books[0].AverageRating
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
