package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpectationsFromString(t *testing.T) {
	str := "{\"k\":{\"key\": \"k\"}}"
	exps := ExpectationsFromString(str)
	assert.Equal(t, "k", exps["k"].Key)
}

func TestConvertationExpectationFromReadCloser(t *testing.T) {
	str := "{\"key\": \"k\"}"
	exp := ExpectationFromReadCloser(ioutil.NopCloser(strings.NewReader(str)))
	assert.Equal(t, "k", exp.Key)
}
