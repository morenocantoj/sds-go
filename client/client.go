package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

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
	fmt.Println("--- CIFRADO DEL CÉSAR ---")
	fmt.Println("1- CODIFICAR")
	fmt.Println("2- DESCODIFICAR")
	fmt.Println("3- CONFIGURAR DESPLAZAMIENTO")
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
	io.Copy(os.Stdout, r.Body) // mostramos el cuerpo de la respuesta (es un reader)
	fmt.Println()
}
