# Vremenska prognoza — Weather Forecast (local mock server)

Jednostavan Go webserver koji prikazuje mock vremenske podatke za nekoliko hrvatskih gradova.

## Struktura
- `main.go` — glavni server (API endpoints i handlers)
- `templates/index.html` — frontend HTML
- `static/app.js` — frontend JavaScript
- `static/styles.css` — frontend CSS

## Build
U direktoriju `proba` pokreni:

```bash
go build -o weather-server .
```

Ili jednostavno:

```bash
go build .
```

## Pokretanje

```bash
./weather-server
```

Server će slušati na `http://localhost:8080`.

## Dostupni endpointi
- `GET /` — dashboard (HTML)
- `GET /api/weather/<grad>` — JSON trenutni podaci, primjer: `/api/weather/zagreb`
- `GET /api/forecast/<grad>` — 5-dnevna prognoza (dani na hrvatskom)
- `GET /ascii/<uvjet>` — ASCII art za uvjet (npr. `/ascii/Sunčano`)

## Napomene
- Podaci su trenutno mock i fiksni (mogu se zamijeniti pozivima prema realnim API-jevima ako želiš).
- Frontend dohvaća podatke iz `/api/*` i prikazuje ih u alert prozorima.

Ako želiš, mogu:
- dodati automatsko ažuriranje podataka (fetch svakih N minuta),
- spojiti prave API-je (npr. OpenWeatherMap) i dodati konfiguraciju API ključa,
- ili napraviti `systemd` servis / Dockerfile.
