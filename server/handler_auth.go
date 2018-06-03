package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
)

/**
 * Server Response Functions
 */
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

// Handler Switch Function
func authHandler(w http.ResponseWriter, req *http.Request) {
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
		loginfo("main", "Acción no válida", "authHandler", "warning", nil)
		response(w, false, "Comando inválido")
	}
}

/**
 * AUTH HANDLERS FUNCTIONS
 */

func login(w http.ResponseWriter, req *http.Request) {
	username := req.Form.Get("username")
	password := req.Form.Get("password")
	loginfo("login", "Usuario "+username+" se intenta loguear en el sistema", "authHandler", "info", nil)

	if checkLogin(username, password) {
		loginfo("login", "Usuario "+username+" autenticado en el sistema", "authHandler", "info", nil)
		token := CreateTokenEndpoint(username, password)
		twoFa := checkTwoFaEnabled(username)
		secret, err := getUserSecretKeyByUsername(username)
		chk(err)

		responseLogin(w, true, twoFa, "Usuario "+username+" autenticado en el sistema", token, secret)
	} else {
		loginfo("login", "Usuario "+username+" ha fallado al autenticarse en el sistema", "authHandler", "warning", nil)
		responseLogin(w, false, false, "Usuario "+username+" autenticado en el sistema", "", "")
	}
}

func register(w http.ResponseWriter, req *http.Request) {
	username := req.Form.Get("username")
	password := req.Form.Get("password")

	if registerUser(username, password) {
		loginfo("register", "Se ha registrado un nuevo usuario "+username, "authHandler", "info", nil)
		response(w, true, "Registrado correctamente")
	} else {
		loginfo("register", "Error al registrar el usuario "+username, "authHandler", "warning", nil)
		response(w, false, "Error al registrar el usuario "+username)
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
