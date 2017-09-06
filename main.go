package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func httpHandleFuncWithLogs(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Println(fmt.Sprintf("%v", req))

		handler(w, r)
	}

	http.HandleFunc(pattern, wrappedHandler)
}

func main() {
	var initSetup string
	flag.StringVar(&initSetup, "init", "{}", "initial setup")
	flag.Parse()
	fmt.Println("initSetup:", initSetup)
	fmt.Println("tail:", flag.Args())

	exps := ExpectationsFromString(initSetup)

	for _, exp := range exps {
		ControllerAddExpectation(exp.Key, exp, nil)
	}

	http.HandleFunc("/gozzmock/status", HandlerStatus)
	httpHandleFuncWithLogs("/gozzmock/add_expectation", HandlerAddExpectation)
	httpHandleFuncWithLogs("/gozzmock/remove_expectation", HandlerRemoveExpectation)
	httpHandleFuncWithLogs("/gozzmock/get_expectations", HandlerGetExpectations)
	httpHandleFuncWithLogs("/", HandlerDefault)
	http.ListenAndServe(":8080", nil)
}
