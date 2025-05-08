package pariksha

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type APITestCase struct {
	Name         string
	Method       string
	URL          string
	HandlerFunc  gin.HandlerFunc
	RequestBody  string
	PathParams   map[string]string
	Headers      map[string]string
	ExpectedCode int
	ContextKeys  map[string]any
	T            *testing.T
	B            *testing.B
	FunctionName string
}

// RunAPITest is a helper function to execute a series of API test cases.
// It iterates over a slice of APITestCase and runs each test case using the
// provided testing framework.
//   - tests: A slice of APITestCase, where each test case contains the logic and metadata for an individual API test.
//
// Each test case is executed in its own subtest, identified by the Name field
// of the APITestCase. The RunSingle method of each test case is invoked to
// perform the actual test logic.
func RunAPITest(tests []APITestCase) {
	for _, tc := range tests {
		tc.T.Run(tc.Name, func(t *testing.T) {
			tc.RunSingle()
		})
	}
}

// RunSingle executes a single API test case by simulating an HTTP request and validating the response against expected values.
func (tc APITestCase) RunSingle() {
	// Mark this as a helper function to clean up test error stack traces
	tc.T.Helper()
	recorder := tc.ExecuteHandler()
	if LogResponse {
		// Log the raw response body for debugging
		tc.T.Log("Raw Response Body:", recorder.Body.String())
	}
	var outer map[string]Resp
	// Parse the JSON response into a map with key "response"
	err := json.Unmarshal(recorder.Body.Bytes(), &outer)
	require.NoError(tc.T, err, "Could not unmarshal JSON response")
	// Extract the "response" object from the outer JSON map
	resp, ok := outer["response"]
	require.True(tc.T, ok, "Missing 'response' key in JSON")
	// Assert that the returned response code matches the expected code
	require.Equal(tc.T, tc.ExpectedCode, resp.Code)
}

// To sets up the Gin test context, builds the HTTP request from the APITestCase fields,
// executes the handler function, and returns the response recorder.
// This is used for testing or benchmarking HTTP handler behavior.
func (tc *APITestCase) ExecuteHandler() *httptest.ResponseRecorder {
	// Create a new HTTP request using the method, URL, and request body defined in the test case
	req, err := http.NewRequest(tc.Method, tc.URL, bytes.NewBufferString(tc.RequestBody))
	if err != nil {
		log.Println("Error while hitting the request..", err)
	}
	// Set any custom headers provided in the test case
	for key, value := range tc.Headers {
		req.Header.Set(key, value)
	}

	// Create a ResponseRecorder to capture the HTTP response
	recorder := httptest.NewRecorder()

	// Initialize a new Gin context using the test recorder
	ctx, _ := gin.CreateTestContext(recorder)

	// Assign the constructed request to the context
	ctx.Request = req

	// Attach any context-specific keys (e.g., for middleware) defined in the test case
	ctx.Keys = tc.ContextKeys

	// Set path parameters in the context, useful for dynamic routes like /user/:id
	ctx.Params = tc.SetAPITestPathParams()

	// Call the handler function with the prepared context
	tc.HandlerFunc(ctx)

	// Return the recorder so the caller can inspect the response
	return recorder
}

// If your Endpoint includes a path parameter, e.g., "/users/123",
// and your handler expects it as "/users/:id", you must set it manually.
func (tc APITestCase) SetAPITestPathParams() (params gin.Params) {
	if tc.PathParams != nil {
		for key, value := range tc.PathParams {
			params = append(params, gin.Param{Key: key, Value: value})
		}
	}
	return
}

// To executes a benchmark test for the API test case.
// It marks the function as a helper for better error reporting and runs
// the benchmark using the testing.B framework. The benchmark repeatedly
// calls the ExecuteHandler method for the number of iterations specified
// by the testing.B instance.
//
// This method is designed to be used in conjunction with the Go testing
// package's benchmarking tools.
func RunBenchmark(tc APITestCase) {
	tc.B.Helper() // Mark this as a helper function for better error reporting
	tc.B.Run(tc.Name, func(b *testing.B) {
		for i := 0; i < tc.B.N; i++ {
			tc.ExecuteHandler()
		}
	})
}

// RunProfiling executes the API test case with profiling enabled.
// This method is useful for analyzing the performance of the test case
// by collecting profiling data during its execution.
func RunProfiling(tc APITestCase) {
	// Mark this method as a test helper to improve test failure reports
	tc.B.Helper()

	// Skip profiling if no output formats are selected
	if len(ProfilingOutputFormats) == 0 {
		return
	}

	// Prepare and create the profile output directory
	profileDir := tc.GetProfileDir()
	MakeDirIfNotExists(profileDir)
	os.MkdirAll(profileDir, os.ModePerm)
	// require.NoError(tc.T, err, "Failed to create profile directory")

	// --- CPU PROFILE (Temporary file) ---
	cpuTempFile, err := os.CreateTemp("", "cpu_profile_*.prof")
	require.NoError(tc.B, err, "Failed to create temp CPU profile file")
	defer os.Remove(cpuTempFile.Name()) // Ensure the temporary file is deleted
	defer cpuTempFile.Close()

	// Start collecting CPU profile data
	err = pprof.StartCPUProfile(cpuTempFile)
	require.NoError(tc.B, err, "Failed to start CPU profile")
	// Enable additional runtime profiling
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	// Run the actual test case
	// Run the benchmark in parallel using goroutines
	tc.B.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tc.ExecuteHandler()
		}
	})

	// Stop CPU profiling after the test case execution
	pprof.StopCPUProfile()

	// Generate user-selected output formats for the CPU profile
	for _, format := range ProfilingOutputFormats {
		if format != "png" && format != "pdf" {
			tc.B.Logf("Invalid profiling output format: %s. Only 'png' and 'pdf' are supported.", format)
			return
		}
		outputPath := filepath.Join(profileDir, fmt.Sprintf("cpu.%s", format))
		GenerateGraph(cpuTempFile.Name(), outputPath, format)
	}

	// Loop through other enabled profiles (e.g., heap, goroutine)
	// Uncomment code in func EnabledProfilingTypes as per your profiling requirement
	for _, profile := range EnabledProfilingTypes {
		WriteProfileAndExport(profile, profileDir)
	}
}

// To make a new directory, if directory not exists in the path.
func MakeDirIfNotExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}
}

// GetProfileDir returns the directory path where profiling files should be stored
func (tc APITestCase) GetProfileDir() (profileDir string) {
	caseName := strings.ReplaceAll(tc.Name, " ", "_")
	functionName := filepath.Base(tc.FunctionName)
	profileDir = filepath.Join("profiles", functionName, caseName)
	return profileDir
}

// WriteProfileAndExport captures a runtime profile (like "heap", "goroutine", etc.)
// writes it to a temporary .prof file, and generates output files (e.g., PNG or PDF)
func WriteProfileAndExport(profileType, dir string) {
	// Return early if no formats selected
	if len(ProfilingOutputFormats) == 0 {
		log.Printf("Skipping %s profile: no output formats selected", profileType)
		return
	}
	// Create temp file for .prof
	temporaryFile, err := os.CreateTemp("", profileType+".prof")
	if err != nil {
		log.Printf("Failed to create temp file for %s profile: %v", profileType, err)
		return
	}
	defer os.Remove(temporaryFile.Name()) // Ensure .prof is deleted
	defer temporaryFile.Close()
	// Write profile data
	if err := pprof.Lookup(profileType).WriteTo(temporaryFile, 0); err != nil {
		log.Printf("Failed to write %s profile: %v", profileType, err)
		return
	}
	log.Printf("%s profile written to temporary file", profileType)
	// Export in selected formats
	for _, format := range ProfilingOutputFormats {
		outputPath := filepath.Join(dir, fmt.Sprintf("%s.%s", profileType, format))
		GenerateGraph(temporaryFile.Name(), outputPath, format)
	}
}

// GenerateGraph uses the Go pprof tool to generate a visualization of the profiling data.
// It takes the path to a .prof file, the desired output path, and the output format ("png" or "pdf").
func GenerateGraph(profilePath, outputPath, format string) {
	cmd := exec.Command("go", "tool", "pprof", fmt.Sprintf("-%s", format), profilePath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to generate %s for %s: %v", format, profilePath, err)
		return
	}
	if err = os.WriteFile(outputPath, output, 0644); err != nil {
		log.Printf("Failed to write %s to %s: %v", format, outputPath, err)
		return
	}
	log.Printf("%s saved to %s", strings.ToUpper(format), outputPath)
}

// GetBsonIdFromUUId converts a UUID string into a MongoDB BinData representation.
// The function removes dashes from the UUID string, decodes it into bytes,
// encodes the bytes into a Base64 string, and formats it as a BinData string.
//   - uuidStr: A string representing the UUID in standard format (e.g., "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx").
//   - binData: A string formatted as "BinData(0, '<Base64EncodedBytes>')", or an empty string if an error occurs during decoding.
func GetBsonIdFromUUId(uuidStr string) (binData string) {
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")
	uuidBytes, err := hex.DecodeString(uuidStr)
	if err != nil {
		return
	}
	base64Str := base64.StdEncoding.EncodeToString(uuidBytes)
	return fmt.Sprintf("BinData(0, '%s')", base64Str)
}

// To converts a MongoDB BinData representation back into a UUID string format.
// The function decodes the Base64-encoded bytes from the BinData string and formats them as a UUID string.
//   - binData: A string formatted as "BinData(0, '<Base64EncodedBytes>')".
//   - uuidStr: A string representing the UUID in standard format (e.g., "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"), or an empty string if an error occurs.
func GetStringUUidFromBsonId(binData string) (uuidStr string, err error) {
	// Extract the Base64-encoded bytes from the BinData string
	if !strings.HasPrefix(binData, "BinData(0, '") || !strings.HasSuffix(binData, "')") {
		err = fmt.Errorf("invalid BinData format")
		return
	}
	base64Str := strings.TrimSuffix(strings.TrimPrefix(binData, "BinData(0, '"), "')")

	// Decode the Base64 string into bytes
	uuidBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return
	}

	// Convert the bytes into a UUID string format
	if len(uuidBytes) != 16 {
		err = fmt.Errorf("invalid UUID byte length")
		return
	}
	uuidStr = fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuidBytes[0:4],
		uuidBytes[4:6],
		uuidBytes[6:8],
		uuidBytes[8:10],
		uuidBytes[10:16],
	)
	return
}
