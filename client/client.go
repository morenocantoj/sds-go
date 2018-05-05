package main

import (
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
)

var tokenSesion = ""

func changeToken(newToken string) {
	tokenSesion = newToken
}

type loginStruct struct {
	Ok    bool
	Msg   string
	Token string
}

type registerStruct struct {
	Ok  bool
	Msg string
}

type JwtToken struct {
	Token string `json:"token"`
}

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("Un ÆCloud cliente mediante TLS/HTTP en Go.")
	client()
}

func menu() string {
	fmt.Println("--- ÆCLOUD MENÚ ---")
	fmt.Println("1- SUBIR FICHERO")
	fmt.Println("2- DESCARGAR FICHERO")
	fmt.Println("3- PROBAR EL TOKEN")
	fmt.Println("Q- SALIR")
	fmt.Print("Opción: ")
	var input string
	fmt.Scanf("%s\n", &input)

	return input
}

func mainMenu() string {
	fmt.Println("¿Es la primera vez que visitas ÆCLOUD?")
	var input string
	fmt.Scanf("%s\n", &input)

	return input
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

func register() (string, string) {
	var username string
	var password []byte
	var password_repeat []byte

	for username == "" || encode64(password) == "" || encode64(password_repeat) == "" {
		fmt.Println("Introduce la cuenta con la que quieres registrarte...")
		fmt.Scanf("%s\n", &username)
		fmt.Println("Ahora introduce la contraseña con la que quieres loguearte...")
		password, err := gopass.GetPasswd()
		chk(err)
		fmt.Println("Repite otra vez la contraseña...")
		password_repeat, err := gopass.GetPasswd()
		chk(err)

		if encode64(password) != encode64(password_repeat) {
			fmt.Printf("¡Las contraseñas no coinciden!")
			return "", ""
		} else {
			// Send data to server encrypted
			passwordHash := sha512.Sum512(password)
			slice := passwordHash[:]
			passBase64 := encode64(slice)

			return username, passBase64

		}
	}
	return "", ""
}

// gestiona el modo cliente
func client() {

	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	var optMainMenu string = mainMenu()
	if optMainMenu != "N" {
		username, passHash := register()

		if username != "" && passHash != "" {
			// Send values for register
			// Estructura de datos
			data := url.Values{}
			data.Set("cmd", "register")
			data.Set("username", username)
			data.Set("password", passHash)

			// Send POST values
			r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
			chk(err)

			b, err := ioutil.ReadAll(r.Body)

			var registerResponse registerStruct
			err = json.Unmarshal(b, &registerResponse)
			chk(err)

			if registerResponse.Ok {
				fmt.Println("¡Te has registrado correctamente en el sistema!")
			} else {
				fmt.Println("Error al registrarte! Puede que tu nombre de usuario ya esté en uso")
			}

			defer r.Body.Close()

		} else {
			fmt.Println("Hay errores en tu formulario de registro")
		}
	}

	var username string
	var password []byte
	fmt.Println("-- Inicio de sesión --")
	fmt.Printf("Introduce usuario: ")
	fmt.Scanf("%s\n", &username)
	fmt.Println()
	fmt.Printf("Introduce contraseña: ")
	password, err := gopass.GetPasswd()
	chk(err)

	// Send data to server encrypted
	passwordHash := sha512.Sum512(password)
	slice := passwordHash[:]
	passBase64 := encode64(slice)

	// Estructura de datos
	data := url.Values{}
	data.Set("cmd", "login")
	data.Set("username", username)
	data.Set("password", passBase64)

	r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
	chk(err)
	// Solo podemos leer una vez el body
	b, err := ioutil.ReadAll(r.Body)

	var loginResponse loginStruct
	err = json.Unmarshal(b, &loginResponse)

	if loginResponse.Ok {
		fmt.Println("Hola de nuevo " + username)
		// Cambiamos el token de sesion
		changeToken(loginResponse.Token)

		fmt.Printf("Debes aplicar el valor de doble autenticación: ")
		var googleauth string
		fmt.Scanf("%s\n", &googleauth)
		fmt.Println(googleauth)

		data := url.Values{}
		data.Set("cmd", "doublelogin")
		data.Set("token", tokenSesion)
		data.Set("otpToken", googleauth)

		r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
		chk(err)
		// Solo podemos leer una vez el body
		b, err := ioutil.ReadAll(r.Body)
		err = json.Unmarshal(b, &loginResponse)

		if loginResponse.Ok {
			fmt.Println("factor de doble autenticación bien!")
		}

		// User menu
		var optMenu string = menu()
		for optMenu != "Q" {
			switch optMenu {
			case "1":
				//TODO: Implement upload menu
				uploadFile()
			case "2":
				//TODO: Implement download menu
			case "3":
				// Check token
				data := url.Values{}
				data.Set("cmd", "tokencheck")
				data.Set("token", tokenSesion)
				_, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
				chk(err)
			default:
				fmt.Println("Opción incorrecta!")
			}
			optMenu = menu()
		}

	} else {
		// Volver atras
		fmt.Println("Error! usuario o contraseña incorrectos")
	}

	defer r.Body.Close()
}

func uploadFile() {
	fmt.Println("Falta por implementar!!")
}
