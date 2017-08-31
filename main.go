package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var initSetup string
	flag.StringVar(&initSetup, "init", "{}", "initial setup")
	flag.Parse()
	fmt.Println("initSetup:", initSetup)
	fmt.Println("tail:", flag.Args())

	exps := Expectations{}
	err := json.Unmarshal([]byte(initSetup), &exps)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for key, exp := range exps {
		ControllerAddExpectation(key, exp, nil)
	}

	http.HandleFunc("/gozzmock/add_expectation", HandlerAddExpectation)
	http.HandleFunc("/gozzmock/remove_expectation", HandlerRemoveExpectation)
	http.HandleFunc("/gozzmock/get_expectations", HandlerGetExpectations)
	http.HandleFunc("/gozzmock/status", HandlerStatus)
	http.HandleFunc("/", HandlerDefault)
	http.ListenAndServe(":8080", nil)
}
