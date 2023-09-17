package main

import (
	"testing"
	"time"
)

func TestLoadJournal(t *testing.T) {
	j := LoadJournal("small.json")
	if len(j) != 4 {
		t.Errorf("Expected 4 entry, got %d", len(j))
	}
}

func TestTokenizeBase(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		tokens   int
	}{
		{"Hello, world!", []string{"Hello", "world"}, 2},
		{"Started listener on [::]:5140", []string{"Started", "listener", "on", "[", "]", "5140"}, 6},
		{"VideoTrimPrefix: /home/cameras/", []string{"VideoTrimPrefix", "/home/cameras/"}, 2},
	}

	for _, test := range tests {
		tokens := Tokenize(test.input)
		if len(tokens) != test.tokens {
			t.Errorf("Expected %d tokens, got %d", test.tokens, len(tokens))
		}
		for i, token := range tokens {
			if token != test.expected[i] {
				t.Errorf("Expected %s, got %s", test.expected[i], token)
			}
		}
	}
}

func TestTokenizerPipeline(t *testing.T) {
	index := TermFrequencyEntryIndex{}
	UpdateTFIndex(Entry{Message: "Hello, world!"}, index)
	if len(index) != 2 {
		t.Errorf("Expected 2 terms, got %d", len(index))
	}

	index = TermFrequencyEntryIndex{}
	UpdateTFIndex(Entry{Message: "Hello, world!"}, index, iTokenize)
	if len(index) != 3 {
		t.Errorf("Expected 3 terms, got %d", len(index))
	}
}

func TestSmallIndexPeformance(t *testing.T) {
	entries := LoadJournal("small.json")
	entryIndex := EntryIndex{}
	tfIndex := TermFrequencyEntryIndex{}

	for _, entry := range entries {
		AddEntryToIndex(entry, entryIndex)
		UpdateTFIndex(entry, tfIndex)
	}
	if len(entryIndex) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entryIndex))
	}
	if len(tfIndex) != 19 {
		t.Errorf("Expected 19 terms, got %d", len(tfIndex))
	}

	startTime := time.Now()
	result := Search("Started", tfIndex, entryIndex)
	elapsed := time.Since(startTime)

	t.Logf("Search took %s", elapsed)
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	result = FuzzySearch("start", tfIndex, entryIndex)
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
}
