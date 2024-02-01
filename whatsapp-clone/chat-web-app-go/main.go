package main

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "strconv"
  "time"
  "sync"
  "os"
  "io"
	"io/ioutil"
	"path/filepath"
  "math/rand"

  "github.com/golang-jwt/jwt/v5"
  "github.com/gorilla/websocket"
  "github.com/gorilla/mux"
  "github.com/lib/pq"
  _ "github.com/mattn/go-sqlite3"
  "golang.org/x/crypto/bcrypt"
)


type ClientList map[int][]websocket.Conn

var (
  // PostgreSQL connection string
  pgConnStr = ""
  // Create the JWT key used to create the signature
  jwtKey = []byte("my_secret_key")
  clients = make(map[int][]websocket.Conn)
  mu       sync.Mutex
  upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
  }
  letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

const (
  host     = "localhost"
  port     = 5432
  user     = "aes"
  password = "mysecretpassword"
  dbname   = "aes"
  charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Create a struct to read the username and password from the request body
type Credentials struct {
  Password string `json:"password"`
  Username string `json:"username"`
}

type Claims struct {
  Username string `json:"username"`
  jwt.RegisteredClaims
}

type Contact struct {
  UserID      int    `json:"user_id,omitempty"`
  ContactID   int    `json:"contact_id"`
  ContactName string `json:"contact_name"`
}

type Manager struct {
  clients ClientList
  sync.RWMutex
}


type Chat struct {
  ChatID    int       `json:"chat_id,omitempty"`
  People    []int     `json:"people"`
  CreatedAt time.Time `json:"created_at,omitempty"`
}

type Group struct {
  GroupID   int       `json:"group_id,omitempty"`
  People    []int     `json:"people"`
  CreatedAt time.Time `json:"created_at,omitempty"`
}

type Extension struct {
  ChatID      int       `json:"chat_id"`
  NotViewed   int       `json:"notviewed"`
  LastUser    int       `json:"last_user"`
  LastMessage string    `json:"last_message"`
  LastMsgDate time.Time `json:"last_msg_date"`
  LastOn1     time.Time `json:"laston1"`
  LastOn2     time.Time `json:"laston2"`
}

type User struct {
  ID        int    `json:"id,omitempty"`
  Firstname string `json:"firstname"`
  Lastname  string `json:"lastname"`
  Phone     string `json:"phone"`
  Username  string `json:"username"`
  Password  string `json:"password"`
  Image     string `json:"image,omitempty"`
  ImgBytes  []byte `json:"image_bytes,omitempty"`
  Bio       string `json:"bio"`
}

type UpdateUser struct {
  Firstname string `json:"firstname,omitempty"`
  Lastname  string `json:"lastname,omitempty"`
  Phone     string `json:"phone,omitempty"`
  Username  string `json:"username,omitempty"`
  //Password  string `json:"password,omitempty"`
  Image     []byte `json:"image,omitempty"`
  Bio       string `json:"bio,omitempty"`
}

type Message struct {
  ID        int       `json:"id"`
  ChatID    int       `json:"chat_id"`
  Sender    int       `json:"sender"`
  Receiver  int       `json:"receiver"`
  Content   string    `json:"content"`
  CreatedAt time.Time `json:"created_at"`
}

func main() {
  InitializePostgres()
  // Initialize routers
  r := mux.NewRouter()
  // Group Endpoints
  // r.HandleFunc("/api/groups", CreateGroup).Methods("POST")
  // r.HandleFunc("/api/groups/{group_id}", DeleteGroup).Methods("DELETE")
  // r.HandleFunc("/api/groups/{group_id}", GetGroupContent).Methods("GET")
  // r.HandleFunc("/api/groups/{group_id}", AddUser).Methods("PATCH")
  // r.HandleFunc("/api/groups/{group_id}/{user_id}", RemoveUser).Methods("DELETE")
  // // Chat Endpoints
  r.HandleFunc("/api/chats", ListChats).Methods("GET")
  r.HandleFunc("/api/chats", CreateChat).Methods("POST")
  r.HandleFunc("/api/chats/{chat_id}", GetChatContent).Methods("GET")
  r.HandleFunc("/api/extensions/{chat_id}", GetExtension).Methods("GET")
  r.HandleFunc("/api/chats/{chat_id}", DeleteChat).Methods("DELETE")
  r.HandleFunc("/api/chats/{chat_id}/messages/{message_id}", DeleteMessage).Methods("DELETE")
  // // List of Contacts Endpoints
  r.HandleFunc("/api/users/{user_id}/contacts", ListUserContacts).Methods("GET")
  r.HandleFunc("/api/users/{user_id}/contacts", AddContact).Methods("POST")
  r.HandleFunc("/api/users/{user_id}/contacts/{contact_id}", DeleteContact).Methods("DELETE")
  // List of Users Endpoints
  r.HandleFunc("/api/users/{user_id}", GetUserInfo).Methods("GET")
  r.HandleFunc("/api/users/{user_id}", UpdateUserInfo).Methods("PATCH")
  r.HandleFunc("/api/users/{user_id}", DeleteUser).Methods("DELETE")
  r.HandleFunc("/api/users", GetUserInfoByKeyword).Methods("GET") // .Queries("keyword", "")
  // Authentication Endpoints
  r.HandleFunc("/api/login", Login).Methods("POST")
  r.HandleFunc("/api/logout", Logout).Methods("GET")
  r.HandleFunc("/api/refresh", Refresh).Methods("GET")
  r.HandleFunc("/api/register", Register).Methods("POST")
  log.Fatal(http.ListenAndServe(":8080", r))
}

func Register(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB maximum
    http.Error(w, "Error parsing form data", http.StatusBadRequest)
    return
  }
  file, _, err := r.FormFile("image")
  if err != nil {
    http.Error(w, "Error retrieving file", http.StatusBadRequest)
    return
  }
  defer file.Close()
  newImageName := generateRandomString(10) + ".jpeg"
  fmt.Printf("Register: newImageName=%v\n", newImageName)
  outFile, err := os.Create("./photos/" + newImageName)
  if err != nil {
    http.Error(w, "Error creating file", http.StatusInternalServerError)
    return
  }
  _, err = io.Copy(outFile, file)
  if err != nil {
    http.Error(w, "Error copying file", http.StatusInternalServerError)
    return
  }
  defer outFile.Close()
  var newUser = User {
    ID:         -1,
    Firstname:  r.FormValue("firstname"),
    Lastname:   r.FormValue("lastname"),
    Phone:      r.FormValue("phone"),
    Username:   r.FormValue("username"),
    Password:   r.FormValue("password"),
    Image:      newImageName,
    Bio:        r.FormValue("Bio"), 
  }
  if newUser.Firstname == "" || newUser.Lastname == "" || newUser.Phone == "" ||
  newUser.Username == "" || newUser.Password == "" || newUser.Bio == "" || newUser.Image == "" {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Invalid request payload")
    return
  }
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
  err = pgdb.QueryRow(`INSERT INTO users (firstname, lastname, phone, username, password, image, bio) 
    VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
    newUser.Firstname, newUser.Lastname, newUser.Phone,
    newUser.Username, string(hashedPassword), newUser.Image, newUser.Bio).Scan(&newUser.ID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Error creating user: %v", err)
    return
  }
  // Generate JWT token
  tokenString, expirationTime, err := generateToken(newUser.ID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Error generating JWT token: %v", err)
    return
  }
  http.SetCookie(w, &http.Cookie{
    Name:    "token",
    Value:   tokenString,
    Expires: expirationTime,
  })
  w.WriteHeader(http.StatusCreated)
  fmt.Fprintf(w, `{"userid": %v}`, newUser.ID)
}

func generateToken(userID int) (string, time.Time, error) {
  expirationTime := time.Now().Add(30 * time.Minute)
  // Create the JWT claims, which includes the username and expiry time
  claims := &Claims{
    Username: strconv.Itoa(userID),
    RegisteredClaims: jwt.RegisteredClaims{
      // In JWT, the expiry time is expressed as unix milliseconds
      ExpiresAt: jwt.NewNumericDate(expirationTime),
    },
  }
  // Declare the token with the algorithm used for signing, and the claims
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  // Create the JWT string
  tokenString, err := token.SignedString(jwtKey)
  return tokenString, expirationTime, err
}

func Login(w http.ResponseWriter, r *http.Request) {
  // Get the JSON body and decode into credentials
  var creds Credentials
  err := json.NewDecoder(r.Body).Decode(&creds)
  if err != nil {
    // If the structure of the body is wrong, return an HTTP error
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  // Get the expected password from our in memory map
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
  id, err := GetUsernameByID(creds.Username)
  if err != nil {
    fmt.Println(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  tokenString, expirationTime, err := generateToken(id)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  http.SetCookie(w, &http.Cookie{
    Name:    "token",
    Value:   tokenString,
    Expires: expirationTime,
  })
  w.Header().Set("Content-Type", "application/json")
  fmt.Fprintf(w, `{"userid": %v}`, id)
  //fmt.Fprintf(w, "POST /api/login - login user %s using username and password.", creds.Username)
}

func GetUsernameByID(username string) (int, error) {
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    return -1, err
  }
  defer pgdb.Close()
  var user User
  err = pgdb.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&user.ID)
  if err != nil {
    return -1, err
  }
  return user.ID, nil
}

func GetIDFromCookie(w http.ResponseWriter, r *http.Request) int {
  c, err := r.Cookie("token")
  if err != nil {
    fmt.Println(err.Error())
    if err == http.ErrNoCookie {
      w.WriteHeader(http.StatusUnauthorized)
      return -1
    }
    w.WriteHeader(http.StatusBadRequest)
    return -1
  }
  tknStr := c.Value
  claims := &Claims{}
  tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (any, error) {
    return jwtKey, nil
  })
  if err != nil {
    fmt.Println(err.Error())
    if err == jwt.ErrSignatureInvalid {
      w.WriteHeader(http.StatusUnauthorized)
      return -1
    }
    w.WriteHeader(http.StatusBadRequest)
    return -1
  }
  if !tkn.Valid {
    w.WriteHeader(http.StatusUnauthorized)
    return -1
  }
  id, err := strconv.Atoi(claims.Username)
  if err != nil {
    fmt.Printf("Could not Convert %v to int\n", claims.Username)
    fmt.Println(err)
  }
  return id
}

func Refresh(w http.ResponseWriter, r *http.Request) {
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
  http.SetCookie(w, &http.Cookie{
    Name:    "token",
    Value:   tokenString,
    Expires: expirationTime,
  })
}

func GetPasswordByUsername(username string) (string, error) {
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    return "", err
  }
  defer pgdb.Close()
  var user User
  err = pgdb.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&user.Password)
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

func CreateGroup(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  if userID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  var newGroup Group
  err := json.NewDecoder(r.Body).Decode(&newGroup)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Invalid request payload")
    fmt.Println(err)
    return
  }
  if newGroup.People[0] != userID {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "First user should be the creator of the chat!")
    return
  }
  if len(newGroup.People) < 2 {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Atleast two people should be in a group.")
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  _, err = pgdb.Query(
    "INSERT INTO groups (people, created_at) VALUES ($1, $2)",
    pq.Array(newGroup.People), time.Now())
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  // json.NewEncoder(w).Encode(map[string]int64{"id": newBasket.ID})
  fmt.Fprintf(w, "POST /api/{user_id}/chat - New chat created")
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  if userID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  groupID := vars["group_id"]
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Unable to connect to DB!\n")
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  result, err := pgdb.Exec(`
    DELETE 	FROM groups
    WHERE 	group_id = $1 AND 
    (people[1] = $2)`,
    groupID, userID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Unable to execute query!\n")
    fmt.Println(err)
    return
  }
  rowsAffected, _ := result.RowsAffected()
  if rowsAffected == 0 {
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintf(w, "group was not found.")
    return
  }
  w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "group with ID %v from user with ID %v deleted", groupID, userID)
}


func GetChatContent(w http.ResponseWriter, r *http.Request) {
  log.Printf("User connected\n")
  userID := GetIDFromCookie(w, r)
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  vars := mux.Vars(r)
  chatID, err := strconv.Atoi(vars["chat_id"])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprintf(w, "{chat_id} must be an int!\n")
    fmt.Println(err)
    return
  }
  var receiver int
  if receiver = IsInChat(userID, chatID); receiver == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "You Don't have access to this chat!\n")
    fmt.Println(err)
    return
  }
  var messages []Message
  err = GetMessagesFromDB(&messages, chatID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Printf("Error while getting messages from DB:: %v\n", err)
    return
  }
  upgrader.CheckOrigin = func(r *http.Request) bool { return true }
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Printf("Failed to make connection: %v", err)
    return
  }
  defer func() {
    mu.Lock()
    defer mu.Unlock()
    conn.Close()
    if _, ok := clients[chatID]; ok {
      for i, c := range clients[chatID] {
        if &c == conn {
          clients[chatID] = remove(clients[chatID], i)
          break
        }
      }
    }
  }()
  mu.Lock()
  if value, ok := clients[chatID]; ok {
    clients[chatID] = append(value, *conn)
  } else {
    newSlice := make([]websocket.Conn, 1)
    newSlice[0] = *conn
    clients[chatID] = newSlice
  }
  mu.Unlock()
  for _, message := range messages {
    var content = 
      strconv.Itoa(message.Sender) + 
      "|" +
      message.CreatedAt.Format("2006-01-02 15:04:05") +
      "|" +
      message.Content 
    conn.WriteMessage(websocket.TextMessage, []byte(content))
    fmt.Println(content)
  }

  //_, err = UpdateChatExtension(chatID, userID, "")
  err = UpdateChatExtension(chatID, userID, "")
  if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return 
  }
  fmt.Println("sent messages to client")
  for {
    fmt.Println("reading message")
    msgType, msg, err := conn.ReadMessage()
    if err != nil {
      if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
        log.Printf("error reading message: %v", err)
      }
      break // Break the loop to close conn & Cleanup
    }
    //var content = ""
    //if (len(string(msg)) != 0) {
    var content = 
      strconv.Itoa(userID) + 
      "|" +
      time.Now().Format("2006-01-02 15:04:05") +
      "|" +
      string(msg) + 
      "|f"
    if len(clients[chatID]) > 1 {
      content += "o"
    } 
    //}
    log.Printf("Number of Open Sockets: %v\n", len(clients[chatID]))
    fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
    for _, c := range clients[chatID] {
      if err = c.WriteMessage(msgType, []byte(content)); err != nil {
        log.Println("connection closed: ", err)
        err = c.WriteMessage(websocket.CloseMessage, nil)
        if &c == conn {
          return
        }
      }
      fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
    }
    err = UpdateChatExtension(chatID, userID, string(msg))
    //_, err = UpdateChatExtension(chatID, userID, string(msg))
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return 
    }
    _, err = pgdb.Query(
      `INSERT INTO messages
      (chat_id, sender, receiver, content, created_at)
      VALUES ($1, $2, $3, $4, $5)`,
      chatID, userID, receiver, msg, time.Now())
    if err != nil {
      log.Println("connection closed: ", err)
      err = conn.WriteMessage(websocket.CloseMessage, nil)
    }
  }
}

func UpdateChatExtension(chatID int, userID int, msg string) error { // (Extension, error) {
  pgdb, err := sql.Open("postgres", pgConnStr)
  var ext Extension
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer pgdb.Close()
  var people pq.Int32Array
  err = pgdb.QueryRow(`
    SELECT people
    FROM chats
    WHERE chat_id = $1
    AND $2 = ANY(people)`,
    chatID, userID).Scan(&people)
  if err != nil {
    fmt.Printf("UpdateChatExtension: %v\n", err)
    return err
  }
  var curr_time = time.Now()
  ext, err = GetExtensionFromDB(chatID)
  if err != nil {
    fmt.Printf("UpdateChatExtension: %v\n", err)
    return err
  }
  if ext.LastUser != userID {
    ext.NotViewed = 0
  }
  if len(msg) != 0 {
    ext.LastMessage = msg
    ext.LastMsgDate = curr_time
    ext.LastUser = userID 
    ext.NotViewed += 1
  }
  if int(people[0]) == userID {
    ext.LastOn1 = curr_time
  } else {
    ext.LastOn2 = curr_time
  }
  stmt, err := pgdb.Prepare(
    `Update extensions SET notviewed=$1, last_user=$2,
    last_message=$3,  last_msg_date=$4, laston1=$5, laston2=$6 WHERE chat_id = $7`)
  if err != nil {
    fmt.Printf("UpdateChatExtension: %v\n", err)
    return err
  }
  fmt.Printf("UpdateChatExtension: %v", ext)
  defer stmt.Close()
  _, err = stmt.Exec(ext.NotViewed, ext.LastUser,
    ext.LastMessage, ext.LastMsgDate, ext.LastOn1, ext.LastOn2, chatID)
  if err != nil {
    fmt.Printf("UpdateChatExtension: %v\n", err)
    return err
  }
  return nil
}

func GetExtensionFromDB(chatID int) (Extension, error) {
  pgdb, err := sql.Open("postgres", pgConnStr)
  var ext Extension
  if err != nil {
    fmt.Println(err)
    return ext, err
  }
  defer pgdb.Close()
  row := pgdb.QueryRow(`SELECT 
    notviewed, last_user, last_message, last_msg_date, laston1, laston2
    FROM extensions 
    WHERE chat_id = $1`, chatID)
  err = row.Scan(&ext.NotViewed, &ext.LastUser,
    &ext.LastMessage, &ext.LastMsgDate, &ext.LastOn1, &ext.LastOn2)
  if err != nil {
    fmt.Printf("getExtensionFromDB: %v\n", err)
    return ext, err
  }
  return ext, nil
}


func GetExtension(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  vars := mux.Vars(r)
  chatID, err := strconv.Atoi(vars["chat_id"])
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprintf(w, "{chat_id} must be an int!\n")
    fmt.Println(err)
    return
  }
  var receiver int
  if receiver = IsInChat(userID, chatID); receiver == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "You Don't have access to this chat!\n")
    fmt.Println(err)
    return
  }
  ext, err := GetExtensionFromDB(chatID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(ext)
}

func GetMessagesFromDB(messages *[]Message, chatID int) error {
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    return err
  }
  rows, err := pgdb.Query(`
    SELECT chat_id, sender, receiver, content, created_at  
    FROM messages WHERE chat_id = $1`,
    chatID)
  if err != nil {
    return err
  }
  fmt.Println("getting chat content")
  defer rows.Close()
  for rows.Next() {
    var message Message
    err := rows.Scan(&message.ChatID, &message.Sender, &message.Receiver,
      &message.Content, &message.CreatedAt)
    if err != nil {
      fmt.Printf("Scanning Error:: %v \n", err)
      return err
    }
    *messages = append(*messages, message)
  }
  return nil
}

func remove(s []websocket.Conn, i int) []websocket.Conn {
  s[i] = s[len(s)-1]
  return s[:len(s)-1]
}


func DeleteMessage(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  if userID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  chatID := vars["chat_id"]
  messageID := vars["message_id"]
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Unable to connect to DB!\n")
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  result, err := pgdb.Exec(`
    DELETE 	FROM messages
    WHERE 	chat_id = $1 AND
    message_id = $2
    sender = $3`,
    chatID, messageID, userID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Unable to execute query!\n")
    fmt.Println(err)
    return
  }
  rowsAffected, _ := result.RowsAffected()
  if rowsAffected == 0 {
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintf(w, "message was not found.")
    return
  }
  w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "chat with ID %v from user with ID %v deleted", chatID, userID)
}

func IsInChat(userID int, chatID int) int {
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    log.Fatalf("DoesUserBelongChat: %v", err)
  }
  fmt.Println("Checking if User belongs to the chat.")
  defer pgdb.Close()
  var people pq.Int32Array
  err = pgdb.QueryRow(`
    SELECT people
    FROM chats
    WHERE chat_id = $1
    AND $2 = ANY(people)`,
    chatID, userID).Scan(&people)
  if err != nil {
    if err == sql.ErrNoRows {
      fmt.Println("No matching rows found.")
      return -1
    } else {
      log.Fatal(err)
    }
  }
  if int(people[0]) != userID {
    return int(people[0])
  }
  return int(people[1])
}

func ListChats(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  rows, err := pgdb.Query(`
    SELECT chat_id, people, created_at 
    FROM chats WHERE people[1] = $1 OR 
    people[2] = $1`,
    userID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  defer rows.Close()
  var chats []Chat
  for rows.Next() {
    var chat Chat
    var intArray pq.Int64Array
    err := rows.Scan(&chat.ChatID, &intArray, &chat.CreatedAt)
    for _, v := range intArray {
      chat.People = append(chat.People, int(v))
    }
    if err != nil {
      fmt.Println(err)
      return
    }
    chats = append(chats, chat)
  }
  if err := rows.Err(); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  responseJSON, err := json.Marshal(chats)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  w.Write(responseJSON)
}

func CreateChat(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  if userID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  var newChat Chat
  err := json.NewDecoder(r.Body).Decode(&newChat)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Invalid request payload")
    fmt.Println(err)
    return
  }
  if newChat.People[0] != userID {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "First user should be the creator of the chat!")
    return
  }
  if len(newChat.People) != 2 {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Atleast two people should be in a chat.")
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    fmt.Println(err)
    return
  }
  var newID int
  defer pgdb.Close()
  var curr_time = time.Now()
  err = pgdb.QueryRow(
    "INSERT INTO chats (people, created_at) VALUES ($1, $2) RETURNING chat_id",
    pq.Array(newChat.People), curr_time).Scan(&newID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  _, err = pgdb.Exec(
    `INSERT INTO extensions 
    (chat_id, notviewed, last_user, last_message, last_msg_date, laston1, laston2) 
    VALUES ($1, $2, $3, $4, $5, $6, $7)`,
    newID, 1, -1, "Hi, let's chat!", curr_time, curr_time, curr_time)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Printf("CreateChat: %v\n", err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  fmt.Fprintf(w, "%v", newID)
}

func DeleteChat(w http.ResponseWriter, r *http.Request) {
  userID := GetIDFromCookie(w, r)
  if userID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  chatID := vars["chat_id"]
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Unable to connect to DB!\n")
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  result, err := pgdb.Exec(`
    DELETE 	FROM chats 
    WHERE 	chat_id = $1 AND 
    (people[1] = $2 OR people[2] = $2)`,
    chatID, userID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Unable to execute query!\n")
    fmt.Println(err)
    return
  }
  rowsAffected, _ := result.RowsAffected()
  if rowsAffected == 0 {
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintf(w, "chat was not found.")
    return
  }
  w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "chat with ID %v from user with ID %v deleted", chatID, userID)
}

func ListUserContacts(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  vars := mux.Vars(r)
  userID := vars["user_id"]
  if strconv.Itoa(expectedID) != userID {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "Unauthorized access!\n")
    fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  rows, err := pgdb.Query("SELECT contact_id, contact_name FROM contacts where user_id = $1",
    userID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  defer rows.Close()
  var contacts []Contact
  for rows.Next() {
    var contact Contact
    err := rows.Scan(&contact.ContactID, &contact.ContactName)
    if err != nil {
      fmt.Println(err)
      return
    }
    contacts = append(contacts, contact)
  }
  if err := rows.Err(); err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  // json.NewEncoder(w).Encode(basketIDs)
  // fmt.Fprintf(w, "GET /api/user/{user_id} - List of baskets")
  responseJSON, err := json.Marshal(contacts)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  w.Write(responseJSON)
}

func AddContact(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  if expectedID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  userID := vars["user_id"]
  if strconv.Itoa(expectedID) != userID {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "Unauthorized access!\n")
    fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
    return
  }
  var newContact Contact
  err := json.NewDecoder(r.Body).Decode(&newContact)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Invalid request payload")
    fmt.Println(err)
    return
  }
  if strconv.Itoa(newContact.ContactID) == userID {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprintf(w, "Can't add yourself to your contacts!\n")
    fmt.Printf("contact_id is equal to user_id: %v", userID)
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    fmt.Println(err)
    return
  }
  defer pgdb.Close()
  _, err = pgdb.Query(
    "INSERT INTO contacts (user_id, contact_id, contact_name) VALUES ($1, $2, $3)",
    userID, newContact.ContactID, newContact.ContactName)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Println(err)
    return
  }
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  // json.NewEncoder(w).Encode(map[string]int64{"id": newBasket.ID})
  fmt.Fprintf(w, "POST /api/{user_id}/contact - New contact created")
}

func DeleteContact(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  if expectedID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  userID := vars["user_id"]
  if strconv.Itoa(expectedID) != userID {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "Unauthorized access!\n")
    fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    log.Fatal(err)
  }
  defer pgdb.Close()
  contactID := vars["contact_id"]
  result, err := pgdb.Exec("DELETE FROM contacts WHERE contact_id = $1 AND user_id = $2", contactID, userID)
  if err != nil {
    log.Fatal(err)
  }
  rowsAffected, _ := result.RowsAffected()
  if rowsAffected == 0 {
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintf(w, "Contact was not found or your account was not found.")
    return
  }
  w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "contact with ID %s from user with ID %s deleted", contactID, userID)
}

func UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  if expectedID < 0 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  userID := vars["user_id"]
  if strconv.Itoa(expectedID) != userID {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "Unauthorized access!\n")
    fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
  }

  if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB maximum
    http.Error(w, "Error parsing form data", http.StatusBadRequest)
    return
  }

  file, _, err := r.FormFile("image")
  if err != nil {
    fmt.Printf("Error retrieving image: %v!\n", err)
    http.Error(w, "Error retrieving file", http.StatusBadRequest)
    return
  }
  defer file.Close()
  newImageName := generateRandomString(10) + ".jpeg"
  outFile, err := os.Create("./photos/" + newImageName)
  if err != nil {
    fmt.Printf("Error creating file in ./photo: %v!\n", err)
    http.Error(w, "Error creating file", http.StatusInternalServerError)
    return
  }
  _, err = io.Copy(outFile, file)
  if err != nil {
    fmt.Printf("Error copying file to ./photo: %v!\n", err)
    http.Error(w, "Error copying file", http.StatusInternalServerError)
    return
  }
  defer outFile.Close()
  var newUser = User {
    ID:         expectedID,
    Firstname:  r.FormValue("firstname"),
    Lastname:   r.FormValue("lastname"),
    Phone:      r.FormValue("phone"),
    Username:   r.FormValue("username"),
    Image:      newImageName,
    Bio:        r.FormValue("Bio"), 
  }
  if newUser.Firstname == "" || newUser.Lastname == "" || newUser.Phone == "" ||
  newUser.Username == "" || newUser.Bio == "" || newUser.Image == "" {
    w.WriteHeader(http.StatusBadRequest)
    fmt.Fprint(w, "Invalid request payload")
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
  res, err := pgdb.Exec(`UPDATE users SET firstname=$1, lastname=$2, phone=$3, username=$4, image=$5, bio=$6
    WHERE id = $7`,
    newUser.Firstname, newUser.Lastname, newUser.Phone,
    newUser.Username, newUser.Image, newUser.Bio, newUser.ID)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Printf("Error updating user: %v", err)
    return
  }
  changes, _ := res.RowsAffected()
  if changes == 0 {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotModified)
    fmt.Fprintf(w, "Error: either user %v was not found or user was completed and therefore not modifiable.", userID)
    return
  }
  w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "user with ID %s updated", userID)
}

func GetUserInfoByKeyword(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  if expectedID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  // vars := mux.Vars(r)
  //userID := vars["user_id"]
  //if strconv.Itoa(expectedID) != userID {
  //  w.WriteHeader(http.StatusUnauthorized)
  //  fmt.Fprintf(w, "Unauthorized access!\n")
  //  fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
  //  return
  //}
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    log.Fatal(err)
  }
  defer pgdb.Close()
  queryParams := r.URL.Query()
  numParams := len(queryParams)
  fmt.Printf("Number of Query Parameters: %d\nOnly the first parameter is used.", numParams)
  fmt.Printf("Query Parameter value: %v\n", r.URL.Query().Get("keyword"))
  //keyword := ""
  //for paramName := range queryParams {
  //  keyword = paramName
  //  fmt.Fprintf(w, "%s\n", paramName)
  //}
  // query := r.URL.Query()
  //rows, err := pgdb.Query("SELECT id, firstname, lastname, phone, username, image, bio FROM users WHERE $1 = $2",
  //  keyword, r.URL.Query().Get(keyword))
  var user User
  row := pgdb.QueryRow(`SELECT 
    id, firstname, lastname, phone, username, image, bio FROM users
    WHERE id = $1`,
    r.URL.Query().Get("keyword"))
  err = row.Scan(&user.ID, &user.Firstname, &user.Lastname,
    &user.Phone, &user.Username, &user.Image, &user.Bio)
  if err != nil {
    fmt.Printf("Error while scanning: %v\n", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  //defer rows.Close()
  //var resultRows []User
  //for rows.Next() {
  //  var user User
  //  err = rows.Scan(&user.ID, &user.Firstname, &user.Lastname,
  //    &user.Username, &user.Password, &user.Image, &user.Bio)
  //  if err != nil {
  //    http.Error(w, err.Error(), http.StatusInternalServerError)
  //    return
  //  }
  //  resultRows = append(resultRows, user)
  //}
  imagePath := filepath.Join(".", "photos", user.Image)
  imageData, err := ioutil.ReadFile(imagePath)
  if err != nil {
    fmt.Printf("Error while reading Image: %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  user.ImgBytes = imageData
  //responseJSON, err := json.Marshal(resultRows)
  //if err != nil {
  //  http.Error(w, err.Error(), http.StatusInternalServerError)
  //  return
  //}
  //w.Header().Set("Content-Type", "application/json")
  //w.Write(responseJSON)
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(user)
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  if expectedID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  userID := vars["user_id"]
  if strconv.Itoa(expectedID) != userID {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "Unauthorized access!\n")
    fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    log.Fatal(err)
  }
  defer pgdb.Close()
  row := pgdb.QueryRow("SELECT id, firstname, lastname, phone, username, image, bio FROM users WHERE id = $1",
    userID)
  var user User
  err = row.Scan(&user.ID, &user.Firstname, &user.Lastname,
    &user.Phone, &user.Username, &user.Image, &user.Bio)
  if err != nil {
    if err == sql.ErrNoRows {
      w.WriteHeader(http.StatusNotFound)
      fmt.Fprintf(w, "user with ID %s not found", userID)
      return
    }
    log.Fatal(err)
  }
  imagePath := filepath.Join(".", "photos", user.Image)
  imageData, err := ioutil.ReadFile(imagePath)
  if err != nil {
    fmt.Printf("Error while reading Image: %v", err)
    w.WriteHeader(http.StatusInternalServerError)
    return
  }
  user.ImgBytes = imageData
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(user)
  //fmt.Fprintf(w, "GET /api/users/%s - Retrieve user's information", userID)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
  expectedID := GetIDFromCookie(w, r)
  if expectedID == -1 {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "invalid cookie format!\n")
  }
  vars := mux.Vars(r)
  userID := vars["user_id"]
  if strconv.Itoa(expectedID) != userID {
    w.WriteHeader(http.StatusUnauthorized)
    fmt.Fprintf(w, "Unauthorized access!\n")
    fmt.Printf("expectedID(%v) != userID(%v)", expectedID, userID)
    return
  }
  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    log.Fatal(err)
  }
  defer pgdb.Close()
  result, err := pgdb.Exec("DELETE FROM users WHERE id = $1", userID)
  if err != nil {
    log.Fatal(err)
  }
  rowsAffected, _ := result.RowsAffected()
  if rowsAffected == 0 {
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintf(w, "Account was not found.")
    return
  }
  w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "User with ID %s deleted", userID)
}

func InitializePostgres() {
  // Connect to the PostgreSQL database
  pgConnStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)

  pgdb, err := sql.Open("postgres", pgConnStr)
  if err != nil {
    log.Fatal(err)
  }
  defer pgdb.Close()

  _, err = pgdb.Exec(`
    CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    firstname VARCHAR(64) NOT NULL,
    lastname VARCHAR(64) NOT NULL,
    phone VARCHAR(64) UNIQUE NOT NULL,
    username VARCHAR(64) UNIQUE NOT NULL,
    password VARCHAR(64) NOT NULL,
    image VARCHAR(255) UNIQUE NOT NULL,
    bio VARCHAR(255) NOT NULL
    );
    `)
  if err != nil {
    log.Fatal(err)
  }
  _, err = pgdb.Exec(`
    CREATE TABLE IF NOT EXISTS contacts (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_name VARCHAR(64) NOT NULL,
    PRIMARY KEY (user_id, contact_id)
    );
    `)
  if err != nil {
    log.Fatal(err)
  }
  _, err = pgdb.Exec(`
    CREATE TABLE IF NOT EXISTS chats (
    chat_id SERIAL PRIMARY KEY,
    people INTEGER[],
    created_at TIMESTAMP NOT NULL
    );
    `)
  if err != nil {
    log.Fatal(err)
  }
  _, err = pgdb.Exec(`
    CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER NOT NULL REFERENCES chats(chat_id) ON DELETE CASCADE,
    sender INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP NOT NULL
    );
    `)
  if err != nil {
    log.Fatal(err)
  }
  _, err = pgdb.Exec(`
    CREATE TABLE IF NOT EXISTS groups (
    group_id SERIAL PRIMARY KEY,
    people INTEGER[],
    created_at TIMESTAMP NOT NULL
    );
    `)
  if err != nil {
    log.Fatal(err)
  }
  // last_user is the id of the last user who was online
  // if a side of the chat's id is not the same as last_on_user
  // then he/she has not seen some messages.
  // notviewed is the number of messages that are not viewed
  // yet by this user.
  _, err = pgdb.Exec(`
    CREATE TABLE IF NOT EXISTS extensions (
    chat_id INTEGER PRIMARY KEY NOT NULL REFERENCES chats(chat_id) ON DELETE CASCADE,
    notviewed INTEGER NOT NULL, 
    last_user INTEGER NOT NULL, 
    last_message VARCHAR(2048) NOT NULL,
    last_msg_date TIMESTAMP NOT NULL,
    laston1 TIMESTAMP NOT NULL,
    laston2 TIMESTAMP NOT NULL
    );
    `)
  if err != nil {
    log.Fatal(err)
  }
}

func init() {
  rand.Seed(time.Now().UnixNano())
}

func generateRandomString(n int) string {
  b := make([]rune, n)
  for i := range b {
    b[i] = letters[rand.Intn(len(letters))]
  }
  return string(b)
}
