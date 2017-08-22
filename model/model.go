package model

import (
	"time"
)

type Headers map[string]string

type ExpectationRequest struct {
	Method  string  `json:"method"`
	Path    string  `json:"path"`
	Body    string  `json:"body"`
	Headers Headers `json:"headers"`
}

type ExpectationForward struct {
	Scheme  string  `json:"scheme"`
	Host    string  `json:"host"`
	Headers Headers `json:"headers"`
}

type ExpectationResponse struct {
	HTTPCode int     `json:"httpcode"`
	Body     string  `json:"body"`
	Headers  Headers `json:"headers"`
}

type Expectation struct {
	Key      string              `json:"key"`
	Request  ExpectationRequest  `json:"request"`
	Forward  ExpectationForward  `json:"forward"`
	Response ExpectationResponse `json:"response"`
	Delay    time.Duration       `json:"delay"`
	Priority int                 `json:"priority"`
}

type ExpectationRemove struct {
	Key string `json:"key"`
}

type Expectations map[string]Expectation

// ExpectationsInt is for sorting expectations by priority. the lowest priority is 0
type ExpectationsInt map[int]Expectation

func (exps ExpectationsInt) Len() int           { return len(exps) }
func (exps ExpectationsInt) Swap(i, j int)      { exps[i], exps[j] = exps[j], exps[i] }
func (exps ExpectationsInt) Less(i, j int) bool { return exps[i].Priority > exps[j].Priority }
