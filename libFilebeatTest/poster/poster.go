package poster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func HttpPostJson(url string, v interface{}) (response *http.Response, err error) {
	postBytes, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("Error while encoding to json: %v", err)
	}
	return Post(url, postBytes, "application/json; charset=UTF-8")
}

func Post(url string, postBytes []byte, contentType string) (response *http.Response, err error) {
	buf := bytes.NewBuffer(postBytes)

	req, err := http.NewRequest("POST", url, buf)
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("Error POSTing to %s: %v", url, err)
	}

	return resp, nil
}
