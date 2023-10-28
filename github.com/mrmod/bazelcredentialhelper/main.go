package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type BazelCredentialHelperData struct {
	Headers map[string][]string `json:"headers"`
}

const (
	envClientCredential = "CLIENT_CREDENTIAL"
)

var (
	clientCredential = "123"
)

func main() {

	if _clientCredential, ok := os.LookupEnv(envClientCredential); ok {
		log.Printf("Using client credential from environment variable %s", envClientCredential)
		clientCredential = _clientCredential
	}
	log.Println("Exchanging `CLIENT_CREDENTIAL` '" + clientCredential + "' with credential store")
	res, err := http.DefaultClient.Do(&http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Scheme: "http", Host: "localhost:8080"},
		Header: http.Header{
			"Authorization": []string{"Bearer " + clientCredential},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Status: %d", res.StatusCode)

	defer res.Body.Close()
	credentialStoreResponse := map[string]string{}
	if err := json.NewDecoder(res.Body).Decode(&credentialStoreResponse); err != nil {
		log.Fatalf("Failed to decode credential store response: %s", err)
	}

	serviceToken := credentialStoreResponse["message"]
	log.Println("Got `SERVICE_TOKEN` '" + serviceToken + "' from credential store by exchanging `CLIENT_CREDENTIAL`")

	c, err := json.MarshalIndent(BazelCredentialHelperData{
		Headers: map[string][]string{
			"Authorization": []string{"Bearer " + serviceToken},
		},
	}, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", c)
}
