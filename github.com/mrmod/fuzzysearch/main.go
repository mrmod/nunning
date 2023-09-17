package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	journalFile := flag.String("journal", "journal.json", "Path to journal file")
	port := flag.String("port", ":8080", "Port to listen on")
	addITokenizer := flag.Bool("add-insensitive-tokenizer", false, "Add iTokenize transformer")

	flag.Parse()
	log.Printf("Starting index server on %s", *port)
	indexServer := &IndexServer{
		TermFrequencyEntryIndex: TermFrequencyEntryIndex{},
		EntryIndex:              EntryIndex{},
	}
	if addITokenizer != nil && *addITokenizer {
		indexServer.Transformers = append(indexServer.Transformers, iTokenize)
	}
	go func() {
		log.Printf("Indexing %s", *journalFile)
		start := time.Now()
		IndexJournal(journalFile, indexServer)
		indexServer.Ready = true
		elapsed := time.Since(start)
		log.Printf("Index server ready after %s", elapsed)
	}()
	if err := http.ListenAndServe(*port, indexServer); err != nil {
		log.Fatal(err)
	}
}
