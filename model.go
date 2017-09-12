package main

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Headers are HTTP headers
type Headers map[string]string

// ExpectationRequest is filter for incoming requests
type ExpectationRequest struct {
	Method  string   `json:"method"`
	Path    string   `json:"path"`
	Body    string   `json:"body"`
	Headers *Headers `json:"headers,omitempty"`
}

// ExpectationForward is forward action if request passes filter
type ExpectationForward struct {
	Scheme  string   `json:"scheme"`
	Host    string   `json:"host"`
	Headers *Headers `json:"headers,omitempty"`
}

// ExpectationResponse is response action if request passes filter
type ExpectationResponse struct {
	HTTPCode int      `json:"httpcode"`
	Body     string   `json:"body"`
	Headers  *Headers `json:"headers,omitempty"`
}

// Expectation is single set of rules: expected request and prepared action
type Expectation struct {
	Key      string               `json:"key"`
	Request  *ExpectationRequest  `json:"request,omitempty"`
	Forward  *ExpectationForward  `json:"forward,omitempty"`
	Response *ExpectationResponse `json:"response,omitempty"`
	Delay    time.Duration        `json:"delay,omitempty"`
	Priority int                  `json:"priority,omitempty"`
}

// ExpectationRemove removes action from list by key
type ExpectationRemove struct {
	Key string `json:"key"`
}

// Expectations is a map for expectations
type Expectations map[string]Expectation

// ExpectationsInt is for sorting expectations by priority. the lowest priority is 0
type ExpectationsInt map[int]Expectation

func (exps ExpectationsInt) Len() int           { return len(exps) }
func (exps ExpectationsInt) Swap(i, j int)      { exps[i], exps[j] = exps[j], exps[i] }
func (exps ExpectationsInt) Less(i, j int) bool { return exps[i].Priority > exps[j].Priority }

// ExpectationFromReadCloser decodes readCloser to expectaion
func ExpectationFromReadCloser(readCloser io.ReadCloser) Expectation {
	fLog := log.With().Str("function", "ExpectationFromReadCloser").Logger()

	exp := Expectation{}
	bodyDecoder := json.NewDecoder(readCloser)
	defer readCloser.Close()
	err := bodyDecoder.Decode(&exp)
	if err != nil {
		fLog.Panic().Err(err)
		return exp
	}
	expectationSetDefaultValues(&exp)
	return exp
}

// ExpectationsFromString decodes string with array of expectations to array of expectaion objects
func ExpectationsFromString(str string) []Expectation {
	fLog := log.With().Str("function", "ExpectationsFromString").Logger()

	exps := make([]Expectation, 0)

	bodyDecoder := json.NewDecoder(strings.NewReader(str))
	err := bodyDecoder.Decode(&exps)
	if err != nil {
		fLog.Panic().Err(err)
		return exps
	}
	for _, exp := range exps {
		expectationSetDefaultValues(&exp)
	}
	return exps
}

// expectationSetDefaultValues sets default values after deserialization
func expectationSetDefaultValues(exp *Expectation) {
	if exp.Forward != nil && exp.Forward.Scheme == "" {
		exp.Forward.Scheme = "http"
	}
}
