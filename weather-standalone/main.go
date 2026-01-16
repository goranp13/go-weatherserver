//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Windows API functions to hide console and set icon
var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = kernel32.NewProc("ShowWindow")
	user32               = syscall.NewLazyDLL("user32.dll")
	procLoadIcon         = user32.NewProc("LoadIconA")
	procSetClassLong     = user32.NewProc("SetClassLongA")
)

const (
	SW_HIDE = 0
	SW_SHOW = 5
)

func hideConsole() {
	if runtime.GOOS == "windows" {
		hwnd, _, _ := procGetConsoleWindow.Call()
		if hwnd != 0 {
			procShowWindow.Call(hwnd, uintptr(SW_HIDE))
		}
	}
}

// Configuration
const (
	CacheRefreshInterval = 5 * time.Minute
	APITimeout           = 10 * time.Second
	MaxRequests          = 100 // Rate limiting: requests per minute
)

// RequestTracker for rate limiting
type RequestTracker struct {
	Count     int
	LastReset time.Time
	Mutex     sync.Mutex
}

var requestTracker = &RequestTracker{
	LastReset: time.Now(),
}

// CachedWeatherData stores weather data and when it was last updated
type CachedWeatherData struct {
	Data      WeatherData
	Timestamp time.Time
	Mutex     sync.RWMutex
}

// HistoricalData stores weather history for trends
type HistoricalData struct {
	Location     string
	Temperatures []int
	Times        []time.Time
	Mutex        sync.RWMutex
}

// Global cache for weather data - updates only every 5 minutes
var weatherCache = make(map[string]*CachedWeatherData)
var cacheLock sync.RWMutex

// Historical data for each city
var historicalData = make(map[string]*HistoricalData)
var historyLock sync.RWMutex

// WeatherData represents current weather information
type WeatherData struct {
	Location        string
	Temperature     int
	Condition       string
	Emoji           string
	Description     string
	DramaticMessage string
	WindSpeed       int
	Humidity        int
	FeelsLike       int
	UVIndex         float64
	PrecipChance    int
}

// ForecastDay represents a day's forecast
type ForecastDay struct {
	Date      string
	High      int
	Low       int
	Emoji     string
	Condition string
}

// WeatherForecast contains current weather and forecast
type WeatherForecast struct {
	Current  WeatherData
	Forecast []ForecastDay
	AsciiArt string
}

// CityCoordinates stores latitude and longitude for a city
type CityCoordinates struct {
	Name      string
	Latitude  float64
	Longitude float64
	Emoji     string
}

// OpenMeteo API response structures
type OpenMeteoResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		Humidity    int     `json:"relative_humidity_2m"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		WeatherCode int     `json:"weather_code"`
		Time        string  `json:"time"`
	} `json:"current"`
	Daily struct {
		Time           []string  `json:"time"`
		WeatherCode    []int     `json:"weather_code"`
		TemperatureMax []float64 `json:"temperature_2m_max"`
		TemperatureMin []float64 `json:"temperature_2m_min"`
	} `json:"daily"`
}

// City coordinates for Croatian cities
var cityCoordinates = map[string]CityCoordinates{
	"zagreb": {
		Name:      "Zagreb 🏛️",
		Latitude:  45.815,
		Longitude: 15.9819,
		Emoji:     "🏛️",
	},
	"split": {
		Name:      "Split 🏖️",
		Latitude:  43.5081,
		Longitude: 16.4402,
		Emoji:     "🏖️",
	},
	"dubrovnik": {
		Name:      "Dubrovnik ⛱️",
		Latitude:  42.6412,
		Longitude: 18.1084,
		Emoji:     "⛱️",
	},
	"rijeka": {
		Name:      "Rijeka 🌊",
		Latitude:  45.3271,
		Longitude: 14.4205,
		Emoji:     "🌊",
	},
	"zadar": {
		Name:      "Zadar 🐚",
		Latitude:  43.1312,
		Longitude: 15.2313,
		Emoji:     "🐚",
	},
	"osijek": {
		Name:      "Osijek 🌾",
		Latitude:  45.5544,
		Longitude: 18.6955,
		Emoji:     "🌾",
	},
}

// Mock data
var locations = map[string]WeatherData{
	"zagreb": {
		Location:    "Zagreb 🏛️",
		Temperature: 3,
		Condition:   "Oblačno",
		Emoji:       "☁️",
		WindSpeed:   12,
		Humidity:    75,
		FeelsLike:   0,
	},
	"split": {
		Location:    "Split 🏖️",
		Temperature: 11,
		Condition:   "Sunčano",
		Emoji:       "☀️",
		WindSpeed:   8,
		Humidity:    65,
		FeelsLike:   10,
	},
	"dubrovnik": {
		Location:    "Dubrovnik ⛱️",
		Temperature: 13,
		Condition:   "Sunčano",
		Emoji:       "☀️",
		WindSpeed:   5,
		Humidity:    60,
		FeelsLike:   12,
	},
	"rijeka": {
		Location:    "Rijeka 🌊",
		Temperature: 5,
		Condition:   "Kišno",
		Emoji:       "🌧️",
		WindSpeed:   18,
		Humidity:    88,
		FeelsLike:   2,
	},
	"zadar": {
		Location:    "Zadar 🐚",
		Temperature: 10,
		Condition:   "Djelomično oblačno",
		Emoji:       "⛅",
		WindSpeed:   10,
		Humidity:    70,
		FeelsLike:   8,
	},
	"osijek": {
		Location:    "Osijek 🌾",
		Temperature: 7,
		Condition:   "Oblačno",
		Emoji:       "☁️",
		WindSpeed:   10,
		Humidity:    72,
		FeelsLike:   5,
	},
}

var dramaticMessages = map[string][]string{
	"Kišno": {
		"Kiša pada - Donesi kišobran!",
		"Mokri ulazak - Čuva se od kiše!",
		"Nebo se prazni - Ostani unutar!",
		"Kiša je ovdje - Bodljikavo vrijeme!",
	},
	"Sunčano": {
		"Sunce sjaji - Divno vrijeme!",
		"Zaštita od sunca preporučena!",
		"Najljepši dan godine!",
		"Idealno za planinu!",
	},
	"Oblačno": {
		"Oblaci pokrivaju nebo!",
		"Blago sive boje - ali ugodno!",
		"Nema sunca ali nije loše!",
		"Tipično zimsko vrijeme!",
	},
	"Djelomično oblačno": {
		"Mješavina sunca i oblaka!",
		"Lijepo, ali može biti hladnije!",
		"Promjenjivo vrijeme!",
		"Oblaci se pojavljuju i nestaju!",
	},
	"Snježno": {
		"Snijeg pada - Zimska čarolija!",
		"Bijela pokrivka na zemlji!",
		"Zimski podaci - Odjevite se toplo!",
		"Snježni pejzaž je spektakularan!",
	},
}

var asciiArts = map[string]string{
	"Kišno": `
    ___
   (____)
   /    \
   | ~~ |
    \ ~~/
     |~~|
    /|  |\
   / |  | \
  `,
	"Sunčano": `
      \  |  /
       \ | /
        \|/
    --- (*) ---
        /|\
       / | \
      /  |  \
  `,
	"Snježno": `
     *  *  *
    *  ❄️  *
     *  *  *
    **  *  **
  *    *    *
    *  *  *
  *  *  *
  `,
	"Oblačno": `
    (    )
     ( )
    _____
   |     |
  `,
	"Djelomično oblačno": `
      \  |  /
       \ | /
        \|/
    --- (*) ---
    (    )
     ( )
  `,
}

func init() {
	rand.Seed(time.Now().UnixNano())
	// Initialize historical data for all cities
	cities := []string{"zagreb", "split", "dubrovnik", "rijeka", "zadar", "osijek"}
	for _, city := range cities {
		historicalData[city] = &HistoricalData{
			Location:     city,
			Temperatures: make([]int, 0),
			Times:        make([]time.Time, 0),
		}
	}
}

// rateLimit checks if request is within limits
func rateLimit(w http.ResponseWriter) bool {
	requestTracker.Mutex.Lock()
	defer requestTracker.Mutex.Unlock()

	now := time.Now()
	if now.Sub(requestTracker.LastReset) > time.Minute {
		requestTracker.Count = 0
		requestTracker.LastReset = now
	}

	if requestTracker.Count >= MaxRequests {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"error": "Rate limit exceeded"})
		return false
	}

	requestTracker.Count++
	return true
}

// setCommonHeaders sets HTTP headers for API responses
func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}

// recordHistory records temperature for trend analysis
func recordHistory(city string, temp int) {
	historyLock.Lock()
	defer historyLock.Unlock()

	if hist, ok := historicalData[city]; ok {
		hist.Mutex.Lock()
		defer hist.Mutex.Unlock()

		hist.Temperatures = append(hist.Temperatures, temp)
		hist.Times = append(hist.Times, time.Now())

		// Keep only last 48 data points (4 hours with 5-min intervals)
		if len(hist.Temperatures) > 48 {
			hist.Temperatures = hist.Temperatures[1:]
			hist.Times = hist.Times[1:]
		}
	}
}

// getTemperatureTrend returns the trend (up/down/stable)
func getTemperatureTrend(city string) string {
	historyLock.RLock()
	defer historyLock.RUnlock()

	if hist, ok := historicalData[city]; ok {
		hist.Mutex.RLock()
		defer hist.Mutex.RUnlock()

		if len(hist.Temperatures) < 2 {
			return "stable"
		}

		current := hist.Temperatures[len(hist.Temperatures)-1]
		previous := hist.Temperatures[len(hist.Temperatures)-2]

		if current > previous {
			return "rising"
		} else if current < previous {
			return "falling"
		}
		return "stable"
	}
	return "stable"
}

// WMO code to Croatian condition mapping
func wmoCodeToCondition(code int) (string, string) {
	switch {
	case code == 0, code == 1:
		return "Sunčano", "☀️"
	case code == 2:
		return "Djelomično oblačno", "⛅"
	case code == 3:
		return "Oblačno", "☁️"
	case code == 45, code == 48:
		return "Magla", "🌫️"
	case code >= 51 && code <= 67:
		return "Kišno", "🌧️"
	case code >= 71 && code <= 77:
		return "Snježno", "❄️"
	case code >= 80 && code <= 82:
		return "Pljuskovi", "⛈️"
	case code >= 85 && code <= 86:
		return "Snježni pljuskovi", "🌨️"
	case code >= 80 && code <= 82 || code >= 85 && code <= 86:
		return "Oluja", "⛈️"
	default:
		return "Oblačno", "☁️"
	}
}

// fetchRealWeather fetches real weather data from Open-Meteo API
func fetchRealWeather(cityKey string) (*WeatherData, error) {
	coords, ok := cityCoordinates[cityKey]
	if !ok {
		return nil, fmt.Errorf("city not found: %s", cityKey)
	}

	// Open-Meteo API endpoint - free, no API key needed
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m&daily=weather_code,temperature_2m_max,temperature_2m_min&timezone=Europe/Belgrade",
		coords.Latitude,
		coords.Longitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var omResponse OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&omResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	condition, emoji := wmoCodeToCondition(omResponse.Current.WeatherCode)

	weatherData := &WeatherData{
		Location:        coords.Name,
		Temperature:     int(omResponse.Current.Temperature),
		Condition:       condition,
		Emoji:           emoji,
		WindSpeed:       int(omResponse.Current.WindSpeed),
		Humidity:        omResponse.Current.Humidity,
		FeelsLike:       int(omResponse.Current.Temperature) - 2, // Rough estimate
		DramaticMessage: "",
		Description:     "",
	}

	weatherData.DramaticMessage = getDramaticMessage(weatherData.Condition)
	weatherData.Description = getAsciiArt(weatherData.Condition)

	return weatherData, nil
}

// getDramaticMessage returns a random dramatic weather message
func getDramaticMessage(condition string) string {
	messages, ok := dramaticMessages[condition]
	if !ok {
		messages = dramaticMessages["Sunčano"]
	}
	return messages[rand.Intn(len(messages))]
}

// getAsciiArt returns ASCII art for the weather condition
func getAsciiArt(condition string) string {
	art, ok := asciiArts[condition]
	if !ok {
		return "   (...weather brewing...)"
	}
	return art
}

// getDayInCroatian converts English day name to Croatian
func getDayInCroatian(englishDay string) string {
	dayMap := map[string]string{
		"Monday":    "Ponedjeljak",
		"Tuesday":   "Utorak",
		"Wednesday": "Srijeda",
		"Thursday":  "Četvrtak",
		"Friday":    "Petak",
		"Saturday":  "Subota",
		"Sunday":    "Nedjelja",
	}
	if hr, ok := dayMap[englishDay]; ok {
		return hr
	}
	return englishDay
}

// generateForecast creates a 5-day forecast
func generateForecast() []ForecastDay {
	forecast := make([]ForecastDay, 5)
	for i := 0; i < 5; i++ {
		englishDay := time.Now().AddDate(0, 0, i+1).Format("Monday")

		// Generate temperature-appropriate conditions
		high := rand.Intn(15) + 5  // 3-18°C
		low := rand.Intn(4) + (-3) // -3-5°C

		// Select condition based on temperature
		var condition, emoji string
		if high < 4 || low < -2 {
			// Very cold - snow is appropriate
			condition = "Snježno"
			emoji = "❄️"
		} else if high < 5 {
			// Cold but above freezing - rain or cloudy
			rainOrCloud := rand.Intn(2)
			if rainOrCloud == 0 {
				condition = "Kišno"
				emoji = "🌧️"
			} else {
				condition = "Oblačno"
				emoji = "☁️"
			}
		} else {
			// Mild temperatures - varied conditions
			conditions := []string{"Sunčano", "Oblačno", "Kišno", "Djelomično oblačno"}
			emojis := []string{"☀️", "☁️", "🌧️", "⛅"}
			idx := rand.Intn(len(conditions))
			condition = conditions[idx]
			emoji = emojis[idx]
		}

		forecast[i] = ForecastDay{
			Date:      getDayInCroatian(englishDay),
			High:      high,
			Low:       low,
			Condition: condition,
			Emoji:     emoji,
		}
	}
	return forecast
}

// getLocalIP returns the local IP address for network access
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// Handler for root path - HTML dashboard
func weatherDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the embedded HTML file
	htmlContent, err := assets.ReadFile("templates/index.html")
	if err != nil {
		http.Error(w, "HTML file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(htmlContent)
}

// Handler for CSS files
func cssHandler(w http.ResponseWriter, r *http.Request) {
	cssContent, err := assets.ReadFile("static/styles.css")
	if err != nil {
		http.Error(w, "CSS file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/css")
	w.Write(cssContent)
}

// Handler for weather API endpoint
func weatherAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	if !rateLimit(w) {
		return
	}

	location := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/api/weather/"))

	// Get cached weather data
	cacheLock.RLock()
	cached, ok := weatherCache[location]
	cacheLock.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		setCommonHeaders(w)
		json.NewEncoder(w).Encode(map[string]string{"error": "Location not found"})
		return
	}

	// Check if cache needs refresh (every 5 minutes)
	cached.Mutex.Lock()
	defer cached.Mutex.Unlock()

	if time.Since(cached.Timestamp) > CacheRefreshInterval {
		// Refresh with real weather data from API
		weather, err := fetchRealWeather(location)
		if err != nil {
			// Use cached data if API fails
		} else {
			// Successfully fetched real data
			cached.Data = *weather
			cached.Timestamp = time.Now()
			recordHistory(location, weather.Temperature)
		}
	}

	// Add trend data
	cached.Data.UVIndex = float64(rand.Intn(12))
	cached.Data.PrecipChance = rand.Intn(100)

	forecast := WeatherForecast{
		Current:  cached.Data,
		Forecast: generateForecast(),
		AsciiArt: cached.Data.Description,
	}

	setCommonHeaders(w)
	json.NewEncoder(w).Encode(forecast)
}

// Handler for forecast endpoint
func forecastAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	if !rateLimit(w) {
		return
	}

	location := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/api/forecast/"))

	// Get cached weather data to ensure location exists
	cacheLock.RLock()
	cached, ok := weatherCache[location]
	cacheLock.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		setCommonHeaders(w)
		json.NewEncoder(w).Encode(map[string]string{"error": "Location not found"})
		return
	}

	// Fetch real forecast data from Open-Meteo API
	coords, _ := cityCoordinates[location]
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&daily=weather_code,temperature_2m_max,temperature_2m_min&timezone=Europe/Belgrade",
		coords.Latitude,
		coords.Longitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		// Fallback to mock forecast
		setCommonHeaders(w)
		json.NewEncoder(w).Encode(WeatherForecast{
			Current:  cached.Data,
			Forecast: generateForecast(),
		})
		return
	}
	defer resp.Body.Close()

	var omResponse OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&omResponse); err != nil {
		// Fallback to mock forecast
		setCommonHeaders(w)
		json.NewEncoder(w).Encode(WeatherForecast{
			Current:  cached.Data,
			Forecast: generateForecast(),
		})
		return
	}

	// Convert real forecast data to ForecastDay format
	forecast := make([]ForecastDay, 0)
	for i := 0; i < len(omResponse.Daily.Time) && i < 5; i++ {
		condition, emoji := wmoCodeToCondition(omResponse.Daily.WeatherCode[i])
		englishDay := time.Now().AddDate(0, 0, i+1).Format("Monday")
		forecast = append(forecast, ForecastDay{
			Date:      getDayInCroatian(englishDay),
			High:      int(omResponse.Daily.TemperatureMax[i]),
			Low:       int(omResponse.Daily.TemperatureMin[i]),
			Condition: condition,
			Emoji:     emoji,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WeatherForecast{
		Current:  cached.Data,
		Forecast: forecast,
	})
}

// Handler for ASCII art display
func asciiHandler(w http.ResponseWriter, r *http.Request) {
	condition := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/ascii/"))
	if condition == "" {
		condition = "Sunčano"
	}

	art := getAsciiArt(condition)

	htmlTemplate := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>🎨 Weather ASCII Art</title>
    <style>
        body {
            background: #1e1e1e;
            color: #00ff00;
            font-family: 'Courier New', monospace;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
        }
        .ascii-container {
            background: #000;
            padding: 30px;
            border: 2px solid #00ff00;
            border-radius: 5px;
            box-shadow: 0 0 20px rgba(0, 255, 0, 0.3);
            white-space: pre;
            font-size: 1.2em;
            line-height: 1.2;
        }
    </style>
</head>
<body>
    <div class="ascii-container">%s</div>
</body>
</html>
	`, art)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlTemplate)
}

func main() {
	// Hide console window immediately
	hideConsole()

	// Initialize cache with real weather data from Open-Meteo API
	cities := []string{"zagreb", "split", "dubrovnik", "rijeka", "zadar", "osijek"}
	for _, city := range cities {
		// Try to fetch real weather data
		weather, err := fetchRealWeather(city)
		if err != nil {
			// Fallback to mock data if API fails
			mockWeather, ok := locations[city]
			if !ok {
				continue
			}
			mockWeather.DramaticMessage = getDramaticMessage(mockWeather.Condition)
			mockWeather.Description = getAsciiArt(mockWeather.Condition)
			weather = &mockWeather
		}

		weatherCache[city] = &CachedWeatherData{
			Data:      *weather,
			Timestamp: time.Now(),
		}
	}

	// Get local IP for network access
	port := "8081"

	http.HandleFunc("/", weatherDashboardHandler)
	http.HandleFunc("/static/styles.css", cssHandler)
	http.HandleFunc("/api/weather/", weatherAPIHandler)
	http.HandleFunc("/api/forecast/", forecastAPIHandler)
	http.HandleFunc("/ascii/", asciiHandler)
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(0)
	})

	// Open browser automatically
	go func() {
		time.Sleep(2 * time.Second) // Wait a moment for server to start
		if runtime.GOOS == "windows" {
			exec.Command("explorer.exe", "http://localhost:"+port).Start()
		} else if runtime.GOOS == "darwin" {
			exec.Command("open", "http://localhost:"+port).Start()
		} else {
			exec.Command("xdg-open", "http://localhost:"+port).Start()
		}
	}()

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		os.Exit(1)
	}
}
