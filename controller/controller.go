package controller

import (
	"bytes"
	"fmt"
	"gozzmock/model"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var expectations model.Expectations

var mu sync.Mutex

// GetExpectations returns list with expectations in concurrent mode
func GetExpectations(expsInjection model.Expectations) model.Expectations {
	if expsInjection != nil {
		return expsInjection
	}

	mu.Lock()
	defer mu.Unlock()

	if expectations == nil {
		expectations = make(model.Expectations)
	}
	return expectations
}

// AddExpectation adds new expectation to list. If expectation with same key exists, updates it
func AddExpectation(key string, exp model.Expectation, expsInjection model.Expectations) model.Expectations {
	var exps = GetExpectations(expsInjection)
	mu.Lock()
	defer mu.Unlock()

	exps[key] = exp
	return exps
}

// RemoveExpectation removes expectation with particular key
func RemoveExpectation(key string, expsInjection model.Expectations) model.Expectations {
	var exps = GetExpectations(expsInjection)
	mu.Lock()
	defer mu.Unlock()

	if _, ok := exps[key]; ok {
		delete(exps, key)
	}
	return exps
}

// TranslateHTTPHeadersToExpHeaders translates http headers into custom headers map
func TranslateHTTPHeadersToExpHeaders(httpHeader http.Header) model.Headers {
	headers := model.Headers{}
	for name, headerLine := range httpHeader {
		name = strings.ToLower(name)
		headers[name] = strings.Join(headerLine, ",")
	}
	return headers
}

// TranslateRequestToExpectation Translates http request to expectation request
func TranslateRequestToExpectation(r *http.Request) model.ExpectationRequest {
	var expRequest = model.ExpectationRequest{}
	expRequest.Method = r.Method
	expRequest.Path = r.URL.Path

	// Buffer the body
	bodyBuffer, error := ioutil.ReadAll(r.Body)
	if error == nil {
		expRequest.Body = string(bodyBuffer)
	}

	if len(r.Header) > 0 {
		expRequest.Headers = TranslateHTTPHeadersToExpHeaders(r.Header)
	}

	return expRequest
}

// StringPassesFilter validates whether the input string has filter string as substring or as a regex
func StringPassesFilter(str string, filter string) bool {
	r, error := regexp.Compile(filter)
	if error != nil {
		return strings.Contains(str, filter)
	}
	return r.Match([]byte(str))
}

// RequestPassFilter validates whether the incoming request passesparticular filter
func RequestPassFilter(req *model.ExpectationRequest, filter *model.ExpectationRequest) bool {
	if len(filter.Method) > 0 && filter.Method != req.Method {
		log.Printf("method %s should be %s", req.Method, filter.Method)
		return false
	}

	if len(filter.Path) > 0 && !StringPassesFilter(req.Path, filter.Path) {
		log.Printf("path %s doesn't pass filter %s", req.Path, filter.Path)
		return false
	}

	if len(filter.Body) > 0 && !StringPassesFilter(req.Body, filter.Body) {
		log.Printf("body %s doesn't pass filter %s", req.Body, filter.Body)
		return false
	}

	if len(filter.Headers) > 0 {
		for fhName, fhValue := range filter.Headers {
			value, ok := req.Headers[fhName]
			if !ok {
				log.Printf("header %s isn't present in the request headers %v", fhName, req.Headers)
				return false
			}
			if !StringPassesFilter(value, fhValue) {
				log.Printf("header %s:%s doesnt' pass filter for value %s", fhName, value, fhValue)
				return false
			}
		}
	}

	return true
}

// SortExpectationsByPriority returns map with int keys sorted by priority DESC.
// 0-indexed element has the highest priority
func SortExpectationsByPriority(exps model.Expectations) model.ExpectationsInt {
	listForSorting := model.ExpectationsInt{}
	i := 0
	for _, exp := range exps {
		listForSorting[i] = exp
		i++
	}
	sort.Sort(listForSorting)
	return listForSorting
}

// CreateHTTPRequest creates an http request based on incoming request and forward rules
func CreateHTTPRequest(req model.ExpectationRequest, fwd model.ExpectationForward) *http.Request {
	fwdURL, err := url.Parse(fmt.Sprintf("%s://%s%s", fwd.Scheme, fwd.Host, req.Path))
	if err != nil {
		panic(err)
	}

	httpReq, err := http.NewRequest(req.Method, fwdURL.String(), bytes.NewBuffer([]byte(req.Body)))
	if err != nil {
		panic(err)
	}

	for name, value := range req.Headers {
		httpReq.Header.Set(name, value)
	}

	for name, value := range fwd.Headers {
		httpReq.Header.Set(name, value)
	}

	return httpReq
}
