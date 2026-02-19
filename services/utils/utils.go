package utils

import (
	"encoding/json"
	"fmt"
)

// PrettyPrint prints the struct in JSON format with indentation.
func PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}
	fmt.Println(string(data))
}
