package main

import (
	"flag"
	"fmt"

	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func httpHandleFuncWithLogs(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	fLog := log.With().Str("function", "httpHandleFuncWithLogs").Logger()

	wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
		req, err := httputil.DumpRequest(r, true)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		fLog.Debug().Msg(string(req))

		handler(w, r)
	}

	http.HandleFunc(pattern, wrappedHandler)
}

func setZeroLogLevel(logLevel string) {

	selectedLevel := zerolog.DebugLevel

	switch logLevel {
	case "debug":
		selectedLevel = zerolog.DebugLevel
	case "info":
		selectedLevel = zerolog.InfoLevel
	case "warn":
		selectedLevel = zerolog.WarnLevel
	case "error":
		selectedLevel = zerolog.ErrorLevel
	case "fatal":
		selectedLevel = zerolog.FatalLevel
	case "panic":
		selectedLevel = zerolog.PanicLevel
	}
	fmt.Println("set log level:", selectedLevel)
	zerolog.SetGlobalLevel(selectedLevel)
}

func main() {
	var initExpectations string
	flag.StringVar(&initExpectations, "expectations", "[]", "set initial expectations")
	var logLevel string
	flag.StringVar(&logLevel, "loglevel", "debug", "set log level: debug, info, warn, error, fatal, panic")
	flag.Parse()

	fmt.Println("initial expectations:", initExpectations)
	fmt.Println("loglevel:", logLevel)
	fmt.Println("tail:", flag.Args())

	setZeroLogLevel(logLevel)

	exps := ExpectationsFromString(initExpectations)

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
