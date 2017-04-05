package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func filebeatTest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)

	var t []map[string]interface{}
	err := decoder.Decode(&t)
	if err != nil {
		fmt.Printf("Oh no there was an error! %v\n", err)
		os.Exit(1)
	}

	printOutput(buildOutput(t))
}

func buildOutput(t []map[string]interface{}) map[string][]string {
	output := make(map[string][]string)
	for _, event := range t {
		m := event["message"].(string)
		s := event["source"].(string)

		existing, ok := output[s]
		if !ok {
			existing = make([]string, 0, 1)
		}
		existing = append(existing, m)
		output[s] = existing
	}

	return output
}

func printOutput(output map[string][]string) {
	for s, ms := range output {
		fmt.Printf("Source: %s\n\n", s)
		for _, m := range ms {
			fmt.Printf("\t%s\n", m)
		}
		fmt.Printf("\n\n")
	}
}

func main() {
	http.HandleFunc("/filebeatTest", filebeatTest)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
