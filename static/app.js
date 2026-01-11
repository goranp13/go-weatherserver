// Store last viewed location for auto-refresh
let lastViewedLocation = null;
let lastRefreshTime = null;

// Function to update a city card with fresh data
function updateCityCard(location, data) {
    const card = document.querySelector(`[onclick*="loadWeather('${location}')"]`);
    if (card) {
        const tempEl = card.querySelector('.temp');
        const conditionEl = card.querySelector('.condition');
        const emojiEl = card.querySelector('.weather-emoji');
        const windEl = card.querySelector('.detail-item:nth-child(1)');
        const humidityEl = card.querySelector('.detail-item:nth-child(2)');
        
        if (tempEl) tempEl.textContent = data.Temperature + '°C';
        if (conditionEl) conditionEl.textContent = data.Condition;
        if (emojiEl) emojiEl.textContent = data.Emoji;
        if (windEl) windEl.textContent = '💨 ' + data.WindSpeed + ' km/h';
        if (humidityEl) humidityEl.textContent = '💧 ' + data.Humidity + '%';
    }
}

// Use injected data from HTML if available
if (typeof window.initialData !== 'undefined' && window.initialData) {
    lastRefreshTime = window.refreshTime;
    lastViewedLocation = 'zagreb';
    updateCityCard('zagreb', window.initialData);
    console.log('Using injected initial data:', window.initialData);
    // Schedule the status update to happen after DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', updateRefreshStatus);
    } else {
        setTimeout(updateRefreshStatus, 0);
    }
}

function updateRefreshStatus() {
    const statusEl = document.getElementById('refreshStatus');
    if (statusEl) {
        if (lastRefreshTime) {
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
        } else {
            statusEl.textContent = 'Zadnja osvježavanja: učitavam...';
        }
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

// Fetch initial data on page load for first city (Zagreb)
function fetchInitialData() {
    lastViewedLocation = 'zagreb';
    const url = '/api/weather/zagreb';
    console.log('Fetching data from:', url);
    
    fetch(url)
        .then(r => {
            console.log('Response status:', r.status);
            if (!r.ok) throw new Error('HTTP error, status=' + r.status);
            return r.json();
        })
        .then(data => {
            lastRefreshTime = new Date();
            updateRefreshStatus();
            console.log('Initial data loaded successfully:', data);
        })
        .catch(err => {
            console.error('Initial fetch failed:', err);
            lastRefreshTime = new Date();
            updateRefreshStatus();
        });
}

// Update refresh status every second to show elapsed time
function startStatusUpdater() {
    setInterval(updateRefreshStatus, 1000);
}

// Start everything when script loads
setTimeout(function() {
    console.log('Initializing app...');
    // If we have injected data, we're already initialized, just start the refresh loops
    if (lastRefreshTime) {
        console.log('Using pre-loaded data, refresh time set to:', lastRefreshTime);
        updateRefreshStatus(); // Update the status message now
    } else {
        // Only fetch if we don't have injected data
        fetchInitialData();
    }
    startAutoRefresh();
    startStatusUpdater();
}, 50);
