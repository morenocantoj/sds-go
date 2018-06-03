package main

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/dgryski/dgoogauth"
	"golang.org/x/crypto/bcrypt"
)

func getUserIdFromToken(token string) int {
	decodedToken, err := VerifyJwt(token, jwtSecret)
	username := decodedToken["username"]

	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getUserIdFromToken", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	// Check if email is already in database
	var userIdString []byte

	row := db.QueryRow("SELECT id FROM users WHERE email = ?", username)
	err = row.Scan(&userIdString)
	chk(err)
	loginfo("getUserIdFromToken", "Obteniendo id del usuario a partir del token", "db.QueryRow", "trace", nil)

	defer db.Close()

	userId, err := strconv.Atoi(string(userIdString))
	chk(err)

	return userId
}

func getUserSecretKeyByUsername(username string) (string, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getUserSecretKeyByUsername", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT secret_key FROM users WHERE email = ?", username)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return "", nil
	}
	chk(err)
	loginfo("getUserSecretKeyByUsername", "Obteniendo la clave secreta del usuario para cifrar archivos", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {
		var userSecret = sqlResponse.String
		return userSecret, nil
	} else {
		return "", errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return "", errors.New("SQL Error: something has gone wrong")
}

func getUserSecretKeyById(userId int) (string, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getUserSecretKeyById", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT secret_key FROM users WHERE id = ?", userId)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return "", nil
	}
	chk(err)
	loginfo("getUserSecretKeyById", "Obteniendo la clave secreta del usuario para cifrar archivos", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {
		var userSecret = sqlResponse.String
		return userSecret, nil
	} else {
		return "", errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return "", errors.New("SQL Error: something has gone wrong")
}

func registerUser(username string, password string) bool {
	if username == "" || password == "" {
		return false
	}

	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("registerUser", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	// Check if email is already in database
	var existingMail sql.NullString

	row := db.QueryRow("SELECT email FROM users WHERE email = ?", username)
	err = row.Scan(&existingMail)

	loginfo("registerUser", "Comprobar si existe cuenta registrada", "db.QueryRow", "trace", nil)

	if existingMail.Valid {
		// User exists
		return false
	} else {
		// User doesnt exists
		passwordSalted, err := bcrypt.GenerateFromPassword(decode64(password), bcrypt.DefaultCost)
		chk(err)
		// Create secret key for encription
		secretKey := generateRandomKey32Bytes()

		result, err := db.Exec("INSERT INTO users (email, password, secret_key) VALUES (?, ?, ?)", username, encode64(passwordSalted), secretKey)
		chk(err)
		idResult, err := result.LastInsertId()
		loginfo("registerUser", "Insertado de cuenta y password en base de datos", "db.Exec", "trace", nil)
		if idResult > 0 {
			return true
		} else {
			return false
		}
	}

	defer db.Close()
	return true
}

/**
* Checks username and password in database
* @param username
* @param password
 */
func checkLogin(username string, password string) bool {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkLogin", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var existingPassword sql.NullString
	row := db.QueryRow("SELECT password FROM users WHERE email = ?", username)
	err = row.Scan(&existingPassword)
	chk(err)
	loginfo("checkLogin", "Comprobado si existe usuario y extracción de contraseña", "db.QueryRow", "trace", nil)

	if existingPassword.Valid {
		// User exists
		var passwordString = existingPassword.String

		// Get bcrypt password from client
		passwordByte := decode64(password)
		chk(err)

		// Compare passwords
		err = bcrypt.CompareHashAndPassword(decode64(passwordString), passwordByte)
		loginfo("checkLogin", "Comparación de password en BBDD y login", "db.QueryRow", "trace", nil)

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

func checkTwoFaEnabled(username string) bool {
	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkTwoFaEnabled", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	// Check if token is already in the user
	var existingToken sql.NullString

	row := db.QueryRow("SELECT token_2fa FROM users WHERE email = ?", username)
	err = row.Scan(&existingToken)
	loginfo("checkTwoFaEnabled", "Comprobado si el usuario tiene habilitado el factor de doble autenticación", "db.QueryRow", "trace", nil)

	defer db.Close()

	if existingToken.Valid {
		return true
	}
	return false
}

func VerifyOtpEndpoint(tokenString string, otpToken string) (string, bool) {
	decodedToken, err := VerifyJwt(tokenString, jwtSecret)
	username := decodedToken["username"]

	// Open database
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("VerifyOtpEndpoint", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	// Check if email is already in database
	var twoFactorToken []byte

	row := db.QueryRow("SELECT token_2fa FROM users WHERE email = ?", username)
	err = row.Scan(&twoFactorToken)
	chk(err)
	loginfo("VerifyOtpEndpoint", "Comprobado si el factor de doble autenticación es correcto", "db.QueryRow", "trace", nil)

	defer db.Close()

	if err != nil {
		return tokenString, false
	}
	otpc := &dgoogauth.OTPConfig{
		Secret:      string(twoFactorToken),
		WindowSize:  3,
		HotpCounter: 0,
	}

	decodedToken["authorized"], _ = otpc.Authenticate(otpToken)
	if decodedToken["authorized"] == false {
		return tokenString, false
	}
	jwToken, _ := SignJwt(decodedToken, jwtSecret)
	return jwToken, true
}
