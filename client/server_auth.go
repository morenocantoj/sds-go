package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
)

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

func checkTokenAuth(tokenVar tokenValid) bool {
	return tokenVar.Code != 401
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

func enableTwoAuth(client *http.Client, username string) bool {
	fmt.Printf("¿Deseas aplicar autenticación en dos pasos? (S/N) \n-> ")
	var enable string
	fmt.Scanf("%s\n", &enable)
	fmt.Print("\n")

	if enable == "S" || enable == "s" {
		// Enable 2FA
		data := url.Values{}
		data.Set("cmd", "enable2fa")
		data.Set("username", username)

		r, err := client.PostForm("https://localhost:10443", data)
		chk(err)
		b, err := ioutil.ReadAll(r.Body)

		// Get enabled token
		var twoFactorResponse twoFactorStruct
		err = json.Unmarshal(b, &twoFactorResponse)
		chk(err)

		if twoFactorResponse.Ok {
			fmt.Printf("Aquí tienes tu código de registro para la doble autenticación %s \n", twoFactorResponse.Token)
			fmt.Println("Este código es único e instransferible. ¡No lo pierdas!")
			return true
		} else {
			fmt.Println("¡Ha habido un error generando el código de registro de doble autenticación!")
		}
	}
	return false
}

func loginTwoAuth(client *http.Client, tokenSesion string) bool {
	// Two factor login needed
	fmt.Printf("Debes aplicar el valor de doble autenticación: ")
	var googleauth string
	fmt.Scanf("%s\n", &googleauth)
	fmt.Print("\n")

	data := url.Values{}
	data.Set("cmd", "doublelogin")
	data.Set("token", tokenSesion)
	data.Set("otpToken", googleauth)

	r, err := client.PostForm("https://localhost:10443", data) // enviamos por POST
	chk(err)
	// Solo podemos leer una vez el body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var loginResponse loginStruct
	err = json.Unmarshal(b, &loginResponse)

	if loginResponse.Ok {
		changeToken(loginResponse.Token)
		changeSecretKey(loginResponse.SecretKey)
		return true
	}

	return false
}
