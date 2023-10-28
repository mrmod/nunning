package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	envServiceToken     = "SERVICE_TOKEN"
	envClientCredential = "CLIENT_CREDENTIAL"
)

var (
	serviceToken     = "123"
	clientCredential = "abc"
)

func main() {

	if _serviceToken, ok := os.LookupEnv(envServiceToken); ok {
		log.Printf("Using service token '%s' from environment variable %s", _serviceToken, envServiceToken)
		serviceToken = _serviceToken
	}

	if _clientCredential, ok := os.LookupEnv(envClientCredential); ok {
		log.Printf("Using client credential '%s' from environment variable %s", _clientCredential, envClientCredential)
		clientCredential = _clientCredential
	}
	log.Println("Starting credentialStore service...")

	// Service requiring a service token
	http.HandleFunc("/artifacts/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Authorizing artifact by `SERVICE_TOKEN`: %s", r.URL.Path)
		authorizationHeader := r.Header.Get("Authorization")
		log.Printf("Headers: %#v", r.Header)
		token := strings.TrimSpace(strings.Split(authorizationHeader, "Bearer ")[1])
		if token != serviceToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Printf("Serving artifact: %s", r.URL.Path)
		http.StripPrefix(
			"/artifacts/",
			http.FileServer(http.Dir("artifacts/")),
		).ServeHTTP(w, r)
	})

	// ClientCredential to serviceToken exchange handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		authorizationHeader := r.Header.Get("Authorization")

		if s := strings.Split(authorizationHeader, "Bearer "); len(s) == 2 {
			credential := strings.TrimSpace(s[1])
			log.Printf("Client sent credential: '%s'", credential)
			if credential == clientCredential {
				log.Printf("Exchanged clientCredential '%s' for serviceToken '%s'", credential, serviceToken)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, "{\"message\":\"%s\"}", serviceToken)
				return
			}
		}
		log.Printf("Invalid clientCredential Authorization header: %s", authorizationHeader)
		w.WriteHeader(http.StatusUnauthorized)
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
