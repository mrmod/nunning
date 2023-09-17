package main

import "strings"

var (
	tokenDelimiters = []string{" ", "\t", ",", ".", "!", "?", ":"}
)

type TokenTransformer func([]string) []string

/*
Tokenize
* Split a string into tokens
TODO: Use an NLP tokenizer
*/
func Tokenize(message string) []string {
	return strings.FieldsFunc(message, func(r rune) bool {
		for _, delimiter := range tokenDelimiters {
			if delimiter == string(r) {
				return true
			}
		}
		return false
	})
}

/*
	iTokenize

Adds lower-case tokens to the token list for insensitive matching
*/
func iTokenize(tokens []string) []string {
	for _, token := range tokens {
		if lt := strings.ToLower(token); lt != token && strings.EqualFold(token, lt) {
			tokens = append(tokens, lt)
		}
	}
	return tokens
}
