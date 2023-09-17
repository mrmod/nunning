package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
)

/*
EntryIndex

Maps Entry.Cursor to an Entry
*/
type EntryIndex map[string]Entry

/*
TermFrequencyEntryIndex

Maps words to the Cursor of the Entry they appear in
*/
type TermFrequencyEntryIndex map[string]map[string]Entry

/*
AddEntryToIndex
Index entries on their cursor field
*/
func AddEntryToIndex(entry Entry, index EntryIndex) {
	index[entry.Cursor] = entry
}

/*
UpdateTFIndex

Updates or inserts an Entry for each of its terms
*/
func UpdateTFIndex(e Entry, index TermFrequencyEntryIndex, transformers ...TokenTransformer) {

	tokens := Tokenize(e.Message)

	for _, transformer := range transformers {
		tokens = transformer(tokens)
	}

	for _, token := range tokens {
		if _, ok := index[token]; !ok {
			index[token] = make(map[string]Entry)
		}
		index[token][e.Cursor] = e
	}
}

/*
	Search

Exactly match terms in the index
*/
func Search(word string, tfi TermFrequencyEntryIndex, ei EntryIndex) []Entry {
	entries := []Entry{}
	for cursor, _ := range tfi[word] {
		entries = append(entries, ei[cursor])
	}
	return entries
}

/*
FuzzySearch
Case-insensitive search for terms like the given term or partial term
*/
func FuzzySearch(term string, tfi TermFrequencyEntryIndex, ei EntryIndex) []Entry {
	entries := []Entry{}
	for word, entryIndex := range tfi {
		if matched, err := regexp.MatchString(fmt.Sprintf(".*%s.*", strings.ToLower(term)), strings.ToLower(word)); matched && err == nil {
			log.Printf("Matched %s to %s", term, word)
			for _, entry := range entryIndex {
				entries = append(entries, entry)
			}
		}
	}
	return entries
}

/* IndexJournal
 * Index a journal file
 */
func IndexJournal(journalFile *string, server *IndexServer) error {
	if journalFile == nil {
		return errors.New("journalFile cannot be nil")
	}
	entries := LoadJournal(*journalFile)

	for _, entry := range entries {
		AddEntryToIndex(entry, server.EntryIndex)
		UpdateTFIndex(entry, server.TermFrequencyEntryIndex, server.Transformers...)
	}

	return nil
}
