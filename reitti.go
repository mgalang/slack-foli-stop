package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func getJSON(path string, result interface{}) error {
	resp, err := http.Get(path)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

func main() {
	type RecordItem struct {
		Destinationdisplay string
		Aimedarrivaltime   int
	}

	type SiriJSON struct {
		Status     string
		Servertime int
		Result     []RecordItem
	}

	j := SiriJSON{}

	err := getJSON("https://data.foli.fi/siri/sm/T4", &j)

	if err != nil {
		panic(err)
	}

	log.Println(j)
}
