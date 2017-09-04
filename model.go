package main

import (
	"encoding/json"
	"io"
	"strings"
	"time"
)

// Headers are HTTP headers
type Headers map[string]string

// ExpectationRequest is filter for incoming requests
type ExpectationRequest struct {
	Method  string  `json:"method"`
	Path    string  `json:"path"`
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`
}

// ExpectationForward is forward action if request passes filter
type ExpectationForward struct {
	Scheme  string  `json:"scheme"`
	Host    string  `json:"host"`
	Headers Headers `json:"headers"`
}

// ExpectationResponse is response action if request passes filter
type ExpectationResponse struct {
	HTTPCode int     `json:"httpcode"`
	Body     string  `json:"body"`
	Headers  Headers `json:"headers"`
}

// Expectation is single set of rules: expected request and prepared action
type Expectation struct {
	Key      string              `json:"key"`
	Request  ExpectationRequest  `json:"request"`
	Forward  ExpectationForward  `json:"forward"`
	Response ExpectationResponse `json:"response"`
	Delay    time.Duration       `json:"delay"`
	Priority int                 `json:"priority"`
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
	exp := Expectation{}
	bodyDecoder := json.NewDecoder(readCloser)
	defer readCloser.Close()
	err := bodyDecoder.Decode(&exp)
	if err != nil {
		panic(err)
	}
	return exp
}

// ExpectationsFromString decodes string to expectaions
func ExpectationsFromString(str string) Expectations {
	exps := Expectations{}
	bodyDecoder := json.NewDecoder(strings.NewReader(str))
	err := bodyDecoder.Decode(&exps)
	if err != nil {
		panic(err)
	}
	return exps
}
