package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"strings"
	"time"
)

const xsrfTokenKey = "X-XSRF-TOKEN"

var execute = flag.Bool("execute", false, "Execution mode")

type labelInfo struct {
	Type 	string `json:"type"`
	Value 	string `json:"value"`
}

func main() {
	flag.Parse()

	if !*execute {
		os.Stderr.WriteString("ERROR Unsupported execution mode")
		os.Exit(1)
	}

	payloadInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR %s", err.Error())
		os.Exit(2)
	}

	err = processPayload(payloadInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR %s", err.Error())
		os.Exit(3)
	}

	os.Exit(0)
}

func processPayload(payloadInput []byte) error {

	payload := make(map[string]interface{})
	err := json.Unmarshal(payloadInput, &payload)
	if err != nil {
		return fmt.Errorf("ERROR %s", err.Error())
	}

	settingsBytes, err := json.Marshal(payload["configuration"])
	if err != nil {
		return fmt.Errorf("ERROR %s", err.Error())
	}

	settings := make(map[string]string)
	err = json.Unmarshal(settingsBytes, &settings)
	if err != nil {
		return fmt.Errorf("ERROR %s", err.Error())
	}

	return createAndSendIncident(settings)
}

func createAndSendIncident(settings map[string]string) error {

	now := time.Now().UTC().Format(time.RFC3339Nano)
	var occured string
	i, err := strconv.ParseInt(strings.Split(settings["occured"], ".")[0], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN Illegal occured epoch time %s", settings["occured"])
		occured = now
	} else {
		occured = time.Unix(i, 0).String()
	}

	severityMap := map[string]int{
		"Unknown":  0,
		"Low":      1,
		"Medium":   2,
		"High":     3,
		"Critical": 4,
	}
	severity := severityMap[settings["severity"]]

	labels := make([]labelInfo, 0, 1)
	labelsFromUser := strings.Split(settings["labels"], ",")
	if len(labelsFromUser) > 0 && len(labelsFromUser[0]) > 0 {
		for _, pair := range labelsFromUser {
			label := strings.Split(pair, ":")
			if len(label) == 2 {
				labels = append(labels, labelInfo{Type: label[0], Value: label[1]})
			}
		}
	}

	incident := make(map[string]interface{})
	incident["artifacts"] = []string{}
	incident["details"] = settings["details"]
	incident["occured"] = occured
	incident["createInvestigation"] = (settings["investigate"] == "1")
	incident["created"] = now
	incident["name"] = settings["name"]
	incident["tasks"] = []string{}
	incident["status"] = 0
	incident["owner"] = ""
	incident["evidence"] = []string{}
	incident["version"] = 0
	incident["type"] = ""
	incident["id"] = ""
	incident["insights"] = 0
	incident["severity"] = severity
	incident["labels"] = labels

	return sendIncident(settings, incident)
}

func sendIncident(settings map[string]string, incident map[string]interface{}) error {

	baseURL := settings["base_url"]
	credentials := fmt.Sprintf(`{"user":"%s", "password":"%s"}`, settings["username"], settings["password"])
	client, xsrfToken, err := createClientAndLogin(baseURL, credentials)
	if err != nil {
		return err
	}

	requestBody, err := json.Marshal(incident)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", baseURL+"/incident", bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")
	req.Header.Add(xsrfTokenKey, xsrfToken)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("ERROR Response %s", res.Status)
	}

	return nil
}

func createClientAndLogin(baseURL, credentials string) (*http.Client, string, error) {
	client, xsrfToken, err := createClient(baseURL)
	if err != nil {
		return nil, "", err
	}
	loginResult, err := login(client, baseURL, xsrfToken, credentials)
	if err != nil {
		return nil, "", err
	}
	if loginResult != http.StatusOK {
		return nil, "", fmt.Errorf("ERROR Bad login response: %d", loginResult)
	}
	return client, xsrfToken, nil
}

func createClient(baseURL string) (*http.Client, string, error) {
	cookieJar, _ := cookiejar.New(nil)

	client := getClient()
	client.Jar = cookieJar
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.Transport = tr
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	return client, getXSRFToken(resp), nil
}

func getXSRFToken(resp *http.Response) string {
	var xsrfToken string
	for _, element := range resp.Cookies() {
		if element.Name == "XSRF-TOKEN" {
			xsrfToken = element.Value
		}
	}
	return xsrfToken
}

func getClient() *http.Client {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{Transport: tr}
}

func login(client *http.Client, baseURL, xsrfToken, credentials string) (int, error) {
	var jsonStr = []byte(credentials)
	req, err := http.NewRequest("POST", baseURL+"/login", bytes.NewBuffer(jsonStr))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")
	req.Header.Add(xsrfTokenKey, xsrfToken)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}
