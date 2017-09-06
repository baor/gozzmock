package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpectationsFromString(t *testing.T) {
	str := "[{\"key\": \"k1\"},{\"key\": \"k2\"}]"
	exps := ExpectationsFromString(str)
	assert.Equal(t, 2, len(exps))
	assert.Equal(t, "k1", exps[0].Key)
	assert.Equal(t, "k2", exps[1].Key)
}

func TestConvertationExpectationFromReadCloser(t *testing.T) {
	str := "{\"key\": \"k\"}"
	exp := ExpectationFromReadCloser(ioutil.NopCloser(strings.NewReader(str)))
	assert.Equal(t, "k", exp.Key)
}
