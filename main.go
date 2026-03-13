package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Canción representa una entrada en el catálogo musical
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

// Cuando algo sale mal, respondemos con esta estructura
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var songs []Song

const archivoData = "./data/songs.json"

func main() {
	cargarCanciones()

	http.HandleFunc("/api/ping", pingHandler)
	http.HandleFunc("/api/songs", songsHandler)
	http.HandleFunc("/api/songs/", songPorIDHandler)

	log.Println("Servidor corriendo en el puerto :24770")
	log.Fatal(http.ListenAndServe(":24770", nil))
}

// Lee el archivo JSON y carga las canciones en memoria al iniciar
func cargarCanciones() {
	contenido, err := os.ReadFile(archivoData)
	if err != nil {
		log.Fatal("No se pudo leer el archivo de canciones:", err)
	}

	if err := json.Unmarshal(contenido, &songs); err != nil {
		log.Fatal("Error al parsear el JSON:", err)
	}

	log.Printf("%d canciones cargadas correctamente", len(songs))
}

// Guarda el estado actual de las canciones de vuelta al archivo
func guardarCanciones() {
	data, err := json.MarshalIndent(songs, "", "  ")
	if err != nil {
		log.Println("Error al convertir canciones a JSON:", err)
		return
	}

	if err := os.WriteFile(archivoData, data, 0644); err != nil {
		log.Println("Error al guardar el archivo:", err)
	}
}

// Escribe la respuesta en formato JSON con el status indicado
func responderJSON(w http.ResponseWriter, status int, datos interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(datos)
}

// Responde con un error estructurado en JSON
func responderError(w http.ResponseWriter, status int, mensaje string) {
	responderJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: mensaje,
	})
}

// Busca el índice de una canción por su ID, retorna -1 si no existe
func buscarIndice(id int) int {
	for i, s := range songs {
		if s.ID == id {
			return i
		}
	}
	return -1
}

// Genera el siguiente ID disponible basándose en el mayor existente
func siguienteID() int {
	max := 0
	for _, s := range songs {
		if s.ID > max {
			max = s.ID
		}
	}
	return max + 1
}

// Verifica que el servidor esté activo
func pingHandler(w http.ResponseWriter, r *http.Request) {
	responderJSON(w, http.StatusOK, map[string]string{"mensaje": "pong"})
}

// Enruta /api/songs según el método HTTP recibido
func songsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listarCanciones(w, r)
	case http.MethodPost:
		crearCancion(w, r)
	default:
		responderError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// Enruta /api/songs/{id} según el método HTTP recibido
func songPorIDHandler(w http.ResponseWriter, r *http.Request) {
	segmento := strings.TrimPrefix(r.URL.Path, "/api/songs/")
	if segmento == "" {
		responderError(w, http.StatusBadRequest, "Falta el ID en la ruta")
		return
	}

	id, err := strconv.Atoi(segmento)
	if err != nil {
		responderError(w, http.StatusBadRequest, "El ID debe ser un número entero")
		return
	}

	switch r.Method {
	case http.MethodGet:
		obtenerCancion(w, r, id)
	case http.MethodPut:
		reemplazarCancion(w, r, id)
	case http.MethodPatch:
		actualizarCancion(w, r, id)
	case http.MethodDelete:
		eliminarCancion(w, r, id)
	default:
		responderError(w, http.StatusMethodNotAllowed, "Método no permitido")
	}
}

// Devuelve todas las canciones, con soporte para filtros opcionales por género, artista y año
func listarCanciones(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Soporte para buscar por ?id=
	if idParam := q.Get("id"); idParam != "" {
		id, err := strconv.Atoi(idParam)
		if err != nil {
			responderError(w, http.StatusBadRequest, "El parámetro id debe ser un número")
			return
		}
		idx := buscarIndice(id)
		if idx == -1 {
			responderError(w, http.StatusNotFound, "No se encontró la canción con id "+idParam)
			return
		}
		responderJSON(w, http.StatusOK, songs[idx])
		return
	}

	genero := strings.ToLower(q.Get("genre"))
	artista := strings.ToLower(q.Get("artist"))
	anioParam := q.Get("year")

	var anio int
	if anioParam != "" {
		a, err := strconv.Atoi(anioParam)
		if err != nil {
			responderError(w, http.StatusBadRequest, "El parámetro year debe ser un número")
			return
		}
		anio = a
	}

	resultado := []Song{}
	for _, s := range songs {
		if genero != "" && !strings.Contains(strings.ToLower(s.Genre), genero) {
			continue
		}
		if artista != "" && !strings.Contains(strings.ToLower(s.Artist), artista) {
			continue
		}
		if anioParam != "" && s.Year != anio {
			continue
		}
		resultado = append(resultado, s)
	}

	responderJSON(w, http.StatusOK, resultado)
}

// Devuelve una sola canción buscada por su ID en la ruta
func obtenerCancion(w http.ResponseWriter, r *http.Request, id int) {
	idx := buscarIndice(id)
	if idx == -1 {
		responderError(w, http.StatusNotFound, "No se encontró la canción con id "+strconv.Itoa(id))
		return
	}
	responderJSON(w, http.StatusOK, songs[idx])
}

// Valida los campos obligatorios de una canción y devuelve los errores encontrados
func validarCancion(s Song) []string {
	var errores []string
	if strings.TrimSpace(s.Title) == "" {
		errores = append(errores, "title")
	}
	if strings.TrimSpace(s.Artist) == "" {
		errores = append(errores, "artist")
	}
	if strings.TrimSpace(s.Album) == "" {
		errores = append(errores, "album")
	}
	if strings.TrimSpace(s.Genre) == "" {
		errores = append(errores, "genre")
	}
	if s.Year < 1900 || s.Year > 2100 {
		errores = append(errores, "year (debe estar entre 1900 y 2100)")
	}
	if s.DurationSeconds <= 0 {
		errores = append(errores, "duration_seconds (debe ser mayor a 0)")
	}
	return errores
}

// Crea una nueva canción a partir del body JSON y la guarda en el archivo
func crearCancion(w http.ResponseWriter, r *http.Request) {
	var nueva Song
	if err := json.NewDecoder(r.Body).Decode(&nueva); err != nil {
		responderError(w, http.StatusBadRequest, "El body no es un JSON válido")
		return
	}

	if errores := validarCancion(nueva); len(errores) > 0 {
		responderError(w, http.StatusBadRequest, "Campos inválidos o faltantes: "+strings.Join(errores, ", "))
		return
	}

	nueva.ID = siguienteID()
	songs = append(songs, nueva)
	guardarCanciones()

	responderJSON(w, http.StatusCreated, nueva)
}

// Reemplaza todos los datos de una canción existente
func reemplazarCancion(w http.ResponseWriter, r *http.Request, id int) {
	idx := buscarIndice(id)
	if idx == -1 {
		responderError(w, http.StatusNotFound, "No se encontró la canción con id "+strconv.Itoa(id))
		return
	}

	var datos Song
	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		responderError(w, http.StatusBadRequest, "El body no es un JSON válido")
		return
	}

	if errores := validarCancion(datos); len(errores) > 0 {
		responderError(w, http.StatusBadRequest, "Campos inválidos o faltantes: "+strings.Join(errores, ", "))
		return
	}

	datos.ID = id
	songs[idx] = datos
	guardarCanciones()

	responderJSON(w, http.StatusOK, songs[idx])
}

// Actualiza solo los campos enviados sin tocar el resto
func actualizarCancion(w http.ResponseWriter, r *http.Request, id int) {
	idx := buscarIndice(id)
	if idx == -1 {
		responderError(w, http.StatusNotFound, "No se encontró la canción con id "+strconv.Itoa(id))
		return
	}

	var campos map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&campos); err != nil {
		responderError(w, http.StatusBadRequest, "El body no es un JSON válido")
		return
	}

	cancion := songs[idx]

	if v, ok := campos["title"].(string); ok {
		if strings.TrimSpace(v) == "" {
			responderError(w, http.StatusBadRequest, "El título no puede estar vacío")
			return
		}
		cancion.Title = v
	}
	if v, ok := campos["artist"].(string); ok {
		if strings.TrimSpace(v) == "" {
			responderError(w, http.StatusBadRequest, "El artista no puede estar vacío")
			return
		}
		cancion.Artist = v
	}
	if v, ok := campos["album"].(string); ok {
		cancion.Album = v
	}
	if v, ok := campos["genre"].(string); ok {
		cancion.Genre = v
	}
	if v, ok := campos["year"].(float64); ok {
		a := int(v)
		if a < 1900 || a > 2100 {
			responderError(w, http.StatusBadRequest, "El año debe estar entre 1900 y 2100")
			return
		}
		cancion.Year = a
	}
	if v, ok := campos["duration_seconds"].(float64); ok {
		if int(v) <= 0 {
			responderError(w, http.StatusBadRequest, "La duración debe ser mayor a 0")
			return
		}
		cancion.DurationSeconds = int(v)
	}
	if v, ok := campos["plays"].(float64); ok {
		cancion.Plays = int64(v)
	}

	songs[idx] = cancion
	guardarCanciones()

	responderJSON(w, http.StatusOK, songs[idx])
}

// Elimina una canción por su ID y guarda los cambios
func eliminarCancion(w http.ResponseWriter, r *http.Request, id int) {
	idx := buscarIndice(id)
	if idx == -1 {
		responderError(w, http.StatusNotFound, "No se encontró la canción con id "+strconv.Itoa(id))
		return
	}

	eliminada := songs[idx]
	songs = append(songs[:idx], songs[idx+1:]...)
	guardarCanciones()

	responderJSON(w, http.StatusOK, map[string]interface{}{
		"mensaje":   "Canción eliminada correctamente",
		"eliminada": eliminada,
	})
}