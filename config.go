package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var (
	config Config
)

func init() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Println(err)
		return
	}
	defer configFile.Close()

	configByte, _ := ioutil.ReadAll(configFile)

	if err := json.Unmarshal(configByte, &config); err != nil {
		log.Println(err)
		return
	}
}
