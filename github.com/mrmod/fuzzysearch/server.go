package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type IndexServer struct {
	TermFrequencyEntryIndex
	EntryIndex
	Ready        bool
	Transformers []TokenTransformer
}

type Searcher func(string, TermFrequencyEntryIndex, EntryIndex) []Entry

func (server *IndexServer) handleGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling %s", r.URL.Path)
	if !server.Ready {
		log.Printf("Warning: Indexing is not yet complete")
	}
	searcher := Search
	switch path := r.URL.Path; path {
	case "/search":
		term := r.URL.Query().Get("term")
		useFuzzy := r.URL.Query().Get("fuzzy") == "true"
		if useFuzzy {
			log.Printf("Using fuzzy search")
			searcher = FuzzySearch
		}
		log.Printf("Searching for %s", term)
		start := time.Now()

		entries := searcher(term, server.TermFrequencyEntryIndex, server.EntryIndex)
		elapsed := time.Since(start)
		log.Printf("Found %d entries in %s", len(entries), elapsed)
		json.NewEncoder(w).Encode(entries)
	case "/messages":
		term := r.URL.Query().Get("term")
		useFuzzy := r.URL.Query().Get("fuzzy") == "true"
		if useFuzzy {
			log.Printf("Using fuzzy search")
			searcher = FuzzySearch
		}
		log.Printf("Searching for %s", term)
		start := time.Now()
		entries := searcher(term, server.TermFrequencyEntryIndex, server.EntryIndex)
		elapsed := time.Since(start)
		log.Printf("Found %d messages in %s", len(entries), elapsed)
		messages := []string{}
		for _, entry := range entries {
			messages = append(messages, entry.Message)
		}
		json.NewEncoder(w).Encode(messages)
	default:
		log.Printf("Index serving %d entries and %d terms", len(server.EntryIndex), len(server.TermFrequencyEntryIndex))
		json.NewEncoder(w).Encode(map[string]int64{
			"entries": int64(len(server.EntryIndex)),
			"terms":   int64(len(server.TermFrequencyEntryIndex)),
		})
	}
}

func (server *IndexServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[%s] Handling %s", r.Method, r.URL.Path)
	switch method := r.Method; method {
	case "GET":
		server.handleGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
