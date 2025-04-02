package weather

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock API response structure
func mockWeatherResponse(temp float64, desc string) string {
	return fmt.Sprintf(`{
		"main": { "temp": %.1f },
		"weather": [{ "description": "%s" }]
	}`, temp, desc)
}

// Test for successful weather processing
func TestWeatherProcessor_Success(t *testing.T) {
	// Mock server for OpenWeather API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "q=London") // Ensure query is passed correctly
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, mockWeatherResponse(20.5, "clear sky"))
	}))
	defer mockServer.Close()

	// Replace real API URL with mock server URL
	processor := &WeatherProcessor{
		apiKey: "test-key",
		client: mockServer.Client(),
	}

	// Mock task payload
	payload := `{"location": "London"}`
	result, err := processor.ProcessTask([]byte(payload))

	assert.NoError(t, err)
	assert.Equal(t, "Weather for London: 20.5°C, clear sky", result)
}

// Test missing location in payload
func TestWeatherProcessor_MissingLocation(t *testing.T) {
	processor := NewWeatherProcessor("test-key")

	payload := `{"location": ""}`
	_, err := processor.ProcessTask([]byte(payload))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "location is required")
}

// Test API failure (e.g., city not found)
func TestWeatherProcessor_APIFailure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"message": "city not found"}`)
	}))
	defer mockServer.Close()

	processor := &WeatherProcessor{
		apiKey: "test-key",
		client: mockServer.Client(),
	}

	payload := `{"location": "UnknownCity"}`
	_, err := processor.ProcessTask([]byte(payload))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "city not found")
}

// Test handling of malformed API response
func TestWeatherProcessor_MalformedResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"main": { "temp": "INVALID" }}`) // Invalid temp format
	}))
	defer mockServer.Close()

	processor := &WeatherProcessor{
		apiKey: "test-key",
		client: mockServer.Client(),
	}

	payload := `{"location": "London"}`
	_, err := processor.ProcessTask([]byte(payload))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse weather API response")
}

// Test handling of empty weather description
func TestWeatherProcessor_EmptyWeatherArray(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"main": { "temp": 22.5 }, "weather": []}`) // Empty weather array
	}))
	defer mockServer.Close()

	processor := &WeatherProcessor{
		apiKey: "test-key",
		client: mockServer.Client(),
	}

	payload := `{"location": "Tokyo"}`
	result, err := processor.ProcessTask([]byte(payload))

	assert.NoError(t, err)
	assert.Equal(t, "Weather for Tokyo: 22.5°C, No weather description available", result)
}

