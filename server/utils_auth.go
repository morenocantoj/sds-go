package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
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

func GetBearerToken(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("An authorization header is required")
	}
	token := strings.Split(header, " ")
	if len(token) != 2 {
		return "", fmt.Errorf("Malformed bearer token")
	}
	return token[1], nil
}

/**
* Creates JWT token
* @param w
* @param username
* @param password
 */
func CreateTokenEndpoint(username string, password string) string {
	token := make(map[string]interface{})
	token["username"] = username
	token["password"] = password
	token["authorized"] = false

	// Check if user has 2FA enabled
	if !checkTwoFaEnabled(username) {
		// We have only one step verification
		token["authorized"] = true
	}

	tokenString, err := SignJwt(token, jwtSecret)
	chk(err)

	return tokenString
}

func SignJwt(claims jwt.MapClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func validateToken(tokenString string) bool {
	loginfo("Validar token", "Validar token de usuario", "validateToken", "info", nil)

	decodedToken, err := VerifyJwt(tokenString, jwtSecret)
	if err != nil {
		return false
	}
	if decodedToken["authorized"] == false {
		return false
	} else {
		return true
	}
}

func VerifyJwt(token string, secret string) (map[string]interface{}, error) {
	jwToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !jwToken.Valid {
		return nil, fmt.Errorf("Invalid authorization token")
	}
	return jwToken.Claims.(jwt.MapClaims), nil
}

// Generates a 80 bit base32 encoded string
func generateSecretEndpoint() string {
	random := make([]byte, 10)
	rand.Read(random)
	secret := base32.StdEncoding.EncodeToString(random)
	return secret
}

// Generates a 32 byte base32 encoded string for AES-256
func generateRandomKey32Bytes() string {
	random := make([]byte, 32)
	rand.Read(random)
	secret := base32.StdEncoding.EncodeToString(random)
	return secret
}
