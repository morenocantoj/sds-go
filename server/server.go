package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgryski/dgoogauth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goinggo/tracelog"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = ""

func loginfo(title string, msg string, function string, level string, err error) {
	switch level {
	case "trace":
		tracelog.Trace(title, function, msg)
	case "info":
		tracelog.Info(title, function, msg)
	case "warning":
		tracelog.Warning(title, function, msg)
	case "error":
		tracelog.Error(err, title, function)
	default:
		tracelog.Info("main", "main", msg)
	}
}

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		loginfo("error", "error", "chk", "error", e)
		panic(e)
	}
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

func responseLogin(w io.Writer, ok bool, twoFa bool, msg string, token string, secret string) {
	r := respLogin{Ok: ok, TwoFa: twoFa, Msg: msg, Token: token, SecretKey: secret} // formateamos respuesta
	rJSON, err := json.Marshal(&r)                                                  // codificamos en JSON
	chk(err)                                                                        // comprobamos error
	w.Write(rJSON)                                                                  // escribimos el JSON resultante
}

func responseFilesList(w io.Writer, list fileList) {
	rJSON, err := json.Marshal(&list) // codificamos en JSON
	chk(err)                          // comprobamos error
	w.Write(rJSON)
}

func response2FA(w io.Writer, ok bool, token string) {
	r := twoFactorStruct{Ok: ok, Token: token}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func responseCreateDropboxFolder(w io.Writer, created bool, msg string) {
	r := respCreateDropboxFolder{Created: created, Msg: msg}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func responseDownloadDropboxfile(w io.Writer, downloaded bool, content []byte, filename string, checksum string) {
	r := DropboxDownloadResponse{Downloaded: downloaded, Content: content, Filename: filename, Checksum: checksum}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func responseListFilesDropbox(w io.Writer, fileList fileListDropbox) {
	rJSON, err := json.Marshal(&fileList)
	chk(err)
	w.Write(rJSON)
}

func responseUploadFileDropbox(w io.Writer, uploaded bool, msg string) {
	r := uploadFileDropboxStruct{Uploaded: uploaded, Msg: msg}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func GetBearerToken(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("An authorization header is required")
	}
	token := strings.Split(header, " ")
	if len(token) != 2 {
		return "", fmt.Errorf("Malformed bearer token")
	}
	return token[1], nil
}

func validateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
		if err != nil {
			loginfo("validateMiddleware", "Error al recuperar el token JWT", "GetBearerToken", "error", err)
			json.NewEncoder(w).Encode(err)
			return
		}

		decodedToken, err := VerifyJwt(bearerToken, jwtSecret)
		if err != nil {
			loginfo("validateMiddleware", "Error al verificar el token JWT", "VerifyJwt", "error", err)
			json.NewEncoder(w).Encode(err)
			return
		}
		if decodedToken["authorized"] == true {
			loginfo("validateMiddleware", "Token de usuario válido", "VerifyJwt", "info", nil)
			next(w, req)
		} else {
			json.NewEncoder(w).Encode("Token no válido! Inicia sesión de nuevo!")
		}
	})
}

/**
* Lists all dropbox files by user
 */
func createDropboxFolder(w http.ResponseWriter, req *http.Request) {
	// Get user id by token and use it for create the new folder
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))
	loginfo("createDropboxFolder", "Usuario "+folderId+" crea carpeta personal", "Dropbox API", "info", nil)

	clientDropbox := &http.Client{}

	var jsonStr = []byte(`{
    "path": "/` + folderId + `",
    "autorename": false
		}`)

	body := bytes.NewBuffer(jsonStr)

	req, _ = http.NewRequest("POST", "https://api.dropboxapi.com/2/files/create_folder_v2", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)

	resp, err := clientDropbox.Do(req)
	chk(err)

	b, _ := ioutil.ReadAll(resp.Body)
	respDropbox := string(b)

	if strings.Contains(respDropbox, "metadata") && strings.Contains(respDropbox, "id") {
		// Folder created successfully
		responseCreateDropboxFolder(w, true, "¡Carpeta creada correctamente!")
	} else if strings.Contains(respDropbox, "error") {
		responseCreateDropboxFolder(w, false, "¡Carpeta personal ya existente!")
	} else {
		responseCreateDropboxFolder(w, false, "¡Error al crear carpeta personal!")
	}
}

func downloadFileDropbox(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))
	queryParams := req.URL.Query()
	filename := queryParams.Get("filename")

	loginfo("downloadfileDropbox", "Usuario "+folderId+" intenta bajar archivo "+filename, "Dropbox API", "info", nil)

	// Make request for dropbox
	clientDropbox := &http.Client{}
	fullPath := "/" + folderId + "/" + filename

	dropboxHeader := `{"path": "` + fullPath + `"}`

	req, err = http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
	chk(err)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)
	req.Header.Set("Dropbox-API-Arg", dropboxHeader)

	resp, err := clientDropbox.Do(req)
	chk(err)

	// Get body response
	b, _ := ioutil.ReadAll(resp.Body)
	contentFile := b

	jsonResult := DropboxDownloadStruct{}
	jsonString := resp.Header.Get("dropbox-api-result")

	err = json.Unmarshal([]byte(jsonString), &jsonResult)
	if err != nil {
		// Fail downloading file
		responseDownloadDropboxfile(w, false, contentFile, "err", "err")

	} else {
		// Correct response

		// Checksum
		checksumFile := sha256.Sum256(contentFile)
		slice := checksumFile[:]
		checksumString := encode64(slice)

		responseDownloadDropboxfile(w, true, contentFile, filename, checksumString)
	}
}

func listFilesDropbox(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))
	loginfo("listFilesDropbox", "Usuario "+folderId+" intenta listar sus archivos", "Dropbox API", "info", nil)

	// Make dropbox Request
	clientDropbox := &http.Client{}

	var jsonStr = []byte(`{
		"path": "/` + folderId + `",
		"recursive": false,
		"include_media_info": false,
    "include_deleted": false,
    "include_has_explicit_shared_members": false,
    "include_mounted_folders": true
		}`)

	body := bytes.NewBuffer(jsonStr)

	req, err = http.NewRequest("POST", "https://api.dropboxapi.com/2/files/list_folder", body)
	chk(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)

	resp, err := clientDropbox.Do(req)
	chk(err)

	b, err := ioutil.ReadAll(resp.Body)
	chk(err)

	var filesDropbox fileListDropbox
	err = json.Unmarshal(b, &filesDropbox)
	chk(err)

	responseListFilesDropbox(w, filesDropbox)
}

// gestiona el modo servidor
func server() {
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler))
	mux.HandleFunc("/dropbox/create/folder", validateMiddleware(createDropboxFolder))
	mux.HandleFunc("/dropbox/files/download", validateMiddleware(downloadFileDropbox))
	mux.HandleFunc("/dropbox/files", validateMiddleware(listFilesDropbox))
	mux.HandleFunc("/dropbox/files/upload", validateMiddleware(uploadFileDropbox))
	mux.HandleFunc("/files", validateMiddleware(handlerFileList))
	mux.HandleFunc("/files/checkPackage", validateMiddleware(handlerPackageCheck))
	mux.HandleFunc("/files/uploadPackage", validateMiddleware(handlerPackageUpload))
	mux.HandleFunc("/files/saveFile", validateMiddleware(handlerFileSave))
	mux.HandleFunc("/files/download", validateMiddleware(handlerFileDownload))
	mux.HandleFunc("/files/delete", validateMiddleware(handlerFileDelete))

	srv := &http.Server{Addr: ":10443", Handler: mux}

	go func() {
		if err := srv.ListenAndServeTLS("certificate/cert.pem", "certificate/server.key"); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")

	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Servidor detenido correctamente")
}

// función para codificar de []bytes a string (Base64)
func encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data) // sólo utiliza caracteres "imprimibles"
}

// función para decodificar de string a []bytes (Base64)
func decode64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s) // recupera el formato original
	chk(err)                                     // comprobamos el error
	return b                                     // devolvemos los datos originales
}

/**
* Checks username and password in database
* @param username
* @param password
 */
func checkLogin(username string, password string) bool {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkLogin", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var existingPassword sql.NullString
	row := db.QueryRow("SELECT password FROM users WHERE email = ?", username)
	err = row.Scan(&existingPassword)
	chk(err)
	loginfo("checkLogin", "Comprobado si existe usuario y extracción de contraseña", "db.QueryRow", "trace", nil)

	if existingPassword.Valid {
		// User exists
		var passwordString = existingPassword.String

		// Get bcrypt password from client
		passwordByte := decode64(password)
		chk(err)

		// Compare passwords
		err = bcrypt.CompareHashAndPassword(decode64(passwordString), passwordByte)
		loginfo("checkLogin", "Comparación de password en BBDD y login", "db.QueryRow", "trace", nil)

		if err == nil {
			// Password matches!
			return true
		} else {
			return false
		}

	} else {
		// No user
		return false
	}

	defer db.Close()
	return false
}

func registerUser(username string, password string) bool {
	if username == "" || password == "" {
		return false
	}

	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("registerUser", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	// Check if email is already in database
	var existingMail sql.NullString

	row := db.QueryRow("SELECT email FROM users WHERE email = ?", username)
	err = row.Scan(&existingMail)

	loginfo("registerUser", "Comprobar si existe cuenta registrada", "db.QueryRow", "trace", nil)

	if existingMail.Valid {
		// User exists
		return false
	} else {
		// User doesnt exists
		passwordSalted, err := bcrypt.GenerateFromPassword(decode64(password), bcrypt.DefaultCost)
		chk(err)
		// Create secret key for encription
		secretKey := generateRandomKey32Bytes()

		result, err := db.Exec("INSERT INTO users (email, password, secret_key) VALUES (?, ?, ?)", username, encode64(passwordSalted), secretKey)
		chk(err)
		idResult, err := result.LastInsertId()
		loginfo("registerUser", "Insertado de cuenta y password en base de datos", "db.Exec", "trace", nil)
		if idResult > 0 {
			return true
		} else {
			return false
		}
	}

	defer db.Close()
	return true
}

func checkTwoFaEnabled(username string) bool {
	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkTwoFaEnabled", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	// Check if token is already in the user
	var existingToken sql.NullString

	row := db.QueryRow("SELECT token_2fa FROM users WHERE email = ?", username)
	err = row.Scan(&existingToken)

	defer db.Close()

	if existingToken.Valid {
		return true
	}
	return false

}

func login(w http.ResponseWriter, req *http.Request) {
	username := req.Form.Get("username")
	password := req.Form.Get("password")
	loginfo("login", "Usuario "+username+" se intenta loguear en el sistema", "handler", "info", nil)

	if checkLogin(username, password) {
		loginfo("login", "Usuario "+username+" autenticado en el sistema", "handler", "info", nil)
		token := CreateTokenEndpoint(username, password)
		twoFa := checkTwoFaEnabled(username)
		secret, err := getUserSecretKeyByUsername(username)
		chk(err)

		responseLogin(w, true, twoFa, "Usuario "+username+" autenticado en el sistema", token, secret)
	} else {
		loginfo("login", "Usuario "+username+" ha fallado al autenticarse en el sistema", "handler", "warning", nil)
		responseLogin(w, false, false, "Usuario "+username+" autenticado en el sistema", "", "")
	}
}

func doubleLogin(w http.ResponseWriter, req *http.Request) {
	otpToken := req.Form.Get("otpToken")
	tokenString := req.Form.Get("token")
	tokenResult, correct := VerifyOtpEndpoint(tokenString, otpToken)
	// TODO: enviar secretKey. New function getUserSecretKeyById() + obtener userId a partir del token
	userId := getUserIdFromToken(tokenResult)
	secretKey, err := getUserSecretKeyById(userId)
	chk(err)
	responseLogin(w, correct, true, "Autenticación en el sistema (2FA)", tokenResult, secretKey)
}

func register(w http.ResponseWriter, req *http.Request) {
	username := req.Form.Get("username")
	password := req.Form.Get("password")

	if registerUser(username, password) {
		loginfo("register", "Se ha registrado un nuevo usuario "+username, "handler", "info", nil)
		response(w, true, "Registrado correctamente")
	} else {
		loginfo("register", "Error al registrar el usuario "+username, "handler", "warning", nil)
		response(w, false, "Error al registrar el usuario "+username)
	}
}

/**
* Creates JWT token
* @param w
* @param username
* @param password
 */
func CreateTokenEndpoint(username string, password string) string {
	token := make(map[string]interface{})
	token["username"] = username
	token["password"] = password
	token["authorized"] = false

	// Check if user has 2FA enabled
	if !checkTwoFaEnabled(username) {
		// We have only one step verification
		token["authorized"] = true
	}

	tokenString, err := SignJwt(token, jwtSecret)
	chk(err)

	return tokenString
}

func SignJwt(claims jwt.MapClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateToken(tokenString string) bool {
	loginfo("Validar token", "Validar token de usuario", "validateToken", "info", nil)

	decodedToken, err := VerifyJwt(tokenString, jwtSecret)
	if err != nil {
		return false
	}
	if decodedToken["authorized"] == false {
		return false
	} else {
		return true
	}
}

func VerifyJwt(token string, secret string) (map[string]interface{}, error) {
	jwToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !jwToken.Valid {
		return nil, fmt.Errorf("Invalid authorization token")
	}
	return jwToken.Claims.(jwt.MapClaims), nil
}

func VerifyOtpEndpoint(tokenString string, otpToken string) (string, bool) {
	decodedToken, err := VerifyJwt(tokenString, jwtSecret)
	username := decodedToken["username"]

	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)

	// Check if email is already in database
	var twoFactorToken []byte

	row := db.QueryRow("SELECT token_2fa FROM users WHERE email = ?", username)
	err = row.Scan(&twoFactorToken)
	chk(err)

	defer db.Close()

	if err != nil {
		return tokenString, false
	}
	otpc := &dgoogauth.OTPConfig{
		Secret:      string(twoFactorToken),
		WindowSize:  3,
		HotpCounter: 0,
	}

	decodedToken["authorized"], _ = otpc.Authenticate(otpToken)
	if decodedToken["authorized"] == false {
		return tokenString, false
	}
	jwToken, _ := SignJwt(decodedToken, jwtSecret)
	return jwToken, true
}

// Generates a 80 bit base32 encoded string
func generateSecretEndpoint() string {
	random := make([]byte, 10)
	rand.Read(random)
	secret := base32.StdEncoding.EncodeToString(random)
	return secret
}

// Generates a 32 byte base32 encoded string for AES-256
func generateRandomKey32Bytes() string {
	random := make([]byte, 32)
	rand.Read(random)
	secret := base32.StdEncoding.EncodeToString(random)
	return secret
}

func enableTwoFactor(w http.ResponseWriter, req *http.Request) {
	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("enableTwoFactor", "Conexión a MySQL abierta", "sql.Open", "trace", nil)
	tokenValue := generateSecretEndpoint()
	username := req.Form.Get("username")

	loginfo("enableTwoFactor", "Usuario "+username+" intenta habilitar 2FA en su cuenta", "enableTwoFactor", "info", nil)

	_, err = db.Exec("UPDATE users SET token_2fa = '" + tokenValue + "' WHERE email = '" + username + "'")

	if err != nil {
		// Send false values
		response2FA(w, false, "")
	} else {
		// Send new secret token for 2fa
		response2FA(w, true, tokenValue)
	}

	defer db.Close()
}

func handler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()                              // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	switch req.Form.Get("cmd") { // comprobamos comando desde el cliente

	case "login": // Check login
		login(w, req)
	case "register":
		register(w, req)
	case "doublelogin": // Check 2FA auth
		doubleLogin(w, req)
	case "enable2fa":
		enableTwoFactor(w, req)
	default:
		loginfo("main", "Acción no válida", "handler", "warning", nil)
		response(w, false, "Comando inválido")
	}
}

func main() {

	tracelog.StartFile(1, "log", 30)

	// Generate a JWT secret each time we restart server
	jwtSecret = generateSecretEndpoint()

	fmt.Println("\n############################################################")
	fmt.Println("###################### ÆCloud Server #######################")
	fmt.Println("############################################################\n")
	fmt.Println("   -- Un server mediante comunicación TLS/HTTP en Go  --")
	fmt.Println("\nServer listening on port 10443 ...\n")

	server()

	tracelog.Stop()
}
