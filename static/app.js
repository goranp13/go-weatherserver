// Store last viewed location for auto-refresh
let lastViewedLocation = null;
let lastRefreshTime = null;

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
        }
    }
}

function loadWeather(location) {
    lastViewedLocation = location;
    fetch('/api/weather/' + location)
        .then(r => r.json())
        .then(data => {
            alert(data.Current.Location + '\n' +
                  'Temperatura: ' + data.Current.Temperature + '°C\n' +
                  'Stanje: ' + data.Current.Condition + '\n' +
                  'Osjeća se kao: ' + data.Current.FeelsLike + '°C\n\n' +
                  data.Current.DramaticMessage);
        });
}

function loadForecast(location) {
    lastViewedLocation = location;
    fetch('/api/forecast/' + location)
        .then(r => r.json())
        .then(data => {
            let cityName = location.charAt(0).toUpperCase() + location.slice(1);
            let msg = 'Prognoza od 5 dana za ' + cityName + ':\n\n';
            data.Forecast.forEach(day => {
                msg += day.Date + ': ' + day.Emoji + ' ' + day.High + '°C / ' + day.Low + '°C\n';
            });
            alert(msg);
        });
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
    fetch('/api/weather/zagreb')
        .then(r => r.json())
        .then(data => {
            lastRefreshTime = new Date();
            updateRefreshStatus();
            console.log('Initial data loaded:', data);
        })
        .catch(err => console.log('Initial fetch failed:', err));
}

// Update refresh status every second to show elapsed time
function startStatusUpdater() {
    setInterval(updateRefreshStatus, 1000);
}

// Start everything when script loads
fetchInitialData();
startAutoRefresh();
startStatusUpdater();
