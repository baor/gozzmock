package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var expectations Expectations

var mu sync.Mutex

// ControllerGetExpectations returns list with expectations in concurrent mode
func ControllerGetExpectations(expsInjection Expectations) Expectations {
	if expsInjection != nil {
		return expsInjection
	}

	mu.Lock()
	defer mu.Unlock()

	if expectations == nil {
		expectations = make(Expectations)
	}
	return expectations
}

// ControllerAddExpectation adds new expectation to list. If expectation with same key exists, updates it
func ControllerAddExpectation(key string, exp Expectation, expsInjection Expectations) Expectations {
	var exps = ControllerGetExpectations(expsInjection)
	mu.Lock()
	defer mu.Unlock()

	exps[key] = exp
	return exps
}

// ControllerRemoveExpectation removes expectation with particular key
func ControllerRemoveExpectation(key string, expsInjection Expectations) Expectations {
	var exps = ControllerGetExpectations(expsInjection)
	mu.Lock()
	defer mu.Unlock()

	if _, ok := exps[key]; ok {
		delete(exps, key)
	}
	return exps
}

// ControllerTranslateHTTPHeadersToExpHeaders translates http headers into custom headers map
func ControllerTranslateHTTPHeadersToExpHeaders(httpHeader http.Header) *Headers {
	headers := Headers{}
	for name, headerLine := range httpHeader {
		headers[name] = strings.Join(headerLine, ",")
	}
	return &headers
}

// ControllerTranslateRequestToExpectation Translates http request to expectation request
func ControllerTranslateRequestToExpectation(r *http.Request) ExpectationRequest {
	var expRequest = ExpectationRequest{}
	expRequest.Method = r.Method
	expRequest.Path = r.URL.Path

	// Buffer the body
	if r.Body != nil {
		bodyBuffer, error := ioutil.ReadAll(r.Body)
		if error == nil {
			expRequest.Body = string(bodyBuffer)
		}
	}

	if len(r.Header) > 0 {
		expRequest.Headers = ControllerTranslateHTTPHeadersToExpHeaders(r.Header)
	}

	return expRequest
}

// ControllerStringPassesFilter validates whether the input string has filter string as substring or as a regex
func ControllerStringPassesFilter(str string, filter string) bool {
	r, error := regexp.Compile(filter)
	if error != nil {
		return strings.Contains(str, filter)
	}
	return r.Match([]byte(str))
}

// ControllerRequestPassesFilter validates whether the incoming request passes particular filter
func ControllerRequestPassesFilter(req *ExpectationRequest, storedExpectation *ExpectationRequest) bool {
	if storedExpectation == nil {
		log.Printf("Stored expectation.request is nil")
		return true
	}

	if len(storedExpectation.Method) > 0 && storedExpectation.Method != req.Method {
		log.Printf("method %s should be %s", req.Method, storedExpectation.Method)
		return false
	}

	if len(storedExpectation.Path) > 0 && !ControllerStringPassesFilter(req.Path, storedExpectation.Path) {
		log.Printf("path %s doesn't pass filter %s", req.Path, storedExpectation.Path)
		return false
	}

	if len(storedExpectation.Body) > 0 && !ControllerStringPassesFilter(req.Body, storedExpectation.Body) {
		log.Printf("body %s doesn't pass filter %s", req.Body, storedExpectation.Body)
		return false
	}

	if storedExpectation.Headers != nil {
		if req.Headers == nil {
			log.Printf("Request is expected to contain headers")
			return false
		}
		for storedHeaderName, storedHeaderValue := range *storedExpectation.Headers {
			value, ok := (*req.Headers)[storedHeaderName]
			if !ok {
				log.Printf("No header %s in the request headers %v", storedHeaderName, req.Headers)
				return false
			}
			if !ControllerStringPassesFilter(value, storedHeaderValue) {
				log.Printf("header %s:%s has been rejected. Expected header value %s", storedHeaderName, value, storedHeaderValue)
				return false
			}
		}
	}

	return true
}

// ControllerSortExpectationsByPriority returns map with int keys sorted by priority DESC.
// 0-indexed element has the highest priority
func ControllerSortExpectationsByPriority(exps Expectations) ExpectationsInt {
	listForSorting := ExpectationsInt{}
	i := 0
	for _, exp := range exps {
		listForSorting[i] = exp
		i++
	}
	sort.Sort(listForSorting)
	return listForSorting
}

// ControllerCreateHTTPRequest creates an http request based on incoming request and forward rules
func ControllerCreateHTTPRequest(req ExpectationRequest, fwd *ExpectationForward) *http.Request {
	fwdURL, err := url.Parse(fmt.Sprintf("%s://%s%s", fwd.Scheme, fwd.Host, req.Path))
	if err != nil {
		panic(err)
	}
	log.Println("Send request to " + fwdURL.String())
	httpReq, err := http.NewRequest(req.Method, fwdURL.String(), bytes.NewBuffer([]byte(req.Body)))
	if err != nil {
		panic(err)
	}

	if req.Headers != nil {
		for name, value := range *req.Headers {
			httpReq.Header.Set(name, value)
		}
	}

	if fwd.Headers != nil {
		for name, value := range *fwd.Headers {
			httpReq.Header.Set(name, value)
		}
	}

	return httpReq
}
