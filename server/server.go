package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goinggo/tracelog"
)

var jwtSecret = ""

func loginfo(title string, msg string, function string, level string, err error) {
	switch level {
	case "trace":
		tracelog.Trace(title, function, msg)
	case "info":
		tracelog.Info(title, function, msg)
	case "warning":
		tracelog.Warning(title, function, msg)
	case "error":
		tracelog.Error(err, title, function)
	default:
		tracelog.Info("main", "main", msg)
	}
}

// función para comprobar errores (ahorra escritura)
func chk(e error) {
	if e != nil {
		loginfo("error", "error", "chk", "error", e)
		panic(e)
	}
}

// función para escribir una respuesta del servidor
func response(w io.Writer, ok bool, msg string) {
	r := resp{Ok: ok, Msg: msg}    // formateamos respuesta
	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)                 // escribimos el JSON resultante
}

func validateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
		if err != nil {
			loginfo("validateMiddleware", "Error al recuperar el token JWT", "GetBearerToken", "error", err)
			json.NewEncoder(w).Encode(err)
			return
		}

		decodedToken, err := VerifyJwt(bearerToken, jwtSecret)
		if err != nil {
			loginfo("validateMiddleware", "Error al verificar el token JWT", "VerifyJwt", "error", err)
			json.NewEncoder(w).Encode(err)
			return
		}
		if decodedToken["authorized"] == true {
			loginfo("validateMiddleware", "Token de usuario válido", "VerifyJwt", "info", nil)
			next(w, req)
		} else {
			json.NewEncoder(w).Encode("Token no válido! Inicia sesión de nuevo!")
		}
	})
}

// gestiona el modo servidor
func server() {
	// suscripción SIGINT
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(authHandler))
	mux.HandleFunc("/dropbox/create/folder", validateMiddleware(createDropboxFolder))
	mux.HandleFunc("/dropbox/files/download", validateMiddleware(downloadFileDropbox))
	mux.HandleFunc("/dropbox/files", validateMiddleware(listFilesDropbox))
	mux.HandleFunc("/dropbox/files/upload", validateMiddleware(uploadFileDropbox))
	mux.HandleFunc("/files", validateMiddleware(handlerFileList))
	mux.HandleFunc("/files/checkPackage", validateMiddleware(handlerPackageCheck))
	mux.HandleFunc("/files/uploadPackage", validateMiddleware(handlerPackageUpload))
	mux.HandleFunc("/files/saveFile", validateMiddleware(handlerFileSave))
	mux.HandleFunc("/files/download", validateMiddleware(handlerFileDownload))
	mux.HandleFunc("/files/delete", validateMiddleware(handlerFileDelete))

	srv := &http.Server{Addr: ":10443", Handler: mux}

	go func() {
		if err := srv.ListenAndServeTLS("certificate/cert.pem", "certificate/server.key"); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan // espera señal SIGINT
	log.Println("Apagando servidor ...")

	// apagar servidor de forma segura
	ctx, fnc := context.WithTimeout(context.Background(), 5*time.Second)
	fnc()
	srv.Shutdown(ctx)

	log.Println("Servidor detenido correctamente")
}

func main() {

	tracelog.StartFile(1, "log", 30)

	// Generate a JWT secret each time we restart server
	jwtSecret = generateSecretEndpoint()

	fmt.Println("\n############################################################")
	fmt.Println("###################### ÆCloud Server #######################")
	fmt.Println("############################################################\n")
	fmt.Println("   -- Un server mediante comunicación TLS/HTTP en Go  --")
	fmt.Println("\nServer listening on port 10443 ...\n")

	server()

	tracelog.Stop()
}
