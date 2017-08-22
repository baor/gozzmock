package handle

import (
	"bytes"
	"encoding/json"
	"gozzmock/model"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"

	"github.com/stretchr/testify/assert"
)

func addExpectation(t *testing.T, exp model.Expectation) *bytes.Buffer {
	handlerAddExpectation := http.HandlerFunc(AddExpectation)

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
func TestHandleAddAndRemoveExpectation(t *testing.T) {
	handlerRemoveExpectation := http.HandlerFunc(RemoveExpectation)
	expectedExp := model.Expectation{Key: "k"}
	expectedExps := model.Expectations{expectedExp.Key: expectedExp}

	body := addExpectation(t, expectedExp)
	expsjson, err := json.Marshal(expectedExps)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, string(expsjson), body.String())

	// remove expectation
	expRemoveJSON, err := json.Marshal(model.ExpectationRemove{Key: expectedExp.Key})
	log.Println(string(expRemoveJSON))
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

func TestHandleAddTwoExpectations(t *testing.T) {
	handlerDefault := http.HandlerFunc(Default)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("response from test server"))
	}))
	defer testServer.Close()
	testServerURL, err := url.Parse(testServer.URL)
	if err != nil {
		panic(err)
	}

	addExpectation(t, model.Expectation{
		Key:      "response",
		Request:  model.ExpectationRequest{Path: "/response"},
		Response: model.ExpectationResponse{HTTPCode: http.StatusOK, Body: "response body"},
		Priority: 1})

	addExpectation(t, model.Expectation{
		Key:      "forward",
		Forward:  model.ExpectationForward{Scheme: testServerURL.Scheme, Host: testServerURL.Host},
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
