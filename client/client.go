package main

import (
	"crypto/sha512"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
)

var tokenSesion = ""
var userSecretKey = ""

func changeToken(newToken string) {
	tokenSesion = newToken
}

func changeSecretKey(newSecret string) {
	userSecretKey = newSecret
}

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("\n############################################################")
	fmt.Println("###################### ÆCloud Client #######################")
	fmt.Println("############################################################\n")
	fmt.Println("   -- Un cliente mediante comunicación TLS/HTTP en Go --\n")

	client()
}

func menu() string {
	fmt.Println("--- ÆCLOUD MENÚ ---")
	fmt.Println("0- VER LISTADO DE FICHEROS")
	fmt.Println("1- SUBIR FICHERO")
	fmt.Println("2- DESCARGAR FICHERO")
	fmt.Println("3- BORRAR FICHERO")
	fmt.Println("4- DROPBOX")
	fmt.Println("Q- SALIR")
	fmt.Print("Opción: ")
	var input string
	fmt.Scanf("%s\n", &input)

	return input
}

func mainMenu() string {
	fmt.Print("¿Es la primera vez que visitas ÆCLOUD? (S/N) \n-> ")
	var input string
	fmt.Scanf("%s\n", &input)
	fmt.Print("\n")
	return input
}

// gestiona el modo cliente
func client() {

	// --- FIXME: Refactor to init function
	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// ----

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
				enableTwoAuth(client, username)
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
	fmt.Println("----------------------------------")
	fmt.Println("-------- Inicio de sesión --------")
	fmt.Println("----------------------------------")
	fmt.Print("\n")
	fmt.Printf("-> Introduce usuario: ")
	fmt.Scanf("%s\n", &username)
	fmt.Println()
	fmt.Printf("-> Introduce contraseña: ")
	password, err := gopass.GetPasswd()
	chk(err)
	fmt.Print("\n\n")

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

	var login bool
	login = loginResponse.Ok

	if login {
		fmt.Println(" *** Hola de nuevo " + username + " ***")
		fmt.Print("\n")
		// Cambiamos el token de sesion
		changeToken(loginResponse.Token)
		changeSecretKey(loginResponse.SecretKey)

		if loginResponse.TwoFa {
			login = loginTwoAuth(client, tokenSesion)

		} else {
			// Try to enable 2FA
			if enableTwoAuth(client, username) {
				login = loginTwoAuth(client, tokenSesion)
			}
		}

		if login {
			// User menu
			var optMenu string = menu()
			for optMenu != "Q" {
				switch optMenu {
				case "0":
					listFiles(client)
				case "1":
					uploadFile(client)
				case "2":
					downloadFile(client)
				case "3":
					deleteFile(client)
				case "4":
					fmt.Print("\n")
					dropboxClient(client)
				default:
					fmt.Println("Opción incorrecta!")
				}
				optMenu = menu()
			}
		} else {
			fmt.Println("Error! token 2FA no válido")
		}

	} else {
		// Volver atras
		fmt.Println("Error! usuario o contraseña incorrectos")
	}
}
