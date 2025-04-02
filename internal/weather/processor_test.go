package weather

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var openWeatherAPIUrl string = "https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&appid=%s&units=metric"

func TestWeatherProcessor_ProcessTask(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name           string
		input          []byte
		mockResponse   string
		mockStatus     int
		expectedOutput string
		expectedError  string
	}{
		{
			name: "Invalid Input - Missing Coordinates",
			input: []byte(`{
				"city": "Chicago"
			}`),
			expectedError: "latitude and longitude are required",
		},
		{
			name: "API Error - Unauthorized",
			input: []byte(`{
				"lat": 33.44,
				"lon": -94.04
			}`),
			mockStatus:    http.StatusUnauthorized,
			expectedError: "weather API returned status 401",
		},
		{
			name:          "Invalid JSON Input",
			input:         []byte(`{invalid json}`),
			expectedError: "invalid request format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != "" {
					_, _ = w.Write([]byte(tt.mockResponse))
				}
			}))
			defer server.Close()

			// Create processor with test server URL and mock API key
			processor := NewWeatherProcessor("test-api-key")
			processor.client = server.Client() // Use test server's client

			// Replace the API URL with our test server URL
			oldURL := openWeatherAPIUrl
			openWeatherAPIUrl = server.URL + "?lat=%f&lon=%f&appid=%s&units=metric"
			defer func() { openWeatherAPIUrl = oldURL }()

			// Execute test
			output, err := processor.ProcessTask(tt.input)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestValidateWeatherRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   WeatherTaskPayload
		wantErr bool
	}{
		{
			name: "Valid Request",
			input: WeatherTaskPayload{
				Lat:  40.71,
				Long: -74.01,
			},
			wantErr: false,
		},
		{
			name: "Missing Latitude",
			input: WeatherTaskPayload{
				Long: -74.01,
			},
			wantErr: true,
		},
		{
			name: "Missing Longitude",
			input: WeatherTaskPayload{
				Lat: 40.71,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWeatherRequest(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
