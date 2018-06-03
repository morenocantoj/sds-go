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
var userSecretKey = ""

func changeToken(newToken string) {
	tokenSesion = newToken
}

func changeSecretKey(newSecret string) {
	userSecretKey = newSecret
}

type loginStruct struct {
	Ok        bool
	TwoFa     bool
	Msg       string
	Token     string
	SecretKey string
}

type twoFactorStruct struct {
	Ok    bool
	Token string
}

type registerStruct struct {
	Ok  bool
	Msg string
}

type JwtToken struct {
	Token string `json:"token"`
}

type createDropboxFolder struct {
	Created bool
	Msg     string
}

type DropboxDownload struct {
	Downloaded bool
	Content    []byte
	Filename   string
	Checksum   string
}

type UserFile struct {
	Id       string
	Filename string
}

type fileEnumDropboxStruct struct {
	Tag             string      `json:".tag"`
	Name            string      `json:"name"`
	Id              string      `json:"id"`
	Client_modified string      `json:"client_modified"`
	Server_modified string      `json:"server_modified"`
	Rev             string      `json:"rev"`
	Size            int         `json:"size"`
	Path_lower      string      `json:"path_lower"`
	Path_display    string      `json:"path_display"`
	Sharing_info    interface{} `json:"sharing_info"`
	Property_groups interface{} `json:"property_groups"`
	Shared          bool        `json:"has_explicit_shared_members"`
	Content_hash    string      `json:"content_hash"`
}

type fileListDropbox struct {
	Entries  []fileEnumDropboxStruct `json:"entries"`
	Cursor   string                  `json:"cursor"`
	Has_more bool                    `json:"has_more"`
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

func enableTwoAuth(client *http.Client, username string) bool {
	fmt.Printf("¿Deseas aplicar autenticación en dos pasos? (S/N)")
	var enable string
	fmt.Scanf("%s\n", &enable)

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

	var login bool
	login = loginResponse.Ok

	if login {
		fmt.Println("Hola de nuevo " + username)
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
					//TODO: Implement list files menu
					listFiles(client)
				case "1":
					//TODO: Implement upload menu
					uploadFile(client)
				case "2":
					downloadFile(client)
				case "4":
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
