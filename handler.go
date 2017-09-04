package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// HandlerAddExpectation handler parses request and adds expectation to global expectations list
func HandlerAddExpectation(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	if r.Method != "POST" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}
	log.Println("Body = " + fmt.Sprint(r.Body))

	exp := Expectation{}
	bodyDecoder := json.NewDecoder(r.Body)
	err := bodyDecoder.Decode(&exp)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var exps = ControllerAddExpectation(exp.Key, exp, nil)

	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	w.Write(expsjson)
}

// HandlerRemoveExpectation handler parses request and deletes expectation from global expectations list
func HandlerRemoveExpectation(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	if r.Method != "POST" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}
	log.Println("Body = " + fmt.Sprint(r.Body))

	requestBody := ExpectationRemove{}
	bodyDecoder := json.NewDecoder(r.Body)
	err := bodyDecoder.Decode(&requestBody)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var exps = ControllerRemoveExpectation(requestBody.Key, nil)
	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	w.Write(expsjson)
}

// HandlerGetExpectations handler parses request and returns global expectations list
func HandlerGetExpectations(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
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
	log.Println(r)

	generateResponseToResponseWriter(&w, ControllerTranslateRequestToExpectation(r))
}

func uploadResponseToResponseWriter(w *http.ResponseWriter, resp *ExpectationResponse) {
	(*w).WriteHeader(resp.HTTPCode)
	(*w).Write([]byte(resp.Body))
	for name, value := range resp.Headers {
		(*w).Header().Set(name, value)
	}
}

func generateResponseToResponseWriter(w *http.ResponseWriter, req ExpectationRequest) {
	exps := ControllerGetExpectations(nil)
	for _, exp := range ControllerSortExpectationsByPriority(exps) {
		if ControllerRequestPassFilter(&req, &exp.Request) {
			expResponseIsEmpty := (exp.Response.HTTPCode == 0 && exp.Response.Body == "")
			if !expResponseIsEmpty {
				uploadResponseToResponseWriter(w, &exp.Response)
				time.Sleep(time.Second * exp.Delay)
				return
			}
			expForwardIsEmpty := (exp.Forward.Scheme == "" && exp.Forward.Host == "")
			if !expForwardIsEmpty {
				httpReq := ControllerCreateHTTPRequest(req, exp.Forward)
				doHTTPRequest(w, httpReq)
				time.Sleep(time.Second * exp.Delay)
				return
			}
		}
	}
	(*w).WriteHeader(http.StatusNotImplemented)
	(*w).Write([]byte("No expectations in gozzmock for request!"))
}

func doHTTPRequest(w *http.ResponseWriter, httpReq *http.Request) {
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

	(*w).WriteHeader(resp.StatusCode)
	(*w).Write(body)
	headers := ControllerTranslateHTTPHeadersToExpHeaders(resp.Header)
	for name, value := range headers {
		(*w).Header().Set(name, value)
	}
}
