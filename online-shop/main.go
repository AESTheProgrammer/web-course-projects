package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	// PostgreSQL connection string
	pgConnStr = ""
	// SQLite connection string
	sqliteConnStr = "sqlite3:basket.db"
	// Create the JWT key used to create the signature
	jwtKey = []byte("my_secret_key")
)

const (
	host     = "localhost"
	port     = 5432
	user     = "aes"
	password = "mysecretpassword"
	dbname   = "aes"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Create a struct that will be encoded to a JWT.
// We add jwt.RegisteredClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Product struct represents the product model
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Basket struct represents the basket model
type Basket struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      string    `json:"data"`
	State     string    `json:"state"`
}

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Cookie   string `json:"cookie"`
}

func main() {
	// InitializeSQLite()
	InitializePostgres()
	// Initialize routers
	r := mux.NewRouter()

	// Endpoints
	r.HandleFunc("/basket/", GetBaskets).Methods("GET")
	r.HandleFunc("/basket/", CreateBasket).Methods("POST")
	r.HandleFunc("/basket/{id}", UpdateBasket).Methods("PATCH")
	r.HandleFunc("/basket/{id}", GetBasket).Methods("GET")
	r.HandleFunc("/basket/{id}", DeleteBasket).Methods("DELETE")
	r.HandleFunc("/signin", Signin).Methods("POST")
	r.HandleFunc("/logout", Logout).Methods("POST")
	r.HandleFunc("/refresh", Refresh).Methods("GET")
	r.HandleFunc("/signup", Signup).Methods("POST")

	// Start server
	log.Fatal(http.ListenAndServe(":8080", r))
}

// SignUp creates a new user and returns a JWT token
func Signup(w http.ResponseWriter, r *http.Request) {
	var newUser User
	// Decode the JSON request body into newUser
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid request payload")
		return
	}
	// Validate input (e.g., check if username and password meet criteria)
	// Hash the password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error hashing password: %v", err)
		return
	}
	// Open PostgreSQL database connection
	pgdb, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error connecting to the database: %v", err)
		return
	}
	defer pgdb.Close()

	// Insert the new user into the 'users' table
	err = pgdb.QueryRow("INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", newUser.Username, string(hashedPassword)).Scan(&newUser.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating user: %v", err)
		return
	}
	// Generate JWT token
	tokenString, expirationTime, err := generateToken(newUser.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error generating JWT token: %v", err)
		return
	}
	// w.WriteHeader(http.StatusCreated)
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// generateToken generates a JWT token for the given username
func generateToken(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	// if err != nil {
	// 	// If there is an error in creating the JWT return an internal server error
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	return tokenString, expirationTime, err
}

// Create the Signin handler
func Signin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Get the expected password from our in memory map
	// expectedPassword, ok := users[creds.Username]
	println(creds.Password)
	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	expectedPassword, err := GetPasswordByUsername(creds.Username)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(expectedPassword), []byte(creds.Password))
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// // Create the JWT string
	tokenString, expirationTime, err := generateToken(creds.Username)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// SetUserCookie(creds.Username, tokenString)
	// Finally, we set the client cookie for "token" as the JWT we just generated
	// we also set an expiry time which is the same as the token itself
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func Authenticate(w http.ResponseWriter, r *http.Request) string {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("token")
	if err != nil {
		fmt.Println(err.Error())
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return ""
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return ""
	}
	// Get the JWT string from the cookie
	tknStr := c.Value
	// Initialize a new instance of `Claims`
	claims := &Claims{}
	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		fmt.Println(err.Error())
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return ""
		}
		w.WriteHeader(http.StatusBadRequest)
		return ""
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return ""
	}
	return claims.Username
}

// SetUserCookie updates the 'cookie' column for a specific user in PostgreSQL
func SetUserCookie(username string, cookieValue string) error {
	// Open PostgreSQL database connection
	pgdb, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return err
	}
	defer pgdb.Close()
	// Update the 'cookie' column for the specified user
	_, err = pgdb.Exec("UPDATE users SET cookie = $1 WHERE username = $2", cookieValue, username)
	if err != nil {
		return err
	}
	fmt.Println("Cookie updated successfully.")
	return nil
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	// (BEGIN) The code until this point is the same as the first part of the `Welcome` route
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tknStr := c.Value
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		fmt.Println("Error:", err.Error())
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	fmt.Println(time.Until(claims.ExpiresAt.Time))
	if time.Until(claims.ExpiresAt.Time) > 120*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Now, create a new token for the current use, with a renewed expiration time
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error:", err.Error())
		return
	}
	// SetUserCookie(claims.Username, tokenString)
	// Set the new token as the users `token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
}

func GetPasswordByUsername(username string) (string, error) {
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return "", err
	}
	defer db.Close()
	var user User
	err = db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&user.Password)
	if err != nil {
		return "", err
	}
	return user.Password, nil
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// immediately clear the token cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
	})
}

// GetBaskets retrieves a list of baskets
func GetBaskets(w http.ResponseWriter, r *http.Request) {
	username := Authenticate(w, r)
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Query to get basket IDs from SQLite database
	rows, err := db.Query("SELECT id FROM baskets where username = $1", username)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	// Slice to store basket IDs
	var basketIDs []int
	// Iterate through the result set and append basket IDs to the slice
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
		basketIDs = append(basketIDs, id)
	}
	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	// Send JSON response with basket IDs
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(basketIDs)
	fmt.Fprintf(w, "GET /basket/ - List of baskets")
}

// CreateBasket creates a new basket
func CreateBasket(w http.ResponseWriter, r *http.Request) {
	username := Authenticate(w, r)
	if username == "" {
		return
	}
	// Open SQLite database connection
	// db, err := sql.Open("sqlite3", sqliteConnStr)
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Create a new Basket object
	newBasket := Basket{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Data:      "{}",
		State:     "PENDING",
	}
	// Insert the new basket into SQLite database
	err = db.QueryRow(
		"INSERT INTO baskets (username, created_at, updated_at, data, state) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		username, newBasket.CreatedAt, newBasket.UpdatedAt, newBasket.Data, newBasket.State).Scan(&newBasket.ID)
	if err != nil {
		log.Fatal(err)
	}
	// Send JSON response with the ID of the new basket
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": newBasket.ID})
	fmt.Fprintf(w, "POST /basket/ - New basket created")
}

// UpdateBasket updates a given basket
func UpdateBasket(w http.ResponseWriter, r *http.Request) {
	username := Authenticate(w, r)
	if username == "" {
		return
	}
	// Open PostgreSQL database connection
	pgdb, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()
	// Extract basket ID from URL
	vars := mux.Vars(r)
	basketID := vars["id"]
	// Extract products_name and count from the request body
	var updateData struct {
		State        string `json:"state"`
		ProductsName string `json:"product_name"`
		Count        int    `json:"count"`
	}
	// Decode the request body into the updateData struct
	err = json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error decoding request body: %v", err)
		return
	}

	if updateData.State != "COMPLETED" && updateData.State != "PENDING" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "status must be either COMPLETED or PENDING not %v\n", updateData.State)
		return
	}
	// Convert updateData to a map for the JSON update
	updateMap := map[string]interface{}{updateData.ProductsName: updateData.Count}
	// Convert the map to a JSON string
	updateJSON, err := json.Marshal(updateMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error encoding update data to JSON: %v", err)
		return
	}
	// Execute the UPDATE query to update the JSON field in PostgreSQL
	tmp := pq.QuoteIdentifier(updateData.ProductsName)
	tmp = fmt.Sprintf(`UPDATE baskets SET
	data = jsonb_set(data::jsonb, '{%v}', to_jsonb($1::int)), state = $4
	WHERE id = $2 AND username = $3 AND state <> 'COMPLETED'`, tmp)
	res, err := pgdb.Exec(tmp, updateData.Count, basketID, username, updateData.State)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error updating basket: %v", err)
		log.Fatal(err)
	}
	changes, _ := res.RowsAffected()
	if changes == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotModified)
		fmt.Fprintf(w, "Error: either basket %v was not found or basket was completed and therefore not modifiable.", basketID)
		return
	}
	// Send JSON response indicating successful update
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Basket with ID %s updated with %s", basketID, updateJSON)
}

// GetBasket retrieves a given basket
func GetBasket(w http.ResponseWriter, r *http.Request) {
	// Open SQLite database connection
	// db, err := sql.Open("sqlite3", sqliteConnStr)
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Extract basket ID from URL
	vars := mux.Vars(r)
	basketID := vars["id"]

	// Query to get basket information from SQLite database
	// row := db.QueryRow("SELECT ID, CreatedAt, UpdatedAt, Data, State FROM baskets WHERE ID = ?", basketID)
	row := db.QueryRow("SELECT id, created_at, updated_at, data, state FROM baskets WHERE ID = $1", basketID)

	// Create a Basket object to store the retrieved information
	var basket Basket

	// Scan the row into the Basket object
	err = row.Scan(&basket.ID, &basket.CreatedAt, &basket.UpdatedAt, &basket.Data, &basket.State)
	if err != nil {
		// Handle the case where the basket with the given ID is not found
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Basket with ID %s not found", basketID)
			return
		}
		log.Fatal(err)
	}

	// Send JSON response with basket information
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(basket)
	fmt.Fprintf(w, "GET /basket/%s - Retrieve basket", basketID)
}

// DeleteBasket deletes a given basket
func DeleteBasket(w http.ResponseWriter, r *http.Request) {
	username := Authenticate(w, r)
	// db, err := sql.Open("sqlite3", sqliteConnStr)
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Extract basket ID from URL
	vars := mux.Vars(r)
	basketID := vars["id"]
	// Execute the DELETE query to remove the basket with the specified ID
	result, err := db.Exec("DELETE FROM baskets WHERE id = $1 AND username = $2", basketID, username)
	if err != nil {
		log.Fatal(err)
	}
	// Check if the basket with the given ID was found and deleted
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Basket with ID %s not found", basketID)
		return
	}

	// Send JSON response indicating successful deletion
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Basket with ID %s deleted", basketID)
}

func InitializePostgres() {
	// Connect to the PostgreSQL database
	pgConnStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create 'baskets' table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS baskets (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			data JSON NOT NULL,
			state VARCHAR(255) NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func InitializeSQLite() {
	// Open the SQLite database (it will be created if it doesn't exist)
	db, err := sql.Open("sqlite3", sqliteConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS baskets (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			CreatedAt DATETIME,
			UpdatedAt DATETIME,
			Data TEXT,
			State TEXT CHECK(State IN ('COMPLETED', 'PENDING'))
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Table created successfully")
}
