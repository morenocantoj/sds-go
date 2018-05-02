package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goinggo/tracelog"
	"golang.org/x/crypto/bcrypt"
)

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

// respuesta del servidor
type resp struct {
	Ok  bool   // true -> correcto, false -> error
	Msg string // mensaje adicional
}

type user struct {
	Username string
	Password string
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

// gestiona el modo servidor
func server() {
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handler))

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
	db, err := sql.Open("mysql", "sds:sds@/sds")
	chk(err)

	var existingPassword sql.NullString
	row := db.QueryRow("SELECT password FROM users WHERE email = ?", username)
	err = row.Scan(&existingPassword)
	chk(err)

	if existingPassword.Valid {
		// User exists
		var passwordString = existingPassword.String

		// Get bcrypt password from client
		passwordByte := decode64(password)
		chk(err)

		// Compare passwords
		err = bcrypt.CompareHashAndPassword(decode64(passwordString), passwordByte)

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
	db, err := sql.Open("mysql", "sds:sds@/sds")
	chk(err)

	// Check if email is already in database

	var existingMail sql.NullString

	row := db.QueryRow("SELECT email FROM users WHERE email = ?", username)
	err = row.Scan(&existingMail)

	if existingMail.Valid {
		// User exists
		return false
	} else {
		// User doesnt exists
		passwordSalted, err := bcrypt.GenerateFromPassword(decode64(password), bcrypt.DefaultCost)
		chk(err)
		result, err := db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", username, encode64(passwordSalted))
		chk(err)
		idResult, err := result.LastInsertId()

		if idResult > 0 {
			return true
		} else {
			return false
		}
	}

	defer db.Close()
	return true
}

func handler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()                              // es necesario parsear el formulario
	w.Header().Set("Content-Type", "text/plain") // cabecera estándar

	switch req.Form.Get("cmd") { // comprobamos comando desde el cliente
	case "hola": // ** registro
		response(w, true, "Hola "+req.Form.Get("mensaje"))
		fmt.Println("Arancha me ha hecho una petición desde su pobre ordenador")

	case "login": // Check login
		username := req.Form.Get("username")
		password := req.Form.Get("password")
		if checkLogin(username, password) {
			fmt.Println("Usuario " + username + " autenticado en el sistema")
			response(w, true, "Hola de nuevo "+username)
		}

	case "register":
		username := req.Form.Get("username")
		password := req.Form.Get("password")

		if registerUser(username, password) {
			fmt.Println("Se ha registrado un nuevo usuario " + username)
			response(w, true, "Registrado correctamente")
		} else {
			fmt.Println("Error al registrar el usuario " + username)
			response(w, false, "Error al registrar el usuario "+username)
		}

	default:
		response(w, false, "Comando inválido")
	}

}

func main() {
	fmt.Println("ÆCloud en GO")
	tracelog.StartFile(1, "log", 30)
	tracelog.Trace("main", "main", "Hola buenas")
	tracelog.Info("main", "main", "Queriamos unos kebabs")
	tracelog.Warning("main", "main", "EL FAKING WARNING")
	fmt.Println("Un ejemplo de server/cliente mediante TLS/HTTP en Go.")
	s := "Introduce srv para funcionalidad de servidor y cli para funcionalidad de cliente"

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "srv":
			fmt.Println("Entrando en modo servidor...")
			server()
		default:
			fmt.Println("Parámetro '", os.Args[1], "' desconocido. ", s)
		}
	} else {
		fmt.Println(s)
	}

	tracelog.Stop()
}
