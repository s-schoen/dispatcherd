package handler_test

import (
	"dispatcherd/handler"
	"dispatcherd/test"
	"net/http"
	"testing"
)

func TestHealthy(t *testing.T) {
	runner := test.NewTestRunner(handler.HandleHealth)
	res := runner.Run(t).ExpectNoError().ExpectStatusCode(http.StatusOK)
	test.AssertSingleAPIResponse(res, "OK")
}
