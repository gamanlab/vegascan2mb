package main

import (
	"io/ioutil"
	"net/http"
)

func FetchVegaData(url string) (string, error) {
	var err error
	var client = &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(resBody), nil
}
