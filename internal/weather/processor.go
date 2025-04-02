package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)


// WeatherTaskPayload defines the structure for the weather task payload
type WeatherTaskPayload struct {
	Lat          float64 `json:"lat" binding:"required" example:"33.44"`
	Long          float64 `json:"lon" binding:"required" example:"-94.04"`
	City         string  `json:"city" example:"Chicago"`
	Units        string  `json:"units" example:"metric"` // "metric" or "imperial"
	ForecastDays int     `json:"forecast_days" example:"3"` // Number of days to forecast (0 = current only)
	Exclude      string  `json:"exclude" example:"minutely,hourly"` // Optional: parts to exclude
	Language     string  `json:"language" example:"en"` // Response language
}

// WeatherResponse defines the complete API response structure
type WeatherResponse struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	TimezoneOffset int     `json:"timezone_offset"`
	Current        struct {
		Dt         int64     `json:"dt"`
		Sunrise    int64     `json:"sunrise"`
		Sunset     int64     `json:"sunset"`
		Temp       float64   `json:"temp"`
		FeelsLike  float64   `json:"feels_like"`
		Pressure   int       `json:"pressure"`
		Humidity   int       `json:"humidity"`
		DewPoint   float64   `json:"dew_point"`
		Uvi        float64   `json:"uvi"`
		Clouds     int       `json:"clouds"`
		Visibility int       `json:"visibility"`
		WindSpeed  float64   `json:"wind_speed"`
		WindDeg    int       `json:"wind_deg"`
		WindGust   float64   `json:"wind_gust"`
		Weather    []Weather `json:"weather"`
	} `json:"current"`
	Minutely []struct {
		Dt          int64   `json:"dt"`
		Precipitation float64 `json:"precipitation"`
	} `json:"minutely"`
	Hourly []struct {
		Dt         int64     `json:"dt"`
		Temp       float64   `json:"temp"`
		FeelsLike  float64   `json:"feels_like"`
		Pressure   int       `json:"pressure"`
		Humidity   int       `json:"humidity"`
		DewPoint   float64   `json:"dew_point"`
		Uvi        float64   `json:"uvi"`
		Clouds     int       `json:"clouds"`
		Visibility int       `json:"visibility"`
		WindSpeed  float64   `json:"wind_speed"`
		WindDeg    int       `json:"wind_deg"`
		WindGust   float64   `json:"wind_gust"`
		Weather    []Weather `json:"weather"`
		Pop        float64   `json:"pop"`
	} `json:"hourly"`
	Daily []struct {
		Dt        int64     `json:"dt"`
		Sunrise   int64     `json:"sunrise"`
		Sunset    int64     `json:"sunset"`
		Moonrise  int64     `json:"moonrise"`
		Moonset   int64     `json:"moonset"`
		MoonPhase float64   `json:"moon_phase"`
		Summary   string    `json:"summary"`
		Temp      struct {
			Day   float64 `json:"day"`
			Min   float64 `json:"min"`
			Max   float64 `json:"max"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"temp"`
		FeelsLike struct {
			Day   float64 `json:"day"`
			Night float64 `json:"night"`
			Eve   float64 `json:"eve"`
			Morn  float64 `json:"morn"`
		} `json:"feels_like"`
		Pressure   int       `json:"pressure"`
		Humidity   int       `json:"humidity"`
		DewPoint   float64   `json:"dew_point"`
		WindSpeed  float64   `json:"wind_speed"`
		WindDeg    int       `json:"wind_deg"`
		WindGust   float64   `json:"wind_gust"`
		Weather    []Weather `json:"weather"`
		Clouds     int       `json:"clouds"`
		Pop        float64   `json:"pop"`
		Rain       float64   `json:"rain"`
		Uvi        float64   `json:"uvi"`
	} `json:"daily"`
	Alerts []struct {
		SenderName  string   `json:"sender_name"`
		Event       string   `json:"event"`
		Start       int64    `json:"start"`
		End         int64    `json:"end"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	} `json:"alerts"`
}

// Weather defines the common weather condition fields
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// WeatherProcessor handles weather API requests
type WeatherProcessor struct {
	apiKey string
	client *http.Client
}

// NewWeatherProcessor creates a new WeatherProcessor
func NewWeatherProcessor(apiKey string) *WeatherProcessor {
	return &WeatherProcessor{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ProcessTask handles weather data processing
func (p *WeatherProcessor) ProcessTask(body []byte) (string, error) {
	var req WeatherTaskPayload
	if err := json.Unmarshal(body, &req); err != nil {
		return "", fmt.Errorf("invalid request format: %w", err)
	}

	if req.Lat == 0 || req.Long == 0 {
		return "", fmt.Errorf("latitude and longitude are required")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&appid=%s&units=metric", 
		req.Lat, req.Long, p.apiKey)

	resp, err := p.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return "", fmt.Errorf("failed to decode weather response: %w", err)
	}

	return p.formatResponse(weatherResp, req.City), nil
}

// formatResponse creates a human-readable weather summary
func (p *WeatherProcessor) formatResponse(resp WeatherResponse, city string) string {
	if len(resp.Current.Weather) == 0 {
		return "No weather data available"
	}

	summary := fmt.Sprintf("Weather for %s (%.2f,%.2f):\n", city, resp.Lat, resp.Lon)
	summary += fmt.Sprintf("- Current: %.1f°C, %s\n", resp.Current.Temp, resp.Current.Weather[0].Description)
	summary += fmt.Sprintf("- Feels like: %.1f°C\n", resp.Current.FeelsLike)
	summary += fmt.Sprintf("- Humidity: %d%%\n", resp.Current.Humidity)
	summary += fmt.Sprintf("- Wind: %.1f m/s, %d°\n", resp.Current.WindSpeed, resp.Current.WindDeg)

	if len(resp.Alerts) > 0 {
		summary += "\nAlerts:\n"
		for _, alert := range resp.Alerts {
			summary += fmt.Sprintf("- %s: %s\n", alert.Event, alert.Description)
		}
	}

	return summary
}
