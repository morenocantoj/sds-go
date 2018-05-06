package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
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
	fmt.Println("3- REGISTRO")
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
	var password string
	var password_repeat string

	for username == "" || password == "" || password_repeat == "" {
		fmt.Println("Introduce la cuenta con la que quieres registrarte...")
		fmt.Scanf("%s\n", &username)
		fmt.Println("Ahora introduce la contraseña con la que quieres loguearte...")
		fmt.Scanf("%s\n", &password)
		fmt.Println("Repite otra vez la contraseña...")
		fmt.Scanf("%s\n", &password_repeat)

		if password != password_repeat {
			fmt.Println("¡Las contraseñas no coinciden!")
			password = ""
		}
	}

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

	//r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
	//chk(err)
	// Solo podemos leer una vez el body
	//b, err := ioutil.ReadAll(r.Body)

	var loginResponse loginStruct
	//err = json.Unmarshal(b, &loginResponse)

	//defer r.Body.Close()

	//fmt.Println(loginResponse)
	// FIXME: borrar
	loginResponse.Ok = true
	if loginResponse.Ok {
		fmt.Println(loginResponse.Msg)

		// User menu
		var optMenu string = menu()
		for optMenu != "Q" {
			switch optMenu {
			case "0":
				//TODO: Implement list files menu
				listFiles()
			case "1":
				//TODO: Implement upload menu
				uploadFile(client)
			case "2":
				//TODO: Implement download menu
				downloadFile()
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
