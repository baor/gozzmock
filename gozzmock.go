package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gozzmock/controller"
	"gozzmock/handler"
	"gozzmock/model"
	"net/http"
	"os"
)

func main() {
	var initSetup string
	flag.StringVar(&initSetup, "init", "{}", "initial setup")
	flag.Parse()
	fmt.Println("initSetup:", initSetup)
	fmt.Println("tail:", flag.Args())

	exps := model.Expectations{}
	err := json.Unmarshal([]byte(initSetup), &exps)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for key, exp := range exps {
		controller.AddExpectation(key, exp, nil)
	}

	http.HandleFunc("/gozzmock/add_expectation", handler.AddExpectation)
	http.HandleFunc("/gozzmock/remove_expectation", handler.RemoveExpectation)
	http.HandleFunc("/gozzmock/get_expectations", handler.GetExpectations)
	http.HandleFunc("/gozzmock/status", handler.Status)
	http.HandleFunc("/", handler.Default)
	http.ListenAndServe(":8080", nil)
}
