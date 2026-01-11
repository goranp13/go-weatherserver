// Store last viewed location for auto-refresh
let lastViewedLocation = 'zagreb';
let lastRefreshTime = new Date(); // Initialize immediately to current time

// Function to update a city card with fresh data
function updateCityCard(location, weatherData) {
    // Find the card that corresponds to this location
    const cards = document.querySelectorAll('.weather-card');
    cards.forEach(card => {
        const locationName = card.querySelector('.location-name');
        if (locationName && locationName.textContent.toLowerCase() === location.toLowerCase()) {
            // Update the card with fresh data
            const tempEl = card.querySelector('.temp');
            const conditionEl = card.querySelector('.condition');
            const emojiEl = card.querySelector('.weather-emoji');
            const windEl = card.querySelector('.detail-item:nth-child(1)');
            const humidityEl = card.querySelector('.detail-item:nth-child(2)');
            
            if (tempEl) tempEl.textContent = weatherData.Temperature + '°C';
            if (conditionEl) conditionEl.textContent = weatherData.Condition;
            if (emojiEl) emojiEl.textContent = weatherData.Emoji;
            if (windEl) windEl.textContent = '💨 ' + weatherData.WindSpeed + ' km/h';
            if (humidityEl) humidityEl.textContent = '💧 ' + weatherData.Humidity + '%';
        }
    });
}

function updateRefreshStatus() {
    const statusEl = document.getElementById('refreshStatus');
    if (statusEl && lastRefreshTime) {
        const now = new Date();
        const diff = Math.floor((now - lastRefreshTime) / 1000);
        let timeStr = '';
        if (diff < 60) {
            timeStr = 'upravo sada';
        } else if (diff < 3600) {
            timeStr = 'prije ' + Math.floor(diff / 60) + ' minuta';
        } else {
            timeStr = 'prije ' + Math.floor(diff / 3600) + ' sati';
        }
        statusEl.textContent = 'Zadnja osvježavanja: ' + timeStr;
    }
}

function loadWeather(location) {
    lastViewedLocation = location;
    fetch('/api/weather/' + location)
        .then(r => r.json())
        .then(data => {
            lastRefreshTime = new Date();
            updateRefreshStatus();
            alert(data.Current.Location + '\n' +
                  'Temperatura: ' + data.Current.Temperature + '°C\n' +
                  'Stanje: ' + data.Current.Condition + '\n' +
                  'Osjeća se kao: ' + data.Current.FeelsLike + '°C\n\n' +
                  data.Current.DramaticMessage);
        })
        .catch(err => console.log('loadWeather failed:', err));
}

function loadForecast(location) {
    lastViewedLocation = location;
    fetch('/api/forecast/' + location)
        .then(r => r.json())
        .then(data => {
            lastRefreshTime = new Date();
            updateRefreshStatus();
            let cityName = location.charAt(0).toUpperCase() + location.slice(1);
            let msg = 'Prognoza od 5 dana za ' + cityName + ':\n\n';
            data.Forecast.forEach(day => {
                msg += day.Date + ': ' + day.Emoji + ' ' + day.High + '°C / ' + day.Low + '°C\n';
            });
            alert(msg);
        })
        .catch(err => console.log('loadForecast failed:', err));
}

// Auto-refresh data every 15 minutes (900000 milliseconds)
function startAutoRefresh() {
    const refreshInterval = 15 * 60 * 1000; // 15 minutes
    setInterval(function() {
        if (lastViewedLocation) {
            // Silently refresh data in the background
            fetch('/api/weather/' + lastViewedLocation)
                .then(r => r.json())
                .then(data => {
                    lastRefreshTime = new Date();
                    updateRefreshStatus();
                    console.log('Auto-refresh completed for:', lastViewedLocation, data);
                })
                .catch(err => console.log('Auto-refresh failed:', err));
        }
    }, refreshInterval);
    console.log('Auto-refresh started. Will refresh every 15 minutes.');
}

// Update refresh status every second to show elapsed time
function startStatusUpdater() {
    setInterval(updateRefreshStatus, 1000);
}

// Initialize the app immediately
// Fetch fresh data for all cities on page load
function loadAllCitiesData() {
    const cities = ['zagreb', 'split', 'dubrovnik', 'rijeka', 'zadar'];
    cities.forEach(city => {
        fetch('/api/weather/' + city)
            .then(r => r.json())
            .then(data => {
                updateCityCard(city, data.Current);
                lastRefreshTime = new Date();
                updateRefreshStatus();
                console.log('Loaded data for', city);
            })
            .catch(err => console.log('Failed to load', city, ':', err));
    });
}

updateRefreshStatus(); // Show current time on page load
loadAllCitiesData(); // Load fresh data for all cities
startAutoRefresh(); // Start 15-minute auto-refresh loop
startStatusUpdater(); // Start 1-second status updater

console.log('App initialized. Loading fresh data for all cities.');
