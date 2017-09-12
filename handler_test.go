package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"

	"github.com/stretchr/testify/assert"
)

func addExpectation(t *testing.T, exp Expectation) *bytes.Buffer {
	handlerAddExpectation := http.HandlerFunc(HandlerAddExpectation)

	expJSON, err := json.Marshal(exp)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", "/gozzmock/add_expectation", bytes.NewBuffer(expJSON))
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder := httptest.NewRecorder()
	handlerAddExpectation.ServeHTTP(httpTestResponseRecorder, req)
	assert.Equal(t, http.StatusOK, httpTestResponseRecorder.Code)

	return httpTestResponseRecorder.Body
}

func TestHandlerNoExpectations(t *testing.T) {
	handlerDefault := http.HandlerFunc(HandlerDefault)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from test server"))
	}))
	defer testServer.Close()

	// do request for response
	req, err := http.NewRequest("GET", "/request", nil)
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder := httptest.NewRecorder()
	handlerDefault.ServeHTTP(httpTestResponseRecorder, req)
	assert.Equal(t, http.StatusNotImplemented, httpTestResponseRecorder.Code)
	assert.Equal(t, "No expectations in gozzmock for request!", httpTestResponseRecorder.Body.String())
}

func TestHandlerAddAndRemoveExpectation(t *testing.T) {
	handlerRemoveExpectation := http.HandlerFunc(HandlerRemoveExpectation)
	expectedExp := Expectation{Key: "k"}
	expectedExps := Expectations{expectedExp.Key: expectedExp}

	body := addExpectation(t, expectedExp)
	expsjson, err := json.Marshal(expectedExps)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, string(expsjson), body.String())

	// remove expectation
	expRemoveJSON, err := json.Marshal(ExpectationRemove{Key: expectedExp.Key})
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", "/gozzmock/remove_expectation", bytes.NewBuffer(expRemoveJSON))
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder := httptest.NewRecorder()
	handlerRemoveExpectation.ServeHTTP(httpTestResponseRecorder, req)
	assert.Equal(t, http.StatusOK, httpTestResponseRecorder.Code)

	assert.Equal(t, "{}", httpTestResponseRecorder.Body.String())
}

func TestHandlerAddTwoExpectations(t *testing.T) {
	handlerDefault := http.HandlerFunc(HandlerDefault)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from test server"))
	}))
	defer testServer.Close()
	testServerURL, err := url.Parse(testServer.URL)
	if err != nil {
		panic(err)
	}

	addExpectation(t, Expectation{
		Key:      "response",
		Request:  &ExpectationRequest{Path: "/response"},
		Response: &ExpectationResponse{HTTPCode: http.StatusOK, Body: "response body"},
		Priority: 1})

	addExpectation(t, Expectation{
		Key:      "forward",
		Forward:  &ExpectationForward{Scheme: testServerURL.Scheme, Host: testServerURL.Host},
		Priority: 0})

	// do request for response
	req, err := http.NewRequest("POST", "/response", bytes.NewBuffer([]byte("request body")))
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder := httptest.NewRecorder()
	handlerDefault.ServeHTTP(httpTestResponseRecorder, req)
	assert.Equal(t, http.StatusOK, httpTestResponseRecorder.Code)

	assert.Equal(t, "response body", httpTestResponseRecorder.Body.String())

	// do request for forward
	req, err = http.NewRequest("POST", "/forward", bytes.NewBuffer([]byte("forward body")))
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder2 := httptest.NewRecorder()
	handlerDefault.ServeHTTP(httpTestResponseRecorder2, req)
	assert.Equal(t, http.StatusOK, httpTestResponseRecorder2.Code)

	assert.Equal(t, "response from test server", httpTestResponseRecorder2.Body.String())
}

func TestHandlerGetExpectations(t *testing.T) {
	handlerGetExpectations := http.HandlerFunc(HandlerGetExpectations)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from test server"))
	}))
	defer testServer.Close()

	expectation := Expectation{
		Key:      "response",
		Request:  &ExpectationRequest{Path: "/response"},
		Response: &ExpectationResponse{HTTPCode: http.StatusOK, Body: "response body"},
		Priority: 1}
	addExpectation(t, expectation)

	// do request for response
	req, err := http.NewRequest("GET", "/gozzmock/get_expectations", nil)
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder := httptest.NewRecorder()
	handlerGetExpectations.ServeHTTP(httpTestResponseRecorder, req)
	assert.Equal(t, http.StatusOK, httpTestResponseRecorder.Code)

	expectedResponse, err := json.Marshal(expectation)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, httpTestResponseRecorder.Body.String(), string(expectedResponse))
}

func TestHandlerStatus(t *testing.T) {
	handlerStatus := http.HandlerFunc(HandlerStatus)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from test server"))
	}))
	defer testServer.Close()

	// do request for response
	req, err := http.NewRequest("GET", "/gozzmock/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	httpTestResponseRecorder := httptest.NewRecorder()
	handlerStatus.ServeHTTP(httpTestResponseRecorder, req)
	assert.Equal(t, http.StatusOK, httpTestResponseRecorder.Code)
	assert.Contains(t, "gozzmock status is OK", httpTestResponseRecorder.Body.String())
}
