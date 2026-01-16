console.log('APP.JS LOADED - Version 3');

//CodeGeex suggestion for fetch retry function
function fetchWithRetry(url, options = {}, retries = 3, delay = 1000) {
    return new Promise((resolve, reject) => {
        const attempt = (n) => {
            fetch(url, options)
                .then(response => {
                    if (!response.ok) throw new Error(`HTTP ${response.status}`);
                    return response.json();
                })
                .then(resolve)
                .catch(error => {
                    if (n <= 0) return reject(error);
                    console.log(`Retrying ${url}, attempts left: ${n}`);
                    setTimeout(() => attempt(n - 1), delay);
                });
        };
        attempt(retries);
    });
}

// Store last viewed location for auto-refresh
let lastViewedLocation = 'zagreb';
let lastRefreshTime = new Date();

// Dark mode handling
function initDarkMode() {
    const darkModeToggle = document.getElementById('darkModeToggle');
    const savedTheme = localStorage.getItem('theme') || 'light';
    
    if (savedTheme === 'dark') {
        document.documentElement.classList.add('dark-mode');
        darkModeToggle.textContent = '☀️';
    }
    
    darkModeToggle.addEventListener('click', function() {
        document.documentElement.classList.toggle('dark-mode');
        const isDark = document.documentElement.classList.contains('dark-mode');
        localStorage.setItem('theme', isDark ? 'dark' : 'light');
        darkModeToggle.textContent = isDark ? '☀️' : '🌙';
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

function updateCardData(city, data) {
    const card = document.querySelector(`[data-city="${city}"]`);
    if (!card) {
        console.error('Card not found for:', city);
        return;
    }
    
    console.log('Updating card for', city, 'with:', data);
    console.log('Card element:', card);
    
    const tempEl = card.querySelector('.temp');
    const conditionEl = card.querySelector('.condition');
    const emojiEl = card.querySelector('.weather-emoji');
    const detailItems = card.querySelectorAll('.detail-item');
        
    console.log('Temperature element:', tempEl);
    console.log('Condition element:', conditionEl);
    console.log('Emoji element:', emojiEl);
    console.log('Detail items:', detailItems);
    
    if (tempEl) tempEl.textContent = data.Temperature + '°C';
    if (conditionEl) conditionEl.textContent = data.Condition;
    if (emojiEl) emojiEl.textContent = data.Emoji;
    if (detailItems[0]) detailItems[0].textContent = '💨 ' + data.WindSpeed + ' km/h';
    if (detailItems[1]) detailItems[1].textContent = '💧 ' + data.Humidity + '%';
    
    console.log('✓ Card updated for', city);
}

function loadWeather(location) {
    lastViewedLocation = location;
    fetch('/api/weather/' + location)
        .then(r => r.json())
        .then(data => {
            lastRefreshTime = new Date();
            updateRefreshStatus();
            const msg = `${data.Current.Location}\n` +
                  `Temperatura: ${data.Current.Temperature}°C\n` +
                  `Stanje: ${data.Current.Condition}\n` +
                  `Osjeća se kao: ${data.Current.FeelsLike}°C\n` +
                  `Vlaga: ${data.Current.Humidity}%\n` +
                  `Vjetar: ${data.Current.WindSpeed} km/h\n\n` +
                  `${data.Current.DramaticMessage}`;
            alert(msg);
        })
        .catch(err => {
            console.error('loadWeather failed:', err);
            alert('Greška pri učitavanju podataka');
        });
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
                msg += day.Date + ': ' + day.Emoji + ' ' + day.High + '°C / ' + day.Low + '°C - ' + day.Condition + '\n';
            });
            alert(msg);
        })
        .catch(err => {
            console.error('loadForecast failed:', err);
            alert('Greška pri učitavanju prognoze');
        });
}

// Auto-refresh data every 15 minutes
function startAutoRefresh() {
    const refreshInterval = 15 * 60 * 1000;
    setInterval(function() {
        if (lastViewedLocation) {
            fetch('/api/weather/' + lastViewedLocation)
                .then(r => r.json())
                .then(data => {
                    lastRefreshTime = new Date();
                    updateRefreshStatus();
                    console.log('Auto-refresh completed');
                })
                .catch(err => console.log('Auto-refresh failed:', err));
        }
    }, refreshInterval);
    console.log('Auto-refresh started');
}

// Update refresh status every second
function startStatusUpdater() {
    setInterval(updateRefreshStatus, 1000);
}

// Load all cities data
function loadAllCitiesData() {
    console.log('Starting to load fresh data for all cities...');
    const cities = ['zagreb', 'split', 'dubrovnik', 'rijeka', 'zadar', 'osijek'];
    
    cities.forEach(city => {
        const url = '/api/weather/' + city;
        console.log('Fetching:', url);
        
        // Use regular fetch to debug the issue
        fetch(url)
            .then(response => {
                console.log('Response received for', city, 'status:', response.status);
                console.log('Response headers:', response.headers);
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                return response.json();
            })
            .then(data => {
                console.log('Data received for', city, ':', data);
                console.log('Data type:', typeof data);
                console.log('Data keys:', data ? Object.keys(data) : 'null/undefined');
                
                // Validate data structure
                if (!data || typeof data !== 'object') {
                    throw new Error('Invalid data format received');
                }
                
                if (data.Current) {
                    // Use data.Current for weather alerts
                    updateCardData(city, data.Current);
                } else if (data.Temperature !== undefined) {
                    // Direct weather data format
                    updateCardData(city, data);
                } else {
                    throw new Error('Missing required weather data fields');
                }
                
                lastRefreshTime = new Date();
                updateRefreshStatus();
            })
            .catch(err => {
                console.error('Failed to load data for', city, ':', err);
                
                // Show user-friendly error message
                const card = document.querySelector(`[data-city="${city}"]`);
                if (card) {
                    const tempEl = card.querySelector('.temp');
                    const conditionEl = card.querySelector('.condition');
                    if (tempEl) tempEl.textContent = 'N/A';
                    if (conditionEl) conditionEl.textContent = 'Data unavailable';
                    
                    // You can also consider displaying an alert or updating the UI with an error message
                }
            });
    });
}


// Add click listeners to weather cards
function initCardClickListeners() {
    const cards = document.querySelectorAll('.weather-card');
    cards.forEach(card => {
        card.addEventListener('click', function() {
            const city = this.getAttribute('data-city');
            if (city) {
                loadWeather(city);
            }
        });
    });
}

// Initialize app
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        console.log('DOM ready');
        initDarkMode();
        initCardClickListeners();
        updateRefreshStatus();
        loadAllCitiesData();
        startAutoRefresh();
        startStatusUpdater();
    });
} else {
    console.log('DOM already ready');
    initDarkMode();
    initCardClickListeners();
    updateRefreshStatus();
    loadAllCitiesData();
    startAutoRefresh();
    startStatusUpdater();
}
