package pariksha_test

import (
	"net/http"
	"testing"

	pariksha "github.com/DevNavix/pariksha"

	"github.com/gin-gonic/gin"
)

func ExampleRunAPITest() {
	var t *testing.T
	tests := []pariksha.APITestCase{
		{
			Name:         "Example Test Case",
			Method:       http.MethodGet,
			URL:          "/example",
			HandlerFunc:  func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "success"}) },
			ExpectedCode: http.StatusOK,
			T:            t,
		},
	}
	// Run the test cases using the Helper framework
	pariksha.RunAPITest(tests)
}

func ExampleRunBenchmark() {
	var b *testing.B
	// Add a true test case which would hit api successfully...
	testCase := pariksha.APITestCase{
		Name:        "BenchmarkExample",
		Method:      http.MethodGet,
		URL:         "/example",
		HandlerFunc: func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "success"}) },
		B:           b,
	}
	// Use func RunBenchmark to perform benchmarking, and Profiling if want to perform profiling.
	pariksha.RunBenchmark(testCase)
}

func ExampleRunProfiling() {
	var b *testing.B
	// Add a true test case which would hit api successfully...
	testCase := pariksha.APITestCase{
		Name:        "BenchmarkExample",
		Method:      http.MethodGet,
		URL:         "/example",
		HandlerFunc: func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "success"}) },
		B:           b,
	}
	// Use func RunBenchmark to perform benchmarking, and Profiling if want to perform profiling.
	pariksha.RunProfiling(testCase)
}
