package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestControllerGetExpectations_NoExpectations_ReturnEmptyList(t *testing.T) {
	var exps = ControllerGetExpectations(Expectations{})
	assert.Empty(t, exps)
}

func TestControllerAddExpectations_NoExpectations_ReturnOneItem(t *testing.T) {
	var exp = Expectation{Key: "k"}

	var exps = ControllerAddExpectation(exp.Key, exp, Expectations{})
	assert.Contains(t, exps, exp.Key)
	assert.Equal(t, exp, exps[exp.Key])
}

func TestControllerAddExpectations_ExistingKey_ReturnUpdatedOneItem(t *testing.T) {
	var exp1 = Expectation{Key: "k", Delay: 1}
	var exp2 = Expectation{Key: "k", Delay: 2}

	var exps = ControllerAddExpectation(exp1.Key, exp1, Expectations{})
	assert.Contains(t, exps, exp1.Key)

	exps = ControllerAddExpectation(exp2.Key, exp2, exps)
	assert.Contains(t, exps, exp2.Key)
	assert.Equal(t, 1, len(exps))
	assert.Equal(t, exp2, exps[exp2.Key])
}

func TestControllerAddExpectations_NewKey_ReturnTwoItems(t *testing.T) {
	var exp1 = Expectation{Key: "k1", Delay: 1}
	var exp2 = Expectation{Key: "k2", Delay: 2}

	var exps = ControllerAddExpectation(exp1.Key, exp1, Expectations{})
	assert.Contains(t, exps, exp1.Key)

	exps = ControllerAddExpectation(exp2.Key, exp2, exps)
	assert.Contains(t, exps, exp2.Key)

	assert.Equal(t, 2, len(exps))
	assert.Equal(t, exp1, exps[exp1.Key])
	assert.Equal(t, exp2, exps[exp2.Key])
}

func TestControllerRemoveExpectations_OneExpectations_ReturnEmptyList(t *testing.T) {
	var exp = Expectation{Key: "k"}

	var exps = ControllerAddExpectation(exp.Key, exp, Expectations{})
	assert.Contains(t, exps, exp.Key)

	exps = ControllerRemoveExpectation(exp.Key, exps)
	assert.Empty(t, exps)
}

func TestControllerRemoveWrongKeyExpectations_OneExpectations_NotReturnError(t *testing.T) {
	var exp = Expectation{Key: "k"}

	var exps = ControllerAddExpectation(exp.Key, exp, Expectations{})
	assert.Contains(t, exps, exp.Key)

	exps = ControllerRemoveExpectation("wrong_key", exps)
	assert.Contains(t, exps, exp.Key)
}

func TestControllerTranslateRequestToExpectation_SimpleRequest_AllFieldsTranslated(t *testing.T) {
	request, err := http.NewRequest("POST", "https://www.host.com/path", strings.NewReader("body text"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("h1", "hv1")

	exp := ControllerTranslateRequestToExpectation(request)

	assert.NotNil(t, exp)
	assert.Equal(t, "POST", exp.Method)
	assert.Equal(t, "/path", exp.Path)
	assert.Equal(t, "body text", exp.Body)
	assert.Equal(t, 1, len(exp.Headers))
	assert.Equal(t, "hv1", exp.Headers["h1"])
}

func TestControllerTranslateHTTPHeadersToExpHeaders_TwoHeaders_HeadersTranslated(t *testing.T) {
	header := http.Header{}
	header.Add("h1", "hv1")
	header.Add("h1", "hv2")

	expHeaders := ControllerTranslateHTTPHeadersToExpHeaders(header)

	assert.Equal(t, 1, len(expHeaders))
	assert.Equal(t, "hv1,hv2", expHeaders["h1"])
}

func TestControllerStringPassesFilter_EmptyFilter_True(t *testing.T) {
	assert.True(t, ControllerStringPassesFilter("abc", ""))
}

func TestControllerStringPassesFilter_ExistingSubstring_True(t *testing.T) {
	assert.True(t, ControllerStringPassesFilter("abc", "ab"))
}

func TestControllerStringPassesFilter_ExistingRegex_True(t *testing.T) {
	assert.True(t, ControllerStringPassesFilter("abc", ".b."))
}

func TestControllerStringPassesFilter_NotExistingSubstring_False(t *testing.T) {
	assert.False(t, ControllerStringPassesFilter("abc", "zz"))
}

func TestControllerStringPassesFilter_NotExistingRegex_False(t *testing.T) {
	assert.False(t, ControllerStringPassesFilter("abc", ".z."))
}

func TestControllerRequestPassFilter_EmptyRequestEmptyFilter_True(t *testing.T) {
	assert.True(t, ControllerRequestPassFilter(
		&ExpectationRequest{},
		&ExpectationRequest{}))
}

func TestControllerRequestPassFilter_MethodsAreEq_True(t *testing.T) {
	assert.True(t, ControllerRequestPassFilter(
		&ExpectationRequest{Method: "POST"},
		&ExpectationRequest{Method: "POST"}))
}

func TestControllerRequestPassFilter_PathsAreEq_True(t *testing.T) {
	assert.True(t, ControllerRequestPassFilter(
		&ExpectationRequest{Path: "/path"},
		&ExpectationRequest{Path: "/path"}))
}

func TestControllerRequestPassFilter_MethodsNotEqAndPathsAreEq_False(t *testing.T) {
	assert.False(t, ControllerRequestPassFilter(
		&ExpectationRequest{Method: "GET", Path: "/path"},
		&ExpectationRequest{Method: "POST", Path: "/path"}))
}

func TestControllerRequestPassFilter_HeadersAreEq_True(t *testing.T) {
	assert.True(t, ControllerRequestPassFilter(
		&ExpectationRequest{Headers: Headers{"h1": "hv1"}},
		&ExpectationRequest{Headers: Headers{"h1": "hv1"}}))
}

func TestControllerRequestPassFilter_HeaderNotEq_False(t *testing.T) {
	result := ControllerRequestPassFilter(
		&ExpectationRequest{Headers: Headers{"h1": "hv1"}},
		&ExpectationRequest{Headers: Headers{"h2": "hv2"}})
	assert.False(t, result)
}

func TestControllerRequestPassFilter_HeaderValueNotEq_False(t *testing.T) {
	assert.False(t, ControllerRequestPassFilter(
		&ExpectationRequest{Headers: Headers{"h1": "hv1"}},
		&ExpectationRequest{Headers: Headers{"h1": "hv2"}}))
}

func TestControllerRequestPassFilter_NoHeaderinReq_False(t *testing.T) {
	assert.False(t, ControllerRequestPassFilter(
		&ExpectationRequest{},
		&ExpectationRequest{Headers: Headers{"h2": "hv2"}}))
}

func TestControllerRequestPassFilter_NoHeaderInFilter_True(t *testing.T) {
	assert.True(t, ControllerRequestPassFilter(
		&ExpectationRequest{Headers: Headers{"h1": "hv1"}},
		&ExpectationRequest{}))
}

func TestControllerRequestPassFilter_BodysEq_True(t *testing.T) {
	assert.True(t, ControllerRequestPassFilter(
		&ExpectationRequest{Body: "body"},
		&ExpectationRequest{Body: "body"}))
}

func TestControllerSortExpectationsByPriority_EmptyExps_OK(t *testing.T) {
	sortedMap := ControllerSortExpectationsByPriority(Expectations{})
	assert.Equal(t, 0, len(sortedMap))
}

func TestControllerSortExpectationsByPriority_ListOfExpectations_OK(t *testing.T) {
	exp1 := Expectation{Key: "k1", Priority: 1}
	exp2 := Expectation{Key: "k0", Priority: 0}
	exp3 := Expectation{Key: "k2", Priority: 2}
	exps := ControllerAddExpectation(exp1.Key, exp1, Expectations{})
	exps = ControllerAddExpectation(exp2.Key, exp2, exps)
	exps = ControllerAddExpectation(exp3.Key, exp3, exps)
	sortedMap := ControllerSortExpectationsByPriority(exps)
	assert.Equal(t, 3, len(sortedMap))
	assert.Equal(t, "k2", sortedMap[0].Key)
	assert.Equal(t, "k1", sortedMap[1].Key)
	assert.Equal(t, "k0", sortedMap[2].Key)
}
