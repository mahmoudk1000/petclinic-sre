package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestResult represents the standard output from a PowerShell script
type TestResult struct {
	TestName        string  `json:"test_name"`
	Status          int     `json:"status"`
	DurationSeconds float64 `json:"duration_seconds"`
}

var (
	testStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "petclinic_test_status",
			Help: "Status of the test (1 = success, 0 = failure)",
		},
		[]string{"test_name"},
	)
	testDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "petclinic_test_duration_seconds",
			Help: "Duration of the test in seconds",
		},
		[]string{"test_name"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(testStatus)
	prometheus.MustRegister(testDuration)
}

// getEnv returns the value of the environment variable or a default value if it is not set
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getTickerInterval returns the ticker interval from the environment variable or a default value if it is not set
func getTickerInterval(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf(
			"Invalid ticker interval value %s, using default %f minutes",
			valueStr,
			defaultValue.Minutes(),
		)
		return defaultValue
	}
	return time.Duration(value) * time.Minute
}

// runTest runs a PowerShell script and returns the parsed result
func runTest(scriptPath string) (*TestResult, error) {
	cmd := exec.Command("pwsh", scriptPath)
	output, err := cmd.CombinedOutput() // Capture both stdout and stderr
	if err != nil {
		log.Printf("Error running PowerShell script: %s", err)
		log.Printf("Script output: %s", output) // Log the script output for debugging
		return nil, err
	}

	var result TestResult
	if err := json.Unmarshal(output, &result); err != nil {
		log.Printf("Error unmarshaling JSON: %s", err)
		log.Printf("Script output: %s", output) // Log the script output for debugging
		return nil, err
	}

	return &result, nil
}

// performTests runs all PowerShell scripts in the Tests directory and updates the metrics
func performTests(testsDir, baseURL string) {
	// Find all PowerShell scripts in the Tests subdirectory
	var scripts []string
	err := filepath.Walk(testsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ps1") {
			scripts = append(scripts, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking through the Tests directory: %v", err)
	}

	// Run tests and update metrics
	for _, script := range scripts {
		// Set environment variables for each test
		os.Setenv("APP_URL", baseURL)

		result, err := runTest(script)
		if err != nil {
			log.Printf("Error running test %s: %v", script, err)
			continue
		}

		// Update metrics based on the test result
		testStatus.WithLabelValues(result.TestName).Set(float64(result.Status))
		testDuration.WithLabelValues(result.TestName).Set(result.DurationSeconds)

		// Log the updated metrics
		log.Printf(
			"Updated metrics for test: %s, status: %d, duration: %f",
			result.TestName,
			result.Status,
			result.DurationSeconds,
		)
	}
}

func main() {
	// Get the port from the environment or use default
	port := getEnv("PORT", "9091")

	// Define environment variables for tests with default fallback values
	baseURL := getEnv("APP_URL", "http://localhost:8080")

	// Define the Tests subdirectory
	testsDir := "Tests"

	// Get the ticker interval from the environment or use default (5 minutes)
	tickerInterval := getTickerInterval("TICKER_INTERVAL", 5*time.Minute)

	// Create a ticker to run tests at the specified interval
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	// Run tests immediately at startup
	go performTests(testsDir, baseURL)

	// Start a goroutine to run the tests at regular intervals
	go func() {
		for range ticker.C {
			performTests(testsDir, baseURL)
		}
	}()

	// Expose metrics via HTTP
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
