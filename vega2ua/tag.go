package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Tag struct {
	Name    string
	NodeId  string
	Address int
	Value   float64
}

var tags = make([]Tag, 0)

func ReadTagsJson() {
	data, err := ioutil.ReadFile("tags.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &tags)
	if err != nil {
		log.Fatal(err)
	}
}
