package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	fmt.Println("Un ÆCloud cliente mediante TLS/HTTP en Go.")
	client()
}

func menu() string {
	fmt.Println("--- ÆCLOUD MENÚ ---")
	fmt.Println("1- SUBIR FICHERO")
	fmt.Println("2- DESCARGAR FICHERO")
	fmt.Println("Q- SALIR")
	fmt.Print("Opción: ")
	var input string
	fmt.Scanf("%s\n", &input)

	return input
}

// gestiona el modo cliente
func client() {

	/* creamos un cliente especial que no comprueba la validez de los certificados
	esto es necesario por que usamos certificados autofirmados (para pruebas) */
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

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
			optMenu = menu()
		}

	} else {
		// Volver atras
		fmt.Println(loginResponse.Msg)

	}

}
