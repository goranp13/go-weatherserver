package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

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
	conditions := []string{"Sunčano", "Oblačno", "Kišno", "Snježno", "Djelomično oblačno"}
	emojis := []string{"☀️", "☁️", "🌧️", "❄️", "⛅"}

	forecast := make([]ForecastDay, 5)
	for i := 0; i < 5; i++ {
		idx := rand.Intn(len(conditions))
		englishDay := time.Now().AddDate(0, 0, i+1).Format("Monday")
		forecast[i] = ForecastDay{
			Date:      getDayInCroatian(englishDay),
			High:      rand.Intn(8) + 5,
			Low:       rand.Intn(5) + (-2),
			Condition: conditions[idx],
			Emoji:     emojis[idx],
		}
	}
	return forecast
}

// Handler for root path - HTML dashboard
func weatherDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the static HTML dashboard (templates/index.html)
	// The frontend JS will fetch API data from /api/* endpoints.
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "templates/index.html")
}

// Handler for weather API endpoint
func weatherAPIHandler(w http.ResponseWriter, r *http.Request) {
	location := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/api/weather/"))

	weather, ok := locations[location]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Location not found"})
		return
	}

	// Generate fresh weather data with slight variations
	weather.Temperature = weather.Temperature + rand.Intn(5) - 2 // ±2°C variation
	weather.FeelsLike = weather.Temperature - 1 - rand.Intn(3)
	weather.WindSpeed = weather.WindSpeed + rand.Intn(5) - 2
	weather.Humidity = weather.Humidity + rand.Intn(10) - 5
	if weather.Humidity < 30 {
		weather.Humidity = 30
	}
	if weather.Humidity > 99 {
		weather.Humidity = 99
	}

	weather.DramaticMessage = getDramaticMessage(weather.Condition)
	weather.Description = getAsciiArt(weather.Condition)

	forecast := WeatherForecast{
		Current:  weather,
		Forecast: generateForecast(),
		AsciiArt: weather.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forecast)
}

// Handler for forecast endpoint
func forecastAPIHandler(w http.ResponseWriter, r *http.Request) {
	location := strings.ToLower(strings.TrimPrefix(r.URL.Path, "/api/forecast/"))

	weather, ok := locations[location]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Location not found"})
		return
	}

	forecast := WeatherForecast{
		Current:  weather,
		Forecast: generateForecast(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forecast)
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
	http.HandleFunc("/", weatherDashboardHandler)
	http.HandleFunc("/api/weather/", weatherAPIHandler)
	http.HandleFunc("/api/forecast/", forecastAPIHandler)
	http.HandleFunc("/ascii/", asciiHandler)

	// Serve static assets (CSS, JS)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println(`
╔════════════════════════════════════════════════════════════════╗
║         🌤️  VREMENSKA PROGNOZA SERVER IS STARTING 🌤️        ║
╚════════════════════════════════════════════════════════════════╝

🌍 Available Locations:
  • /api/weather/zagreb
  • /api/weather/split
  • /api/weather/dubrovnik
  • /api/weather/rijeka
  • /api/weather/zadar

📊 API Endpoints:
  GET / ............................ Interactive Dashboard
  GET /api/weather/<location> ....... JSON Weather Data
  GET /api/forecast/<location> ...... 5-Day Forecast
  GET /ascii/<condition> ............ ASCII Weather Art

🚀 Starting server on http://localhost:8080
Press Ctrl+C to stop...
	`)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}
