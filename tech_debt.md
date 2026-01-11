# Technical Debt & Enhancement Opportunities

## Overview
This document outlines architectural improvements and code enhancements for the todo-go project.

---

## 1. Separate Layers (Clean Architecture) 🏗️
**Priority**: HIGH

### Current Issue
All code is in a single file (`todo.go` + `main.go`), mixing data models, database operations, and HTTP handlers.

### Proposed Structure
```
todo-go/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── models/
│   │   └── todo.go              # Todo struct
│   ├── repository/
│   │   ├── interface.go         # TodoRepository interface
│   │   └── postgres.go          # PostgreSQL implementation
│   ├── service/
│   │   └── todo_service.go      # Business logic layer
│   └── handler/
│       └── todo_handler.go      # HTTP handlers
├── pkg/
│   ├── database/
│   │   └── postgres.go          # DB connection setup
│   └── config/
│       └── config.go            # Configuration management
├── migrations/
│   └── 001_create_todos.sql    # SQL migrations
├── docker-compose.dev.yml
└── README.md
```

### Benefits
- Clear separation of concerns
- Easier testing (mock interfaces)
- Better maintainability
- Standard Go project layout

---

## 2. Use Interfaces for Testability 🧪
**Priority**: HIGH

### Current Issue
Direct database calls make unit testing impossible without a real database.

### Solution
Define repository interface:

```go
type TodoRepository interface {
    Add(ctx context.Context, todo models.Todo) (int, error)
    Update(ctx context.Context, todo models.Todo) error
    Delete(ctx context.Context, id int) error
    GetByID(ctx context.Context, id int) (*models.Todo, error)
    List(ctx context.Context) ([]models.Todo, error)
}
```

### Benefits
- Easy mocking with `gomock` or `testify/mock`
- Swap implementations (e.g., in-memory for tests)
- Dependency injection support

---

## 3. Add Missing CRUD Operation 🔍
**Priority**: MEDIUM

### Current Issue
No `GetByID` method exists. Only list all or nothing.

### Solution
```go
func GetByID(db *sql.DB, id int) (*Todo, error) {
    sqlStatement := `SELECT id, title, description, is_done, created_at, updated_at 
                     FROM todos WHERE id=$1`
    var todo Todo
    err := db.QueryRow(sqlStatement, id).Scan(
        &todo.ID, &todo.Title, &todo.Description, 
        &todo.IsDone, &todo.CreatedAt, &todo.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("todo with id %d not found", id)
    }
    if err != nil {
        return nil, err
    }
    return &todo, nil
}
```

### API Route
```go
http.HandleFunc("GET /todos/{id}", getTodoByIDHandler)
```

---

## 4. Fix Timestamp Bug in Update ⚠️
**Priority**: CRITICAL

### Current Issue
In `todo.go:27`, `Update()` overwrites `created_at` with current time:
```go
// WRONG: Don't update created_at!
sqlStatement := "UPDATE todos SET title=$1, description=$2, is_done=$3, created_at=$4, updated_at=$5 WHERE id=$6"
```

### Fix
```go
func Update(db *sql.DB, todo Todo) error {
    sqlStatement := `UPDATE todos 
                     SET title=$1, description=$2, is_done=$3, updated_at=$4 
                     WHERE id=$5`
    res, err := db.Exec(sqlStatement, 
        todo.Title, todo.Description, todo.IsDone, time.Now(), todo.ID)
    if err != nil {
        return err
    }
    rows, _ := res.RowsAffected()
    if rows == 0 {
        return sql.ErrNoRows
    }
    return nil
}
```

---

## 5. Add Context Support 🕐
**Priority**: HIGH

### Current Issue
No timeout or cancellation support. Long-running queries can block indefinitely.

### Solution
Update all database functions:
```go
func Add(ctx context.Context, db *sql.DB, todo Todo) (int, error) {
    sqlStatement := `INSERT INTO todos 
                     (title, description, is_done, created_at, updated_at) 
                     VALUES ($1, $2, $3, $4, $5) RETURNING id`
    var id int
    err := db.QueryRowContext(ctx, sqlStatement, 
        todo.Title, todo.Description, false, time.Now(), time.Now()).Scan(&id)
    if err != nil {
        return 0, err
    }
    return id, nil
}
```

### Handler Example
```go
func createTodoHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    // Use ctx in repository calls
    id, err := repo.Add(ctx, todo)
    // ...
}
```

---

## 6. Replace database/sql with sqlx or GORM 📦
**Priority**: MEDIUM

### Option A: sqlx (Lightweight)
```go
import "github.com/jmoiron/sqlx"

func AllList(db *sqlx.DB) ([]Todo, error) {
    todos := []Todo{}
    err := db.Select(&todos, "SELECT * FROM todos")
    return todos, err
}
```

### Option B: GORM (Full ORM)
```go
import "gorm.io/gorm"

type Todo struct {
    gorm.Model
    Title       string
    Description string
    IsDone      bool
}

func AllList(db *gorm.DB) ([]Todo, error) {
    var todos []Todo
    result := db.Find(&todos)
    return todos, result.Error
}
```

### Benefits
- Less boilerplate code
- Automatic struct scanning
- Built-in connection pooling (sqlx)
- Migration support (GORM)

---

## 7. Add Input Validation 🛡️
**Priority**: HIGH

### Current Issue
No validation on incoming data. Can insert empty titles, SQL injection risks, etc.

### Solution
```go
package models

import (
    "errors"
    "strings"
)

func (t *Todo) Validate() error {
    t.Title = strings.TrimSpace(t.Title)
    t.Description = strings.TrimSpace(t.Description)
    
    if t.Title == "" {
        return errors.New("title is required")
    }
    if len(t.Title) > 255 {
        return errors.New("title must be less than 255 characters")
    }
    if len(t.Description) > 1000 {
        return errors.New("description must be less than 1000 characters")
    }
    return nil
}
```

### Integration
```go
func createTodoHandler(w http.ResponseWriter, r *http.Request) {
    var todo Todo
    json.NewDecoder(r.Body).Decode(&todo)
    
    if err := todo.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // ...
}
```

### Consider Using
- [`go-playground/validator`](https://github.com/go-playground/validator) for struct tags:
  ```go
  type Todo struct {
      Title string `json:"title" validate:"required,max=255"`
  }
  ```

---

## 8. Implement Error Handling Middleware 🚨
**Priority**: MEDIUM

### Current Issue
Inconsistent error responses, HTTP 200 returned even on errors.

### Solution
```go
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

func errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := h(w, r); err != nil {
            var appErr *AppError
            if errors.As(err, &appErr) {
                w.WriteHeader(appErr.Code)
                json.NewEncoder(w).Encode(appErr)
                return
            }
            
            // Unknown error
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(AppError{
                Code:    500,
                Message: "Internal server error",
            })
        }
    }
}
```

### Usage
```go
http.HandleFunc("POST /todos", errorHandler(createTodo))

func createTodo(w http.ResponseWriter, r *http.Request) error {
    // Return errors instead of handling them inline
    if err := validate(todo); err != nil {
        return &AppError{Code: 400, Message: err.Error()}
    }
    // ...
}
```

---

## 9. Environment-Based Configuration 🔧
**Priority**: HIGH

### Current Issue
Hardcoded database credentials in `main.go`:
```go
psqlInfo := "host=localhost port=5432 user=postgres password=postgres dbname=todo_db sslmode=disable"
```

### Solution: Create Config Package
```go
// pkg/config/config.go
package config

import (
    "fmt"
    "os"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
}

type ServerConfig struct {
    Port string
    Env  string // development, staging, production
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),
            Env:  getEnv("APP_ENV", "development"),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", "postgres"),
            DBName:   getEnv("DB_NAME", "todo_db"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
    }
}

func (c *DatabaseConfig) ConnectionString() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
    )
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

### .env File
```env
# .env.development
APP_ENV=development
SERVER_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=todo_db
DB_SSLMODE=disable
```

### Use godotenv
```go
import "github.com/joho/godotenv"

func main() {
    godotenv.Load(".env.development")
    cfg := config.Load()
    // ...
}
```

---

## 10. Add Unit Tests 🧪
**Priority**: HIGH

### Current Issue
Zero test coverage.

### Example: Repository Tests
```go
// internal/repository/postgres_test.go
package repository

import (
    "context"
    "testing"
    "time"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()
    
    repo := NewPostgresTodoRepository(db)
    
    todo := models.Todo{
        Title:       "Test Todo",
        Description: "Test Description",
    }
    
    rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
    mock.ExpectQuery("INSERT INTO todos").
        WithArgs(todo.Title, todo.Description, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
        WillReturnRows(rows)
    
    id, err := repo.Add(context.Background(), todo)
    
    assert.NoError(t, err)
    assert.Equal(t, 1, id)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()
    
    repo := NewPostgresTodoRepository(db)
    
    mock.ExpectQuery("SELECT (.+) FROM todos WHERE id=\\$1").
        WithArgs(999).
        WillReturnError(sql.ErrNoRows)
    
    _, err := repo.GetByID(context.Background(), 999)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

### Example: Handler Tests
```go
// internal/handler/todo_handler_test.go
func TestCreateTodoHandler(t *testing.T) {
    mockRepo := new(mocks.TodoRepository)
    handler := NewTodoHandler(mockRepo)
    
    todo := models.Todo{Title: "Test", Description: "Desc"}
    mockRepo.On("Add", mock.Anything, todo).Return(1, nil)
    
    body, _ := json.Marshal(todo)
    req := httptest.NewRequest("POST", "/todos", bytes.NewReader(body))
    w := httptest.NewRecorder()
    
    handler.CreateTodo(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    mockRepo.AssertExpectations(t)
}
```

### Test Coverage Goal
- Repository: 80%+
- Service: 70%+
- Handlers: 60%+

---

## 11. Add Logging 📝
**Priority**: MEDIUM

### Current Issue
Only `fmt.Println` in error cases. No structured logging.

### Solution: Use slog (Go 1.21+)
```go
import "log/slog"

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    
    logger.Info("Server starting", "port", cfg.Server.Port)
    // ...
}
```

### In Handlers
```go
func createTodoHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Creating todo", "title", todo.Title)
    
    id, err := repo.Add(ctx, todo)
    if err != nil {
        logger.Error("Failed to create todo", "error", err)
        // ...
    }
    
    logger.Info("Todo created successfully", "id", id)
}
```

### Consider
- Request ID middleware for tracing
- Log levels per environment (DEBUG in dev, INFO in prod)

---

## 12. Add Database Migration Tool 🗄️
**Priority**: MEDIUM

### Current Issue
Manual SQL execution for schema changes.

### Solution: golang-migrate
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Migration Files
```sql
-- migrations/000001_create_todos_table.up.sql
CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_done BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_todos_is_done ON todos(is_done);
```

```sql
-- migrations/000001_create_todos_table.down.sql
DROP TABLE IF EXISTS todos;
```

### Run Migrations
```go
import "github.com/golang-migrate/migrate/v4"

func runMigrations(dbURL string) error {
    m, err := migrate.New(
        "file://migrations",
        dbURL,
    )
    if err != nil {
        return err
    }
    return m.Up()
}
```

---

## 13. Add Pagination 📄
**Priority**: MEDIUM

### Current Issue
`AllList()` returns all records. Will cause memory issues with 10k+ todos.

### Solution
```go
type ListOptions struct {
    Limit  int
    Offset int
    SortBy string // "created_at", "updated_at", "title"
    Order  string // "ASC", "DESC"
}

func List(db *sql.DB, opts ListOptions) ([]Todo, int, error) {
    // Default values
    if opts.Limit <= 0 {
        opts.Limit = 20
    }
    if opts.SortBy == "" {
        opts.SortBy = "created_at"
    }
    if opts.Order == "" {
        opts.Order = "DESC"
    }
    
    // Count total
    var total int
    db.QueryRow("SELECT COUNT(*) FROM todos").Scan(&total)
    
    // Query with pagination
    query := fmt.Sprintf(
        `SELECT * FROM todos ORDER BY %s %s LIMIT $1 OFFSET $2`,
        opts.SortBy, opts.Order,
    )
    rows, err := db.Query(query, opts.Limit, opts.Offset)
    // ...scan todos...
    
    return todos, total, err
}
```

### API Response
```json
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "page": 1,
    "total_pages": 8
  }
}
```

---

## 14. Add Request Validation Middleware 🛡️
**Priority**: MEDIUM

### Solution
```go
func validateContentType(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" || r.Method == "PUT" {
            ct := r.Header.Get("Content-Type")
            if ct != "application/json" {
                http.Error(w, "Content-Type must be application/json", 415)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}

func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## 15. Add Health Check Endpoint 🏥
**Priority**: LOW

### Solution
```go
func healthHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
        defer cancel()
        
        // Check DB connection
        if err := db.PingContext(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "database": "down",
            })
            return
        }
        
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
            "database": "up",
            "version": "1.0.0",
        })
    }
}
```

### Routes
```go
http.HandleFunc("GET /health", healthHandler(db))
http.HandleFunc("GET /ready", readinessHandler(db))
```

---

## 16. Use Chi or Gin Router 🛣️
**Priority**: MEDIUM

### Current Issue
Standard `http.ServeMux` lacks middleware chaining, route groups, and params.

### Option A: Chi (Lightweight)
```go
import "github.com/go-chi/chi/v5"

func main() {
    r := chi.NewRouter()
    
    // Middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(cors)
    
    // Routes
    r.Route("/todos", func(r chi.Router) {
        r.Get("/", listTodosHandler)
        r.Post("/", createTodoHandler)
        r.Get("/{id}", getTodoHandler)
        r.Put("/{id}", updateTodoHandler)
        r.Delete("/{id}", deleteTodoHandler)
    })
    
    http.ListenAndServe(":8080", r)
}
```

### Option B: Gin (Full-featured)
```go
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    
    todos := r.Group("/todos")
    {
        todos.GET("", listTodosHandler)
        todos.POST("", createTodoHandler)
        todos.GET("/:id", getTodoHandler)
        todos.PUT("/:id", updateTodoHandler)
        todos.DELETE("/:id", deleteTodoHandler)
    }
    
    r.Run(":8080")
}
```

---

## 17. Add API Documentation (OpenAPI/Swagger) 📚
**Priority**: LOW

### Solution: swaggo
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Annotate Handlers
```go
// @Summary      Create a new todo
// @Description  Add a new todo item to the database
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        todo  body      Todo  true  "Todo object"
// @Success      201   {object}  Todo
// @Failure      400   {object}  AppError
// @Router       /todos [post]
func createTodoHandler(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

### Generate Docs
```bash
swag init -g cmd/api/main.go
```

### Serve Swagger UI
```go
import httpSwagger "github.com/swaggo/http-swagger"

http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
```

Access at: `http://localhost:8080/swagger/index.html`

---

## 18. Add Docker Multi-Stage Build Optimization 🐳
**Priority**: LOW

### Current Issue
Dev Dockerfile doesn't optimize layers.

### Production Dockerfile
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-app cmd/api/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /todo-app .
EXPOSE 8080
CMD ["./todo-app"]
```

---

## Implementation Priority

### Phase 1 (Immediate)
1. ✅ Fix timestamp bug in Update (#4)
2. ✅ Add environment configuration (#9)
3. ✅ Separate layers (#1)
4. ✅ Add context support (#5)

### Phase 2 (Week 1-2)
5. ✅ Use interfaces (#2)
6. ✅ Add validation (#7)
7. ✅ Add GetByID (#3)
8. ✅ Implement error middleware (#8)
9. ✅ Add unit tests (#10)

### Phase 3 (Month 1)
10. ✅ Add logging (#11)
11. ✅ Database migrations (#12)
12. ✅ Replace with Chi/Gin (#16)
13. ✅ Add pagination (#13)

### Phase 4 (Future)
14. Consider sqlx/GORM (#6)
15. API documentation (#17)
16. Health checks (#15)
17. Docker optimization (#18)

---

## Useful Resources
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Testify - Testing toolkit](https://github.com/stretchr/testify)
- [sqlmock - Mock SQL driver](https://github.com/DATA-DOG/go-sqlmock)

---

**Last Updated**: January 10, 2026  
**Maintainer**: Development Team
