// Store last viewed location for auto-refresh
let lastViewedLocation = 'zagreb';
let lastRefreshTime = new Date(); // Initialize immediately to current time
let citiesData = {}; // Store fetched data

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

// Render all city cards with current data
function renderCityCards() {
    console.log('Rendering cards with data:', citiesData);
    const grid = document.getElementById('locationsGrid');
    if (!grid) {
        console.error('Grid element not found');
        return;
    }
    
    // Clear existing cards
    grid.innerHTML = '';
    
    const cities = ['zagreb', 'split', 'dubrovnik', 'rijeka', 'zadar'];
    let cardCount = 0;
    
    cities.forEach(city => {
        const data = citiesData[city];
        if (!data) {
            console.log('No data for', city, 'skipping');
            return;
        }
        
        console.log('Creating card for', city, ':', data);
        cardCount++;
        
        const card = document.createElement('div');
        card.className = 'weather-card';
        card.onclick = function() { loadWeather(city); };
        
        card.innerHTML = `
            <div class="location-name">${data.Location}</div>
            <div class="weather-emoji">${data.Emoji}</div>
            <div class="temp">${data.Temperature}°C</div>
            <div class="condition">${data.Condition}</div>
            <div class="details">
                <div class="detail-item">💨 ${data.WindSpeed} km/h</div>
                <div class="detail-item">💧 ${data.Humidity}%</div>
            </div>
            <button class="forecast-btn" onclick="event.stopPropagation(); loadForecast('${city}')">Prognoza od 5 dana</button>
        `;
        
        grid.appendChild(card);
    });
    
    console.log('✓ Cards rendered:', cardCount, 'cards created');
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
    console.log('Starting to load fresh data for all cities...');
    const cities = ['zagreb', 'split', 'dubrovnik', 'rijeka', 'zadar'];
    let loadedCount = 0;
    
    // Add a timeout to show error if nothing loads after 5 seconds
    setTimeout(() => {
        if (loadedCount === 0) {
            console.error('⚠️ No data loaded after 5 seconds. API might be down.');
            const grid = document.getElementById('locationsGrid');
            if (grid) {
                grid.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 40px; color: red;">❌ Greška pri učitavanju podataka</div>';
            }
        }
    }, 5000);
    
    cities.forEach(city => {
        const url = '/api/weather/' + city;
        console.log('Fetching:', url);
        
        fetch(url)
            .then(r => {
                console.log('✓ Response received for', city, 'status:', r.status);
                if (!r.ok) {
                    throw new Error('HTTP ' + r.status + ' ' + r.statusText);
                }
                return r.json();
            })
            .then(data => {
                console.log('✓ JSON parsed for', city, 'data keys:', Object.keys(data));
                console.log('  Current:', data.Current);
                
                if (data.Current) {
                    citiesData[city] = data.Current;
                    loadedCount++;
                    console.log('✓ Stored data for', city, '(' + loadedCount + '/5)');
                    console.log('  Temperature:', data.Current.Temperature, 'Condition:', data.Current.Condition);
                    
                    // Re-render cards once we have any data
                    console.log('Calling renderCityCards with:', Object.keys(citiesData));
                    renderCityCards();
                    lastRefreshTime = new Date();
                    updateRefreshStatus();
                } else {
                    console.error('⚠️ No Current field in response for', city);
                }
            })
            .catch(err => {
                console.error('❌ Failed to load', city, ':', err.message);
                console.error('  Stack:', err.stack);
            });
    });
}

console.log('App script loaded, DOM ready:', document.readyState);
console.log('Grid element found:', !!document.getElementById('locationsGrid'));

// Ensure DOM is ready before initializing
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        console.log('DOMContentLoaded fired');
        updateRefreshStatus();
        loadAllCitiesData();
        startAutoRefresh();
        startStatusUpdater();
    });
} else {
    console.log('DOM already ready, initializing immediately');
    updateRefreshStatus(); // Show current time on page load
    loadAllCitiesData(); // Load fresh data for all cities
    startAutoRefresh(); // Start 15-minute auto-refresh loop
    startStatusUpdater(); // Start 1-second status updater
}

console.log('✓ App initialization scheduled. Waiting for data to load...');
