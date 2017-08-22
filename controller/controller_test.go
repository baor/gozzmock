package controller

import (
	"gozzmock/model"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExpectations_NoExpectations_ReturnEmptyList(t *testing.T) {
	var exps = GetExpectations(model.Expectations{})
	assert.Empty(t, exps)
}

func TestAddExpectations_NoExpectations_ReturnOneItem(t *testing.T) {
	var exp = model.Expectation{Key: "k"}

	var exps = AddExpectation(exp.Key, exp, model.Expectations{})
	assert.Contains(t, exps, exp.Key)
	assert.Equal(t, exp, exps[exp.Key])
}

func TestAddExpectations_ExistingKey_ReturnUpdatedOneItem(t *testing.T) {
	var exp1 = model.Expectation{Key: "k", Delay: 1}
	var exp2 = model.Expectation{Key: "k", Delay: 2}

	var exps = AddExpectation(exp1.Key, exp1, model.Expectations{})
	assert.Contains(t, exps, exp1.Key)

	exps = AddExpectation(exp2.Key, exp2, exps)
	assert.Contains(t, exps, exp2.Key)
	assert.Equal(t, 1, len(exps))
	assert.Equal(t, exp2, exps[exp2.Key])
}

func TestAddExpectations_NewKey_ReturnTwoItems(t *testing.T) {
	var exp1 = model.Expectation{Key: "k1", Delay: 1}
	var exp2 = model.Expectation{Key: "k2", Delay: 2}

	var exps = AddExpectation(exp1.Key, exp1, model.Expectations{})
	assert.Contains(t, exps, exp1.Key)

	exps = AddExpectation(exp2.Key, exp2, exps)
	assert.Contains(t, exps, exp2.Key)

	assert.Equal(t, 2, len(exps))
	assert.Equal(t, exp1, exps[exp1.Key])
	assert.Equal(t, exp2, exps[exp2.Key])
}

func TestRemoveExpectations_OneExpectations_ReturnEmptyList(t *testing.T) {
	var exp = model.Expectation{Key: "k"}

	var exps = AddExpectation(exp.Key, exp, model.Expectations{})
	assert.Contains(t, exps, exp.Key)

	exps = RemoveExpectation(exp.Key, exps)
	assert.Empty(t, exps)
}

func TestRemoveWrongKeyExpectations_OneExpectations_NotReturnError(t *testing.T) {
	var exp = model.Expectation{Key: "k"}

	var exps = AddExpectation(exp.Key, exp, model.Expectations{})
	assert.Contains(t, exps, exp.Key)

	exps = RemoveExpectation("wrong_key", exps)
	assert.Contains(t, exps, exp.Key)
}

func TestTranslateRequestToExpectation_SimpleRequest_AllFieldsTranslated(t *testing.T) {
	request, err := http.NewRequest("POST", "https://www.host.com/path", strings.NewReader("body text"))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Add("h1", "hv1")

	exp := TranslateRequestToExpectation(request)

	assert.NotNil(t, exp)
	assert.Equal(t, "POST", exp.Method)
	assert.Equal(t, "/path", exp.Path)
	assert.Equal(t, "body text", exp.Body)
	assert.Equal(t, 1, len(exp.Headers))
	assert.Equal(t, "hv1", exp.Headers["h1"])
}

func TestTranslateHTTPHeadersToExpHeaders_TwoHeaders_HeadersTranslated(t *testing.T) {
	header := http.Header{}
	header.Add("h1", "hv1")
	header.Add("h1", "hv2")

	expHeaders := TranslateHTTPHeadersToExpHeaders(header)

	assert.Equal(t, 1, len(expHeaders))
	assert.Equal(t, "hv1,hv2", expHeaders["h1"])
}

func TestStringPassesFilter_EmptyFilter_True(t *testing.T) {
	assert.True(t, StringPassesFilter("abc", ""))
}

func TestStringPassesFilter_ExistingSubstring_True(t *testing.T) {
	assert.True(t, StringPassesFilter("abc", "ab"))
}

func TestStringPassesFilter_ExistingRegex_True(t *testing.T) {
	assert.True(t, StringPassesFilter("abc", ".b."))
}

func TestStringPassesFilter_NotExistingSubstring_False(t *testing.T) {
	assert.False(t, StringPassesFilter("abc", "zz"))
}

func TestStringPassesFilter_NotExistingRegex_False(t *testing.T) {
	assert.False(t, StringPassesFilter("abc", ".z."))
}

func TestRequestPassFilter_EmptyRequestEmptyFilter_True(t *testing.T) {
	assert.True(t, RequestPassFilter(
		&model.ExpectationRequest{},
		&model.ExpectationRequest{}))
}

func TestRequestPassFilter_MethodsAreEq_True(t *testing.T) {
	assert.True(t, RequestPassFilter(
		&model.ExpectationRequest{Method: "POST"},
		&model.ExpectationRequest{Method: "POST"}))
}

func TestRequestPassFilter_PathsAreEq_True(t *testing.T) {
	assert.True(t, RequestPassFilter(
		&model.ExpectationRequest{Path: "/path"},
		&model.ExpectationRequest{Path: "/path"}))
}

func TestRequestPassFilter_MethodsNotEqAndPathsAreEq_False(t *testing.T) {
	assert.False(t, RequestPassFilter(
		&model.ExpectationRequest{Method: "GET", Path: "/path"},
		&model.ExpectationRequest{Method: "POST", Path: "/path"}))
}

func TestRequestPassFilter_HeadersAreEq_True(t *testing.T) {
	assert.True(t, RequestPassFilter(
		&model.ExpectationRequest{Headers: model.Headers{"h1": "hv1"}},
		&model.ExpectationRequest{Headers: model.Headers{"h1": "hv1"}}))
}

func TestRequestPassFilter_HeaderNotEq_False(t *testing.T) {
	result := RequestPassFilter(
		&model.ExpectationRequest{Headers: model.Headers{"h1": "hv1"}},
		&model.ExpectationRequest{Headers: model.Headers{"h2": "hv2"}})
	assert.False(t, result)
}

func TestRequestPassFilter_HeaderValueNotEq_False(t *testing.T) {
	assert.False(t, RequestPassFilter(
		&model.ExpectationRequest{Headers: model.Headers{"h1": "hv1"}},
		&model.ExpectationRequest{Headers: model.Headers{"h1": "hv2"}}))
}

func TestRequestPassFilter_NoHeaderinReq_False(t *testing.T) {
	assert.False(t, RequestPassFilter(
		&model.ExpectationRequest{},
		&model.ExpectationRequest{Headers: model.Headers{"h2": "hv2"}}))
}

func TestRequestPassFilter_NoHeaderInFilter_True(t *testing.T) {
	assert.True(t, RequestPassFilter(
		&model.ExpectationRequest{Headers: model.Headers{"h1": "hv1"}},
		&model.ExpectationRequest{}))
}

func TestRequestPassFilter_BodysEq_True(t *testing.T) {
	assert.True(t, RequestPassFilter(
		&model.ExpectationRequest{Body: "body"},
		&model.ExpectationRequest{Body: "body"}))
}

func TestSortExpectationsByPriority_EmptyExps_OK(t *testing.T) {
	sortedMap := SortExpectationsByPriority(model.Expectations{})
	assert.Equal(t, 0, len(sortedMap))
}

func TestSortExpectationsByPriority_ListOfExpectations_OK(t *testing.T) {
	exp1 := model.Expectation{Key: "k1", Priority: 1}
	exp2 := model.Expectation{Key: "k0", Priority: 0}
	exp3 := model.Expectation{Key: "k2", Priority: 2}
	exps := AddExpectation(exp1.Key, exp1, model.Expectations{})
	exps = AddExpectation(exp2.Key, exp2, exps)
	exps = AddExpectation(exp3.Key, exp3, exps)
	sortedMap := SortExpectationsByPriority(exps)
	assert.Equal(t, 3, len(sortedMap))
	assert.Equal(t, "k2", sortedMap[0].Key)
	assert.Equal(t, "k1", sortedMap[1].Key)
	assert.Equal(t, "k0", sortedMap[2].Key)
}
