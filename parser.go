package main

import (
	"encoding/json"
)

func decodeFromJson(jsonStr string) {

	if jsonStr == nil {
		panic("Error: attempting to parse json string")
	}

	decoder := json.NewDecoder()

}
