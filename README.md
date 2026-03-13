# 🎵 Music API — Go

API REST construida en Go (solo librería estándar) para gestionar canciones musicales.

**Puerto:** `24770`  
**Carnet:** 24770

---

## 📁 Estructura del proyecto

```
.
├── main.go
├── data/
│   └── songs.json
├── Dockerfile
└── docker-compose.yml
```

---

## 🎯 Tema

Esta API gestiona un catálogo de **canciones** con los siguientes campos:

| Campo              | Tipo   | Descripción                        |
|--------------------|--------|------------------------------------|
| `id`               | int    | Identificador único (auto-generado)|
| `title`            | string | Título de la canción               |
| `artist`           | string | Artista o banda                    |
| `album`            | string | Nombre del álbum                   |
| `genre`            | string | Género musical                     |
| `year`             | int    | Año de lanzamiento                 |
| `duration_seconds` | int    | Duración en segundos               |
| `plays`            | int64  | Número de reproducciones           |

---

## 🚀 Ejecución

### Con Docker Compose

```bash
# Copiar el ejemplo
cp docker-compose.yml.example docker-compose.yml

# Levantar
docker compose up --build
```

El servidor escucha en `http://localhost:24770`.

---

## 📡 Endpoints

### `GET /api/ping`

Verifica que el servidor esté activo.

```bash
curl http://localhost:24770/api/ping
```

```json
{ "message": "pong" }
```

---

### `GET /api/songs`

Devuelve todas las canciones.

```bash
curl http://localhost:24770/api/songs
```

---

### `GET /api/songs?id=1`

Busca una canción por query parameter.

```bash
curl http://localhost:24770/api/songs?id=1
```

---

### `GET /api/songs?genre=Soul`

Filtra por género (case-insensitive, búsqueda parcial).

```bash
curl "http://localhost:24770/api/songs?genre=Soul"
```

---

### `GET /api/songs?artist=adele&year=2011`

**Múltiples filtros combinados** — género, artista y año pueden combinarse libremente.

```bash
curl "http://localhost:24770/api/songs?artist=adele&year=2011"
```

---

### `GET /api/songs/{id}`

Busca una canción por **path parameter**.

```bash
curl http://localhost:24770/api/songs/3
```

```json
{
  "id": 3,
  "title": "Someone Like You",
  "artist": "Adele",
  "album": "21",
  "genre": "Soul",
  "year": 2011,
  "duration_seconds": 285,
  "plays": 2100000000
}
```

---

### `POST /api/songs`

Crea una nueva canción. Todos los campos son requeridos.

```bash
curl -X POST http://localhost:24770/api/songs \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Anti-Hero",
    "artist": "Taylor Swift",
    "album": "Midnights",
    "genre": "Pop",
    "year": 2022,
    "duration_seconds": 200,
    "plays": 3000000000
  }'
```

**Respuesta `201 Created`:**

```json
{
  "id": 13,
  "title": "Anti-Hero",
  ...
}
```

---

### `PUT /api/songs/{id}`

Reemplaza completamente una canción existente.

```bash
curl -X PUT http://localhost:24770/api/songs/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Blinding Lights (Remix)",
    "artist": "The Weeknd",
    "album": "After Hours",
    "genre": "Synth-pop",
    "year": 2020,
    "duration_seconds": 215,
    "plays": 5000000000
  }'
```

---

### `PATCH /api/songs/{id}`

Actualiza parcialmente una canción (solo los campos enviados).

```bash
curl -X PATCH http://localhost:24770/api/songs/1 \
  -H "Content-Type: application/json" \
  -d '{"plays": 9999999999}'
```

---

### `DELETE /api/songs/{id}`

Elimina una canción.

```bash
curl -X DELETE http://localhost:24770/api/songs/1
```

```json
{
  "message": "Song deleted successfully",
  "deleted": { ... }
}
```

---

## ❌ Errores estructurados

Todos los errores devuelven JSON con el siguiente formato:

```json
{
  "error": "Bad Request",
  "code": 400,
  "message": "Missing or invalid fields: title, year (must be between 1900 and 2100)"
}
```

### Casos de error comunes

| Situación                        | Status |
|----------------------------------|--------|
| ID no existe                     | 404    |
| JSON inválido en body            | 400    |
| Campo requerido faltante         | 400    |
| Año fuera de rango               | 400    |
| duration_seconds <= 0            | 400    |
| Método HTTP no soportado         | 405    |

---

## 💾 Persistencia

Los cambios (POST, PUT, PATCH, DELETE) se guardan automáticamente en `data/songs.json`, por lo que **persisten entre reinicios del servidor**.

---

## 🧩 Resumen de criterios

| Criterio                              | Implementado |
|---------------------------------------|-------------|
| GET todos los elementos               | ✅           |
| GET por query parameter (?id=)        | ✅           |
| POST crear elemento                   | ✅           |
| PUT actualización completa            | ✅           |
| PATCH actualización parcial           | ✅           |
| DELETE eliminar                       | ✅           |
| Path parameters (/api/songs/1)        | ✅           |
| Query parameters con filtros          | ✅           |
| Múltiples filtros combinados          | ✅           |
| Validación robusta                    | ✅           |
| Errores JSON estructurados            | ✅           |
| Persistencia en archivo JSON          | ✅           |
| Dockerfile funcional                  | ✅           |
| Puerto según carnet (24770)           | ✅           |