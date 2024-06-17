# Projektname: Go Todo App

## Beschreibung

Diese einfache ToDo-Anwendung wurde serverseitig mit Go und clientseitig mit React und Typescript entwickelt. Für die Entwicklung der Datenbank kam SQLite zum Einsatz. Sie dient dazu Aufgaben zu erstellen, zu bearbeiten und mit anderen Benutzern zu teilen. Dabei können sich Benutzer registrieren und haben somit nur Zugriff auf ihre eigenen Aufgaben bzw. mit ihnen geteilte Aufgaben. Die Aufgaben bestehen aus einem Titel und einer Beschreibung, sowie einer Kategorie, welche der Benutzer frei anlegen und ändern kann. Die Reihenfolge der angezeigten Aufgaben kann ebenfalls benutzerdefiniert angepasst werden. Die Änderungen an den geteilten Aufgaben werden automatisch per WebSocket mit dem Backend geteilt und aktualisiert.

## Installation

### Voraussetzungen

- Go (Version 1.16+)
- SQLite
- Node.js und npm (für das Frontend)

### Backend-Installation

1. Repository klonen:

   ```sh
   git clone https://github.com/Knipp05/go-todo.git
   cd go-todo
   ```

2. Abhängigkeiten installieren:

   ```sh
   go mod download
   ```

3. Datenbank initialisieren:
   ```sh
   go run main.go
   ```

### Frontend-Installation

1. Gehe zum Frontend-Verzeichnis:

   ```sh
   cd client
   ```

2. Abhängigkeiten installieren:

   ```sh
   npm install
   ```

3. Entwicklungsserver starten:
   ```sh
   npm start
   ```

## Verwendung

1. Backend starten:

   ```sh
   go run main.go
   ```

2. Frontend starten:

   ```sh
   cd client
   npm start
   ```

3. Öffne deinen Browser und gehe zu `http://localhost:5173`.

## API-Dokumentation

### Endpunkte

- **POST /api/users/new** - Benutzerregistrierung
- **POST /api/users** - Benutzeranmeldung
- **POST /api/tasks** - Aufgabe hinzufügen
- **DELETE /api/tasks/:id** - Aufgabe löschen
- **PATCH /api/tasks/:id** - Aufgabe aktualisieren
- **POST /api/tasks/:id/:target** - Aufgabe teilen
- **DELETE /api/tasks/:id/:target** - Teilen der Aufgabe beenden
- **PATCH /api/tasks/:idUp/:idDown** - Reihenfolge zweier Aufgaben tauschen
- **POST /api/categories** - Kategorie hinzufügen
- **PATCH /api/categories/:id/delete** - Kategorie löschen
- **PATCH /api/categories/:id** - Kategorie aktualisieren

## WebSocket-Kommunikation

Die WebSocket-Verbindung wird verwendet, um Änderungen an geteilten Aufgaben in Echtzeit zu synchronisieren und andere Clients über die Änderungen zu informieren.

### WebSocket-Endpunkt

- **GET /ws?token=JWT_TOKEN**

Nach der Anmeldung stellt der Client eine Verbindung zum WebSocket-Server her und sendet das JWT-Token als Query-Parameter.

### Nachrichtenformat

## Ordnerstruktur

```plaintext
.

├── client
│   ├── public
│   ├── src
│   │   ├── components
│   │   ├── App.js
│   │   └── index.js
│   ├── package.json
│   └── ...
├── main.go
├── go.mod
├── go.sum
└── README.md
```
