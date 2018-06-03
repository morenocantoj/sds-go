package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

func dropboxClient(client *http.Client) {
	optMenu := dropboxMenu()

	for optMenu != "Q" {
		switch optMenu {
		case "0":
			listFilesDropboxClient(client, tokenSesion)
		case "1":
			uploadFileDropboxClient(client, tokenSesion)
		case "2":
			downloadFileDropboxClient(client, tokenSesion)
		case "3":
			createDropboxFolderClient(client, tokenSesion)
		default:
			fmt.Println("\nOpción incorrecta!\n")
		}
		optMenu = dropboxMenu()
	}
}

func uploadFileDropboxClient(client *http.Client, token string) {
	var inputFile string
	fmt.Println("Introduce el archivo a subir (P.ej.: /Users/username/Desktop/file.txt) ")
	fmt.Println("-- El fichero debe ser menor de 150MB --")
	fmt.Print("-> ")
	fmt.Scanf("%s\n", &inputFile)

	var fileAbsPath string
	isAbsolute := filepath.IsAbs(inputFile)
	if isAbsolute {
		fileAbsPath = inputFile
	} else {
		folderPath, err := filepath.Abs("./")
		chk(err)
		fileAbsPath = folderPath + "/" + inputFile
	}
	filename := filepath.Base(fileAbsPath)

	fileData, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		fmt.Printf("\nERROR!! No se encuentra el archivo introducido\n\n")
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	reader := bytes.NewReader(fileData)
	part, err := writer.CreateFormFile("file", filename)
	chk(err)

	_, err = io.Copy(part, reader)
	chk(err)

	// File checksum
	checksumFile := sha256.Sum256(fileData)
	slice := checksumFile[:]
	checksumString := encode64(slice)

	params := map[string]string{
		"filename": filename,
		"checksum": checksumString,
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	chk(err)

	url := "https://localhost:10443/dropbox/files/upload"
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	chk(err)

	// Get body response
	b, _ := ioutil.ReadAll(resp.Body)
	var uploadResponse uploadFileDropbox
	err = json.Unmarshal(b, &uploadResponse)

	// Verify token
	var tokenValid tokenValid
	_ = json.Unmarshal(b, &tokenValid)
	if !checkTokenAuth(tokenValid) {
		fmt.Println("Sesión caducada! Loguéate de nuevo para continuar!")
		return
	}

	fmt.Printf("\n%s\n\n", uploadResponse.Msg)
}

func listFilesDropboxClient(client *http.Client, token string) {
	url := "https://localhost:10443/dropbox/files"
	req, err := http.NewRequest("GET", url, nil)
	chk(err)

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	chk(err)

	b, err := ioutil.ReadAll(resp.Body)
	chk(err)

	var filesDropbox fileListDropbox
	err = json.Unmarshal(b, &filesDropbox)

	// Verify token
	var tokenValid tokenValid
	_ = json.Unmarshal(b, &tokenValid)
	if !checkTokenAuth(tokenValid) {
		fmt.Println("Sesión caducada! Loguéate de nuevo para continuar!")
		return
	}

	if err != nil {
		fmt.Println("\n¡Ha habido un error en la petición!\n")
	} else {
		// Check if we have files
		if len(filesDropbox.Entries) > 0 {
			fmt.Println("\nEstos son los ficheros disponibles en Dropbox:\n")
			fmt.Println("\t ID - Filename")
			fmt.Println("\t-------------------------------------")

			for i, file := range filesDropbox.Entries {
				fmt.Printf("\t  %d - %s \n", i+1, file.Name)
			}
			fmt.Printf("\n\n")
		} else {
			fmt.Println("\n¡No tienes ficheros subidos en la plataforma!\n\n")
		}
	}

}

func downloadFileDropboxClient(client *http.Client, token string) {
	listFilesDropboxClient(client, token)

	fmt.Print("Introduce el nombre del archivo a descargar: ")
	var filename string
	fmt.Scanf("%s\n", &filename)
	fmt.Print("\n")

	url := "https://localhost:10443/dropbox/files/download?filename=" + filename
	req, err := http.NewRequest("GET", url, nil)
	chk(err)

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	chk(err)

	b, err := ioutil.ReadAll(resp.Body)
	chk(err)

	var downloadedFile DropboxDownload
	err = json.Unmarshal(b, &downloadedFile)
	chk(err)

	// Verify token
	var tokenValid tokenValid
	_ = json.Unmarshal(b, &tokenValid)
	if !checkTokenAuth(tokenValid) {
		fmt.Println("Sesión caducada! Loguéate de nuevo para continuar!")
		return
	}

	if downloadedFile.Downloaded == true {
		// Compare file checksum
		checksumFile := sha256.Sum256(downloadedFile.Content)
		slice := checksumFile[:]
		checksumString := encode64(slice)

		if checksumString == downloadedFile.Checksum {
			// Save file
			err = ioutil.WriteFile("./downloads/"+downloadedFile.Filename, downloadedFile.Content, 0644)
			chk(err)
			fmt.Println("Fichero descargado correctamente en el directorio 'downloads' \n\n")

		} else {
			// Corrupted file
			fmt.Println("¡Error! El archivo no concuerda. Puede que el fichero esté corrompido")
		}
	} else {
		// Error downloading file
		fmt.Println("¡Error al descargar el fichero!\n\n")
	}
}

func createDropboxFolderClient(client *http.Client, token string) {

	fmt.Println("Si tienes alguna carpeta ya creada no se hará nada")

	// Create folder
	req, err := http.NewRequest("POST", "https://localhost:10443/dropbox/create/folder", nil)
	chk(err)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	chk(err)

	b, err := ioutil.ReadAll(resp.Body)
	chk(err)

	var dropboxFolderResponse createDropboxFolder
	err = json.Unmarshal(b, &dropboxFolderResponse)

	// Verify token
	var tokenValid tokenValid
	_ = json.Unmarshal(b, &tokenValid)
	if !checkTokenAuth(tokenValid) {
		fmt.Println("Sesión caducada! Loguéate de nuevo para continuar!")
		return
	}

	if dropboxFolderResponse.Created {
		fmt.Println(dropboxFolderResponse.Msg)
	} else {
		fmt.Printf("%s \n", dropboxFolderResponse.Msg)
	}
}

func dropboxMenu() string {
	fmt.Println("--- ÆCLOUD DROPBOX MENÚ ---")
	fmt.Println("0- VER LISTADO DE FICHEROS DE TU CARPETA DE DROPBOX")
	fmt.Println("1- SUBIR FICHERO A DROPBOX")
	fmt.Println("2- DESCARGAR FICHERO DE DROPBOX")
	fmt.Println("3- CREAR CARPETA PERSONAL DE DROPBOX")
	fmt.Println("Q- SALIR")
	fmt.Print("Opción: ")
	var input string
	fmt.Scanf("%s\n", &input)

	return input
}
