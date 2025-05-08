package pariksha

var (

	// Add common contexts required in api testing.
	CommonContext = map[string]any{
		"user_name": "Navjot Sharma",
	}
	// Add common headers required in api test cases.
	CommonHeaders = map[string]string{
		"Content-Type": "application/json",
	}
	// To print api response we should enable this flag
	LogResponse = false
	// Formats in which profiling results will be saved.
	// Options: "png", "pdf". Add both to enable both outputs.
	ProfilingOutputFormats = []string{"png"} //ex: []string{"png"} or []string{"pdf"} or []string{"png", "pdf"}
)

// Change `EnabledProfilingTypes` below to enable or disable specific profiling types.
// Available options:
// - "heap"          - "goroutine"      - "block"        - "mutex"        - "threadcreate"
// Example:
// To enable only heap and goroutine profiling, change it to:
//
//	var EnabledProfilingTypes = []string{"heap", "goroutine"}
//
// CPU profiling (cpu.prof)is always enabled by default, Uncomment code .
var EnabledProfilingTypes = []string{
	"heap",         // memory allocations
	"goroutine",    // number of goroutines
	"block",        // blocking profile
	"mutex",        // mutex contention
	"threadcreate", // thread creation
}

type Resp struct {
	Code      int    `json:"code"`
	ApiStatus int    `json:"api_status"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	MetaData  any    `json:"meta_data,omitempty"`
}
