VERDICT: CHANGES_REQUESTED

Gesamteinschätzung: Der Code erfüllt wichtige Sicherheits- und Datenschutzanforderungen der Spezifikation bereits gut (Body-Limit, HTML-Escaping, `nosniff`, Fehler ohne interne Details, keine `content`-Felder in Listen, crypto/rand-ID). Für eine DSGVO-konforme und CRA-taugliche Auslieferung fehlen jedoch behebbare technische und dokumentarische Maßnahmen, insbesondere Transportverschlüsselung, Zugriffskontrolle für `DELETE`, Rate-Limiting, aktive Löschung abgelaufener Pasts und CRA-Dokumentation. Da keine grundsätzlich unheilbaren Verstöße erkennbar sind, aber mehrere konkrete Änderungen nötig sind: `CHANGES_REQUESTED`.

---

## 1. DSGVO

### D-01 — Fehlende dokumentierte Rechtsgrundlage und Transparenzpflicht
**Schweregrad:** hoch  
**Datei:** `README.md` (ggf. zusätzlich neue Datei `PRIVACY.md`)  
**Befund:** Die API verarbeitet nutzergenerierte `content`-Inhalte, die personenbezogene Daten enthalten können. Eine Rechtsgrundlage nach Art. 6 DSGVO, ein Verarbeitungszweck, Verantwortlicher, Speicherdauer und Betroffenenrechte sind im vorgelegten Stand nicht dokumentiert. Der Code selbst kann die Rechtsgrundlage nicht festlegen; bei Betrieb ohne diese Festlegung entsteht ein DSGVO-Verstoß.  
**Maßnahme:** Im `README.md` oder in einer neuen `PRIVACY.md` ausdrücklich festhalten:
- Zweck der Verarbeitung (öffentlich abrufbare Kurztextablage),
- Rechtsgrundlage (z. B. Art. 6 Abs. 1 lit. b DSGVO bei Vertragsverhältnis oder Art. 6 Abs. 1 lit. f DSGVO mit dokumentierter Interessenabwägung),
- Kategorien personenbezogener Daten (Inhalte, ggf. IP-Adressen des vorgeschalteten Proxys),
- feste Löschfristen und Verweis auf `expires_in_seconds`,
- Hinweis auf Betroffenenrechte und Kontaktadresse des Verantwortlichen.

---

### D-02 — Unverschlüsselte Übertragung per HTTP
**Schweregrad:** hoch  
**Datei:** `main.go`  
**Befund:** `http.ListenAndServe(":"+port, mux)` stellt ausschließlich unverschlüsseltes HTTP bereit. Wenn personenbezogene Inhalte über das Netzwerk übertragen werden, verletzt das ohne TLS die Anforderungen aus Art. 32 DSGVO (Integrität und Vertraulichkeit).  
**Maßnahme:**
- In `main.go` optionale TLS-Unterstützung ergänzen, z. B.:
  - `TLSCertFile`/`TLSKeyFile` aus Umgebungsvariablen lesen und bei Vorhandensein `http.ListenAndServeTLS` verwenden;
  - Ist kein Zertifikat konfiguriert, im README verbindlich festschreiben, dass das Produkt ausschließlich hinter einer TLS-terminierenden Reverse-Proxy-Schicht betrieben werden darf.
- `README.md`: Deployment-Abschnitt mit klarer Anweisung „Betrieb nur mit TLS-Terminierung (HTTPS); kein Klartext-HTTP im öffentlichen Netz“.

---

### D-03 — Fehlende Zugriffskontrolle für `DELETE` und zu geringe ID-Entropie ohne Rate-Limiting
**Schweregrad:** hoch  
**Dateien:** `internal/api/delete.go`, `internal/store/store.go`, `main.go`  
**Befund:** Jeder, der eine 16-stellige Hex-ID (64 Bit Entropie) kennt, kann den zugehörigen Paste löschen. Es gibt weder ein Eigentümer-Geheimnis noch eine Zugriffsbeschränkung oder ein Rate-Limit. Bei öffentlich erreichbarem Dienst lassen sich IDs durch Brute-Force-Abfragen erraten und fremde Pasts unbefugt auslesen/löschen. Das gefährdet Vertraulichkeit und Integrität personenbezogener Inhalte (Art. 5 Abs. 1 lit. f, Art. 25, Art. 32 DSGVO).  
**Maßnahme:**
- `internal/store/store.go`: In `Add` zusätzlich ein kryptografisches Eigentümer-Token mit mindestens 128 Bit Entropie erzeugen und an `Paste` hängen. Das Token nur in der 201-Antwort zurückgeben, nicht in `PasteMeta`/Listen.
- `internal/api/delete.go`: DELETE nur ausführen, wenn ein gültiges Owner-Token per Header (z. B. `X-Paste-Owner-Token` oder `Authorization: Bearer …`) übermittelt wird; sonst 401/403 JSON-Fehler.
- `internal/store/store.go`: ID von 8 auf 16 Zufallsbytes erhöhen (32 Hex-Zeichen), sofern die Produktspezifikation entsprechend angepasst wird — oder mindestens bestehende 64 Bit nur in Kombination mit Rate-Limiting akzeptieren.
- `main.go`: Einfaches, konfigurierbares Rate-Limit pro Client (z. B. IP-basiertes Token-Bucket) vor `POST /pastes`, `GET /pastes/{id}` und `DELETE /pastes/{id}` schalten. Wichtig: Limit so konzipieren, dass die eigenen `httptest`-Tests nicht blockiert werden (z. B. Rate-Limiter injizierbar oder im Test deaktiviert).

---

### D-04 — Ablaufentsorgung nur bei Zugriff (Lazy Deletion)
**Schweregrad:** mittel  
**Datei:** `internal/store/store.go`  
**Befund:** `Get` und `List` löschen abgelaufene Pasts nur, wenn diese Endpunkte aufgerufen werden. Werden keine Anfragen gestellt, bleiben abgelaufene personenbezogene Inhalte im Speicher und damit faktisch über die festgelegte Speicherfrist hinaus vorhanden. Das widerspricht dem Grundsatz der Speicherbegrenzung (Art. 5 Abs. 1 lit. e DSGVO).  
**Maßnahme:** Im `Store` eine aktive Bereinigung ergänzen:
- In `New()` einen `time.Ticker` starten (z. B. jede Minute) und in einer Goroutine abgelaufene Einträge unter `sync.Mutex` löschen;
- oder bei jedem `Add` zusätzlich eine kleine `sweep()`-Funktion aufrufen.
- Falls ein Hintergrund-Ticker verwendet wird, ein `Close()`/`Stop()` für Tests und sauberes Herunterfahren vorsehen.

---

### D-05 — Unbegrenzte Anzahl und Gesamtgröße gespeicherter Pasts
**Schweregrad:** mittel  
**Dateien:** `internal/api/create.go`, `internal/store/store.go`  
**Befund:** Jeder Request darf bis zu 1 MiB speichern; eine Obergrenze für die Anzahl oder Gesamtgröße der Pasts existiert nicht. Bei öffentlichem Betrieb ermöglicht das Speicher- und DoS-Risiken und kann indirekt zu unverhältnismäßiger Verarbeitung personenbezogener Daten führen (Art. 25 DSGVO).  
**Maßnahme:**
- Im `Store` eine konfigurierbare Höchstanzahl oder Gesamtgröße einführen; bei Überschreiten älteste/längst abgelaufene Pasts verwerfen oder neue Anfrage mit `507 Insufficient Storage`/`429 Too Many Requests` ablehnen.
- `README.md`: Obergrenze dokumentieren.

---

### D-06 — Keine Obergrenze für `expires_in_seconds`
**Schweregrad:** niedrig  
**Datei:** `internal/api/create.go`  
**Befund:** Positive Werte für `expires_in_seconds` sind unbegrenzt; ein Wert wie `2147483647` führt zu einer Speicherung über Jahrzehnte. Ohne dokumentierte Obergrenze kann die Aufbewahrungsfrist unverhältnismäßig lang werden.  
**Maßnahme:**
- In `Create` eine Obergrenze definieren (z. B. 30 Tage) und bei Überschreitung `400` mit kurzer Meldung zurückgeben.
- `README.md`: Obergrenze und Verhalten bei `0` dokumentieren.

---

### D-07 — Betroffenenrechte nicht operationalisiert
**Schweregrad:** niedrig  
**Datei:** `README.md`  
**Befund:** Es gibt keinen dokumentierten Prozess für Auskunfts- oder Löschverlangen Betroffener. Die API erlaubt zwar `DELETE`, aber nur per ID; ein internes Löschverfahren ist nicht beschrieben.  
**Maßnahme:** Im `README.md` oder in `PRIVACY.md` einen Betreiberprozess für Betroffenenanfragen ergänzen (z. B. manuelle Identifikation von Pasts und Löschung über interne Funktion oder API mit Owner-Nachweis).

---

### D-08 — JSON-Decoder akzeptiert nachfolgenden Datenmüll
**Schweregrad:** niedrig  
**Datei:** `internal/api/create.go`  
**Befund:** `dec.Decode(&req)` liest nur das erste JSON-Objekt. Ein Request wie `{"content":"x"}garbage` wird nicht als ungültig erkannt. Das ist primär Robustheit, kann aber unspezifizierte Request-Verarbeitung begünstigen.  
**Maßnahme:** Nach `Decode` prüfen, dass kein weiterer Inhalt folgt:
```go
if err := dec.Decode(&struct{}{}); err != io.EOF {
    WriteError(w, http.StatusBadRequest, "invalid json")
    return
}
```
Benötigt zusätzlichen Import `io`.

---

## 2. EU Cyber Resilience Act (CRA)

### C-01 — Keine sichtbare SBOM / Abhängigkeitsdokumentation
**Schweregrad:** mittel  
**Dateien:** `go.mod`, `README.md`, ggf. neue `SBOM.md`  
**Befund:** Der vorgelegte Stand zeigt `go.mod` (3 Zeilen) ohne dokumentierte Sicherheits-/Abhängigkeitsliste. Der CRA verlangt für Produkte mit digitalen Elementen eine nachvollziehbare Software-Stückliste (SBOM), auch wenn aktuell vor allem die Standardbibliothek genutzt wird.  
**Maßnahme:**
- `go.mod`/`go.sum` versionsgenau pflegen.
- Neue `SBOM.md` oder Abschnitt im `README.md`: alle direkten und transitiven Abhängigkeiten, Versionen, Support-/Patch-Zeitraum.
- Bei reiner Standardbibliothek explizit vermerken: „Keine externen Drittanbieter-Abhängigkeiten; Laufzeit: Go-Standardbibliothek in Go-Version X.Y“.

---

### C-02 — Kein dokumentiertes Update-/Patch-Verfahren und keine Produktversion
**Schweregrad:** mittel  
**Datei:** `README.md`, ggf. neue `SECURITY.md`  
**Befund:** Der Code enthält keine Versionskennzeichnung und keine beschriebene Update-/Patch-Fähigkeit. Der CRA verlangt, dass Sicherheitsupdates über den Supportzeitraum bereitgestellt und dokumentiert werden können.  
**Maßnahme:**
- In `main.go` eine Version als Konstante (z. B. `const version = "0.1.0"`) aufnehmen und in `/healthz` oder separatem `/version` ausgeben.
- `README.md`: Abschnitt „Sicherheitsupdates und Support“, Release-/Versionsschema, Zuständigkeit für Sicherheitspatches.

---

### C-03 — Sicherheitsdokumentation unvollständig
**Schweregrad:** niedrig bis mittel  
**Datei:** `README.md` oder `SECURITY.md`  
**Befund:** Wesentliche Sicherheitseigenschaften sind implementiert, aber nicht zentral dokumentiert. Der CRA verlangt eine nachvollziehbare Beschreibung der Sicherheitsarchitektur und der Restrisiken.  
**Maßnahme:**
- Dokumentieren: 1 MiB Body-Limit, HTML-Escaping, `X-Content-Type-Options: nosniff`, In-Memory-Speicher mit Mutex, `crypto/rand`-IDs, Fehler ohne interne Details.
- Restrisiken benennen: fehlende eingebaute TLS-Terminierung (siehe D-02), kein eingebautes Rate-Limit (siehe D-03), Lazy Deletion (siehe D-04).
- Sichere Defaults und Konfigurationsparameter auflisten (`PORT`, optionale TLS-Umgebungsvariablen, künftige Limits).

---

### C-04 — Transportverschlüsselung als sichere Standardeigenschaft
**Schweregrad:** hoch  
**Datei:** `main.go`  
**Befund:** Wie D-02; der CRA verlangt Schutz vor unbefugtem Zugriff und sichere Konfiguration ab Werk. Ein Dienst, der ohne TLS ausgeliefert wird, ist keine sichere Standardkonfiguration.  
**Maßnahme:** Wie D-02 umsetzen.

---

## 3. EU AI Act

**Nicht anwendbar.** Im vorgelegten Stand ist keine KI-Funktion, kein ML-Modell und keine automatisierte Entscheidungsfindung enthalten. Es bestehen daher keine Pflichten nach AI Act.

---

## 4. Pflichttexte und Benutzeroberfläche

**Für dieses Projekt nicht anwendbar.** Es handelt sich um ein reines Go-Backend ohne öffentliche Web-UI. Daher bestehen keine Legal-Notice-/Cookie-/Consent-Banner-Pflichten im Code. Die datenschutzrechtlichen Transparenzpflichten treffen den Betreiber (siehe D-01), nicht die API selbst. Eine eingebettete Datenschutzerklärung ist technisch nicht erforderlich, solange sie auf der zugehörigen Web-Präsenz bereitgestellt wird.

---

## 5. Barrierefreiheit

**Für dieses Projekt nicht anwendbar.** Es gibt keine Endbenutzer-Weboberfläche; die API liefert ausschließlich JSON. WCAG/BITV/EAA-Pflichten greifen hier nicht.

---

## 6. Positiv erfüllte Anforderungen

- **Datenminimierung:** `GET /pastes` liefert nur Metadaten ohne `content`; nur `GET /pastes/{id}` liefert Inhalte.
- **Kryptografische ID:** `crypto/rand` + Hex-Encoding erzeugt 16-stellige Hex-IDs (AC-15).
- **Body-Limit:** `http.MaxBytesReader` verhindert unbegrenztes Puffern (AC-11).
- **JSON-/Browser-Sicherheit:** `Content-Type: application/json`, `X-Content-Type-Options: nosniff`, HTML-Escaping durch `encoding/json` (AC-12, AC-14).
- **Fehlerbehandlung:** Fehlerantworten enthalten ausschließlich `{"error":"..."}` ohne interne Details (AC-13).
- **Ablaufprüfung bei Lesezugriffen:** Abgelaufene Pasts werden bei `GET`/`List` nicht zurückgegeben (AC-06, AC-17).
- **Keine PII in Anwendungslogs:** `main.go` loggt nur den Startport, keine Request-Inhalte.

---

## 7. Hinweis zur Auflagen-Kompatibilität

Alle geforderten Maßnahmen müssen so umgesetzt werden, dass der bestehende Funktionsumfang weiterhin korrekt läuft. Insbesondere:
- Ein Rate-Limit darf die `httptest`-Tests nicht blockieren und muss daher konfigurierbar oder injizierbar sein.
- TLS-Ergänzungen dürfen den lokalen Entwicklungsbetrieb nicht brechen; TLS sollte optional per Konfiguration aktivierbar sein.
- Ein Hintergrund-Ticker zur Löschung abgelaufener Pasts muss sauber mit dem Store-Lebenszyklus verbunden sein, damit Tests und kurze Prozesse keine Ressourcenlecks oder Flakes erzeugen.
- Ein Owner-Token für `DELETE` muss in der 201-Antwort an den Erzeuger zurückgegeben werden, ohne in Listen oder öffentlichen Metadaten zu erscheinen.