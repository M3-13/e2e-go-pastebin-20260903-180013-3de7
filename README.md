# Pastebin-REST-API

Eine kleine Pastebin-REST-API in Go, die ausschließlich mit `net/http` aus der
Standardbibliothek gebaut ist. Pasts werden in einem thread-sicheren
In-Memory-Store (geschützt durch `sync.Mutex`) gehalten und laufen über
`expires_in_seconds` ab. Die API bietet Anlegen, Abrufen, Auflisten und Löschen
von Pasts mit sauberen Statuscodes und JSON-Fehlerantworten.

## Tech Stack

- **Sprache**: Go 1.22
- **Framework**: `net/http` (Standardbibliothek)
- **Speicher**: In-Memory (`sync.Mutex` + `map[string]Paste`)
- **Tests**: `httptest` / Standardbibliothek `testing`
- **Modul**: `pastebin` (Pakete `internal/store` und `internal/api`)

## Installation

```sh
go mod download
```

## Start (Development)

```sh
go run .
```

Der Dienst lauscht auf dem Port aus der Umgebungsvariable `PORT` (Standard
`8080`). Einen anderen Port setzen:

```sh
# Windows (PowerShell)
$env:PORT = "9090"; go run .

# Unix
PORT=9090 go run .
```

## Build (Production)

```sh
go build ./...
```

## Tests

```sh
go test ./...
go test -race ./...
```

## Endpunkte

| Methode | Pfad            | Beschreibung                                                        |
| ------- | --------------- | ------------------------------------------------------------------- |
| GET     | `/healthz`      | Health-Check, antwortet `200 {"status":"ok"}`                       |
| POST    | `/pastes`       | Paste anlegen (Body `{"content": "...", ...}`)                      |
| GET     | `/pastes`       | Alle Pasts als Metadaten-Liste (ohne `content`) auflisten           |
| GET     | `/pastes/{id}`  | Einzelnen Paste inkl. `content` abrufen                             |
| DELETE  | `/pastes/{id}`  | Paste löschen (antwortet `204`)                                     |

Fehlerantworten haben immer das Format `{"error": "<kurze Meldung>"}`.

Alle Antworten setzen `Content-Type: application/json` und
`X-Content-Type-Options: nosniff`; JSON-Antworten escapen HTML-Sonderzeichen.

## Umgebungsvariablen

| Variable | Pflicht | Standard | Beschreibung          |
| -------- | ------- | -------- | --------------------- |
| `PORT`   | nein    | `8080`   | Port, auf dem der Dienst lauscht |

## Feature-Liste

- Health-Check-Endpoint `/healthz`
- Paste anlegen mit zufälliger 16-stelliger Hex-ID (`crypto/rand`)
- Pasts mit Ablaufzeit (`expires_in_seconds`, Standard 86400 s)
- Abgelaufene Pasts werden lazy aus dem Store entfernt
- JSON-Fehler ohne interne Details; HTML-Escaping in Antworten
