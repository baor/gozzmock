package handler

import (
	"encoding/json"
	"fmt"
	"gozzmock/controller"
	"gozzmock/model"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// AddExpectation handler parses request and adds expectation to global expectations list
func AddExpectation(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	if r.Method != "POST" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}
	log.Println("Body = " + fmt.Sprint(r.Body))

	exp := model.Expectation{}
	bodyDecoder := json.NewDecoder(r.Body)
	err := bodyDecoder.Decode(&exp)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var exps = controller.AddExpectation(exp.Key, exp, nil)

	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	w.Write(expsjson)
}

// RemoveExpectation handler parses request and deletes expectation from global expectations list
func RemoveExpectation(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	if r.Method != "POST" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}
	log.Println("Body = " + fmt.Sprint(r.Body))

	requestBody := model.ExpectationRemove{}
	bodyDecoder := json.NewDecoder(r.Body)
	err := bodyDecoder.Decode(&requestBody)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var exps = controller.RemoveExpectation(requestBody.Key, nil)
	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	w.Write(expsjson)
}

// GetExpectations handler parses request and returns global expectations list
func GetExpectations(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	if r.Method != "GET" {
		panic(fmt.Sprintf("Wrong method %s", r.Method))
	}

	var exps = controller.GetExpectations(nil)
	expsjson, err := json.Marshal(exps)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(expsjson))
}

// Status handler returns applications status
func Status(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
	fmt.Fprint(w, "gozzmock status is OK")
}

// Default handler is an entry point for all incoming requests
func Default(w http.ResponseWriter, r *http.Request) {
	log.Println(r)

	generateResponseToResponseWriter(&w, controller.TranslateRequestToExpectation(r))
}

func uploadResponseToResponseWriter(w *http.ResponseWriter, resp *model.ExpectationResponse) {
	(*w).WriteHeader(resp.HTTPCode)
	(*w).Write([]byte(resp.Body))
	for name, value := range resp.Headers {
		(*w).Header().Set(name, value)
	}
}

func generateResponseToResponseWriter(w *http.ResponseWriter, req model.ExpectationRequest) {
	exps := controller.GetExpectations(nil)
	for _, exp := range controller.SortExpectationsByPriority(exps) {
		if controller.RequestPassFilter(&req, &exp.Request) {
			expResponseIsEmpty := (exp.Response.HTTPCode == 0 && exp.Response.Body == "")
			if !expResponseIsEmpty {
				uploadResponseToResponseWriter(w, &exp.Response)
				time.Sleep(time.Second * exp.Delay)
				return
			}
			expForwardIsEmpty := (exp.Forward.Scheme == "" && exp.Forward.Host == "")
			if !expForwardIsEmpty {
				httpReq := controller.CreateHTTPRequest(req, exp.Forward)
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
	headers := controller.TranslateHTTPHeadersToExpHeaders(resp.Header)
	for name, value := range headers {
		(*w).Header().Set(name, value)
	}
}
