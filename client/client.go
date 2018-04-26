package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
)

type loginStruct struct {
	Ok  bool
	Msg string
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
	fmt.Println("3- REGISTRO")
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

func register() {
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
			fmt.Println("¡Las contraseñas no coinciden!")

		} else {
			// Send data to server encrypted
		}
	}

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
		register()
	}

	var username string
	var password string
	fmt.Printf("Introduce usuario: ")
	fmt.Scanf("%s\n", &username)
	fmt.Println()
	fmt.Printf("Introduce contraseña: ")
	fmt.Scanf("%s\n", &password)

	// Estructura de datos
	data := url.Values{}
	data.Set("cmd", "login")
	data.Set("username", username)
	data.Set("password", password)

	r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
	chk(err)
	// Solo podemos leer una vez el body
	b, err := ioutil.ReadAll(r.Body)

	var loginResponse loginStruct
	err = json.Unmarshal(b, &loginResponse)

	defer r.Body.Close()

	fmt.Println(loginResponse)
	if loginResponse.Ok {
		fmt.Println(loginResponse.Msg)

		// User menu
		var optMenu string = menu()
		for optMenu != "Q" {
			switch optMenu {
			case "1":
				//TODO: Implement upload menu
				uploadFile()
			case "2":
				//TODO: Implement download menu
			default:
				fmt.Println("Opción incorrecta!")
			}
			optMenu = menu()
		}

	} else {
		// Volver atras
		fmt.Println(loginResponse.Msg)
	}
}

func uploadFile() {
	fmt.Println("Falta por implementar!!")
}
