package util

import (
	"encoding/json"
	"fmt"
)

func Print(output interface{}) {
	json, _ := json.Marshal(output)
	fmt.Println(string(json))
}
