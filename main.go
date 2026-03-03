package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Users struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var mySecretKey = []byte("super-secret-password")

var dbPool *pgxpool.Pool

func initDB() {

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:pass@localhost:5432/godb?sslmode=disable"
	}

	// Retry configuration (important for Docker startup race)
	maxRetries := 5
	retryDelay := 2 * time.Second

	var err error

	for i := 1; i <= maxRetries; i++ {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		dbPool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			err = dbPool.Ping(ctx)
		}

		cancel()

		if err == nil {
			fmt.Println("Connected to PostgreSQL")
			break
		}

		fmt.Printf("Database not ready (attempt %d/%d): %v\n", i, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatal("Could not connect to database after retries:", err)
	}

	// Optional: Tune connection pool (production-friendly defaults)
	dbPool.Config().MaxConns = 10
	dbPool.Config().MinConns = 2
	dbPool.Config().MaxConnLifetime = time.Hour
	dbPool.Config().MaxConnIdleTime = 30 * time.Minute

	// Create table
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL
	);
	`

	_, err = dbPool.Exec(ctx, createTableSQL)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	fmt.Println("Users table ready.")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("URL: %s | Method: %s", r.URL.Path, r.Method)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")

		if user != "admin" {
			http.Error(w, "Please login as admin (?user=admin)", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader != "Bearer valid-token-123" {
			http.Error(w, "Unauthorized: Token missing or invalid", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to Home Page!")
}

func sendEmail(ctx context.Context, email string, wg *sync.WaitGroup, statusChain chan string) {
	defer wg.Done()

	select {
	case <-time.After(1 * time.Second):
		statusChain <- "Success: Sent to " + email
	case <-ctx.Done():
		statusChain <- "Timeout: Could not send to " + email
	}
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("empty email")
	}
	return nil
}

func generateToken() {
	fmt.Println("Token generated!")
}

func UserHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/favicon.ico" {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {

	// GET ALL USERS
	case http.MethodGet:
		fmt.Println("Fetching all users")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := dbPool.Query(ctx, "SELECT id, name, email FROM users")
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Users])
		if err != nil {
			http.Error(w, "Failed to read data", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(users)

	// CREATE USER
	case http.MethodPost:

		var newUser Users

		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := validateEmail(newUser.Email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		statusChain := make(chan string)
		var wg sync.WaitGroup

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"

		err := dbPool.QueryRow(
			ctx,
			query,
			newUser.Name,
			newUser.Email,
		).Scan(&newUser.ID)

		if err != nil {
			http.Error(w, "Insert failed (maybe duplicate email)", http.StatusInternalServerError)
			return
		}

		fmt.Println("Signup process started...")
		fmt.Println("User created in Database.")

		wg.Add(1)
		go sendEmail(ctx, newUser.Email, &wg, statusChain)

		generateToken()
		fmt.Println("User is now looking at the Dashboard.")

		go func() {
			wg.Wait()
			close(statusChain)
		}()

		var reports []string
		for rpt := range statusChain {
			reports = append(reports, rpt)
		}

		fmt.Println("All background tasks finished.")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User Created Successfully & Login Success!",
			"token":   "valid-token-123",
			"user":    newUser,
			"reports": reports,
		})
		return

	// DELETE USER
	case http.MethodDelete:

		idParam := r.URL.Query().Get("id")
		if idParam == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		commandTag, err := dbPool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)

		if err != nil {
			http.Error(w, "Delete failed", http.StatusInternalServerError)
			return
		}

		if commandTag.RowsAffected() == 0 {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message": "User deleted successfully",
		})
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

}
func main() {

	initDB()
	defer dbPool.Close()

	mux := http.NewServeMux()

	mux.Handle("/signup", loggingMiddleware(http.HandlerFunc(UserHandler)))
	mux.Handle("/dashboard", jwtMiddleware(http.HandlerFunc(homeHandler)))

	finalHandler := http.HandlerFunc(homeHandler)
	mux.Handle("/", loggingMiddleware(authMiddleware(finalHandler)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("Server running at http://localhost:8080/signup")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Listen error:", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("Shutdown error:", err)
	}

	fmt.Println("Server gracefully stopped")
}
