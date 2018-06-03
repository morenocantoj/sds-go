package main

import (
	"database/sql"
	"errors"
	"strconv"
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
