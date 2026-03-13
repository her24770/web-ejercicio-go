package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)


// Modelos de objetos

type Song struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	Album           string `json:"album"`
	Genre           string `json:"genre"`
	Year            int    `json:"year"`
	DurationSeconds int    `json:"duration_seconds"`
	Plays           int64  `json:"plays"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// -------------------------
// In-memory store
// -------------------------

var songs []Song

const dataFile = "./data/songs.json"

// -------------------------
// Entry point
// -------------------------

func main() {
	loadSongs()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/songs", songsHandler)
	http.HandleFunc("/api/songs/", songByIDHandler)

	log.Println("Music API running on :24770")
	log.Fatal(http.ListenAndServe(":24770", nil))
}

// -------------------------
// Load / Save
// -------------------------

func loadSongs() {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		log.Fatal("Error reading songs file:", err)
	}
	if err := json.Unmarshal(file, &songs); err != nil {
		log.Fatal("Error parsing songs JSON:", err)
	}
	log.Printf("Loaded %d songs from %s", len(songs), dataFile)
}

func saveSongs() {
	data, err := json.MarshalIndent(songs, "", "  ")
	if err != nil {
		log.Println("Error marshaling songs:", err)
		return
	}
	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		log.Println("Error writing songs file:", err)
	}
}

// -------------------------
// Helpers
// -------------------------

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println("Error encoding response:", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	})
}

func generateNextID() int {
	maxID := 0
	for _, s := range songs {
		if s.ID > maxID {
			maxID = s.ID
		}
	}
	return maxID + 1
}

func findSongIndex(id int) int {
	for i, s := range songs {
		if s.ID == id {
			return i
		}
	}
	return -1
}

// -------------------------
// Handlers
// -------------------------

func pingHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
}

// /api/songs  →  GET (list + filters) | POST (create)
func songsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetSongs(w, r)
	case http.MethodPost:
		handleCreateSong(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Supported: GET, POST")
	}
}

// /api/songs/{id}  →  GET | PUT | PATCH | DELETE
func songByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/songs/3
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/songs/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		writeError(w, http.StatusBadRequest, "Missing song ID in path")
		return
	}

	id, err := strconv.Atoi(pathParts[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid song ID: must be a number")
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGetSongByID(w, r, id)
	case http.MethodPut:
		handleUpdateSong(w, r, id)
	case http.MethodPatch:
		handlePatchSong(w, r, id)
	case http.MethodDelete:
		handleDeleteSong(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed. Supported: GET, PUT, PATCH, DELETE")
	}
}

// -------------------------
// GET /api/songs
// Filters: ?id=1  ?genre=Pop  ?artist=Adele  ?year=2019
// Combinable: ?genre=Soul&year=2010
// -------------------------

func handleGetSongs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Legacy ?id= query param support
	if idParam := q.Get("id"); idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid id parameter: must be a number")
			return
		}
		idx := findSongIndex(id)
		if idx == -1 {
			writeError(w, http.StatusNotFound, "Song with id "+idParam+" not found")
			return
		}
		writeJSON(w, http.StatusOK, songs[idx])
		return
	}

	// Multi-filter support
	genreFilter := strings.ToLower(q.Get("genre"))
	artistFilter := strings.ToLower(q.Get("artist"))
	yearFilter := q.Get("year")

	var yearVal int
	if yearFilter != "" {
		y, err := strconv.Atoi(yearFilter)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid year parameter: must be a number")
			return
		}
		yearVal = y
	}

	result := []Song{}
	for _, s := range songs {
		if genreFilter != "" && !strings.Contains(strings.ToLower(s.Genre), genreFilter) {
			continue
		}
		if artistFilter != "" && !strings.Contains(strings.ToLower(s.Artist), artistFilter) {
			continue
		}
		if yearFilter != "" && s.Year != yearVal {
			continue
		}
		result = append(result, s)
	}

	writeJSON(w, http.StatusOK, result)
}

// -------------------------
// GET /api/songs/{id}
// -------------------------

func handleGetSongByID(w http.ResponseWriter, r *http.Request, id int) {
	idx := findSongIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "Song with id "+strconv.Itoa(id)+" not found")
		return
	}
	writeJSON(w, http.StatusOK, songs[idx])
}

// -------------------------
// POST /api/songs
// -------------------------

func handleCreateSong(w http.ResponseWriter, r *http.Request) {
	var input Song
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
		return
	}

	// Validation
	var missing []string
	if strings.TrimSpace(input.Title) == "" {
		missing = append(missing, "title")
	}
	if strings.TrimSpace(input.Artist) == "" {
		missing = append(missing, "artist")
	}
	if strings.TrimSpace(input.Album) == "" {
		missing = append(missing, "album")
	}
	if strings.TrimSpace(input.Genre) == "" {
		missing = append(missing, "genre")
	}
	if input.Year < 1900 || input.Year > 2100 {
		missing = append(missing, "year (must be between 1900 and 2100)")
	}
	if input.DurationSeconds <= 0 {
		missing = append(missing, "duration_seconds (must be > 0)")
	}
	if len(missing) > 0 {
		writeError(w, http.StatusBadRequest, "Missing or invalid fields: "+strings.Join(missing, ", "))
		return
	}

	input.ID = generateNextID()
	songs = append(songs, input)
	saveSongs()

	writeJSON(w, http.StatusCreated, input)
}

// -------------------------
// PUT /api/songs/{id}  (full replace)
// -------------------------

func handleUpdateSong(w http.ResponseWriter, r *http.Request, id int) {
	idx := findSongIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "Song with id "+strconv.Itoa(id)+" not found")
		return
	}

	var input Song
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
		return
	}

	// Validation
	var missing []string
	if strings.TrimSpace(input.Title) == "" {
		missing = append(missing, "title")
	}
	if strings.TrimSpace(input.Artist) == "" {
		missing = append(missing, "artist")
	}
	if strings.TrimSpace(input.Album) == "" {
		missing = append(missing, "album")
	}
	if strings.TrimSpace(input.Genre) == "" {
		missing = append(missing, "genre")
	}
	if input.Year < 1900 || input.Year > 2100 {
		missing = append(missing, "year (must be between 1900 and 2100)")
	}
	if input.DurationSeconds <= 0 {
		missing = append(missing, "duration_seconds (must be > 0)")
	}
	if len(missing) > 0 {
		writeError(w, http.StatusBadRequest, "Missing or invalid fields: "+strings.Join(missing, ", "))
		return
	}

	input.ID = id
	songs[idx] = input
	saveSongs()

	writeJSON(w, http.StatusOK, songs[idx])
}

// -------------------------
// PATCH /api/songs/{id}  (partial update)
// -------------------------

func handlePatchSong(w http.ResponseWriter, r *http.Request, id int) {
	idx := findSongIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "Song with id "+strconv.Itoa(id)+" not found")
		return
	}

	// Decode into a map to allow partial fields
	var fields map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
		return
	}

	song := songs[idx]

	if v, ok := fields["title"].(string); ok {
		if strings.TrimSpace(v) == "" {
			writeError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		song.Title = v
	}
	if v, ok := fields["artist"].(string); ok {
		if strings.TrimSpace(v) == "" {
			writeError(w, http.StatusBadRequest, "artist cannot be empty")
			return
		}
		song.Artist = v
	}
	if v, ok := fields["album"].(string); ok {
		song.Album = v
	}
	if v, ok := fields["genre"].(string); ok {
		song.Genre = v
	}
	if v, ok := fields["year"].(float64); ok {
		y := int(v)
		if y < 1900 || y > 2100 {
			writeError(w, http.StatusBadRequest, "year must be between 1900 and 2100")
			return
		}
		song.Year = y
	}
	if v, ok := fields["duration_seconds"].(float64); ok {
		if int(v) <= 0 {
			writeError(w, http.StatusBadRequest, "duration_seconds must be > 0")
			return
		}
		song.DurationSeconds = int(v)
	}
	if v, ok := fields["plays"].(float64); ok {
		song.Plays = int64(v)
	}

	songs[idx] = song
	saveSongs()

	writeJSON(w, http.StatusOK, songs[idx])
}

// -------------------------
// DELETE /api/songs/{id}
// -------------------------

func handleDeleteSong(w http.ResponseWriter, r *http.Request, id int) {
	idx := findSongIndex(id)
	if idx == -1 {
		writeError(w, http.StatusNotFound, "Song with id "+strconv.Itoa(id)+" not found")
		return
	}

	deleted := songs[idx]
	songs = append(songs[:idx], songs[idx+1:]...)
	saveSongs()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Song deleted successfully",
		"deleted": deleted,
	})
}