package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/rs/zerolog/log"
)

// HandlerAddExpectation handler parses request and adds expectation to global expectations list
func HandlerAddExpectation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}

	exp := ExpectationFromReadCloser(r.Body)

	var exps = ControllerAddExpectation(exp.Key, exp, nil)

	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	w.Write(expsjson)
}

// HandlerRemoveExpectation handler parses request and deletes expectation from global expectations list
func HandlerRemoveExpectation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}
	defer r.Body.Close()

	requestBody := ExpectationRemove{}
	bodyDecoder := json.NewDecoder(r.Body)
	err := bodyDecoder.Decode(&requestBody)
	if err != nil {
		panic(err)
	}

	var exps = ControllerRemoveExpectation(requestBody.Key, nil)
	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	w.Write(expsjson)
}

// HandlerGetExpectations handler parses request and returns global expectations list
func HandlerGetExpectations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}

	var exps = ControllerGetExpectations(nil)
	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(expsjson))
}

// HandlerStatus handler returns applications status
func HandlerStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "gozzmock status is OK")
}

// HandlerDefault handler is an entry point for all incoming requests
func HandlerDefault(w http.ResponseWriter, r *http.Request) {
	req, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	log.Print("Request: " + string(req))

	generateResponseToResponseWriter(&w, ControllerTranslateRequestToExpectation(r))
}

func uploadResponseToResponseWriter(w *http.ResponseWriter, resp *ExpectationResponse) {
	(*w).WriteHeader(resp.HTTPCode)
	(*w).Write([]byte(resp.Body))
	if resp.Headers != nil {
		for name, value := range *resp.Headers {
			(*w).Header().Set(name, value)
		}
	}
}

func generateResponseToResponseWriter(w *http.ResponseWriter, req ExpectationRequest) {
	storedExpectations := ControllerGetExpectations(nil)
	orderedStoredExpectations := ControllerSortExpectationsByPriority(storedExpectations)
	for i := 0; i < len(orderedStoredExpectations); i++ {
		exp := orderedStoredExpectations[i]

		if !ControllerRequestPassesFilter(&req, exp.Request) {
			continue
		}

		time.Sleep(time.Second * exp.Delay)

		if exp.Response != nil {
			log.Print("Apply response expectation")
			uploadResponseToResponseWriter(w, exp.Response)
			return
		}

		if exp.Forward != nil {
			log.Print("Apply forward expectation")
			httpReq := ControllerCreateHTTPRequest(req, exp.Forward)
			doHTTPRequest(w, httpReq)
			return
		}
	}

	(*w).WriteHeader(http.StatusNotImplemented)
	(*w).Write([]byte("No expectations in gozzmock for request!"))
}

func doHTTPRequest(w *http.ResponseWriter, httpReq *http.Request) {
	httpClient := &http.Client{}

	// disable gzip compression
	httpReq.Header.Set("Accept-Encoding", "deflate")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Printf("Response body: %s", body)

	(*w).WriteHeader(resp.StatusCode)
	(*w).Write(body)

	headers := *ControllerTranslateHTTPHeadersToExpHeaders(resp.Header)
	for name, value := range headers {
		(*w).Header().Set(name, value)
	}
}
