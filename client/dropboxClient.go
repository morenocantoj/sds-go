package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func dropboxClient(client *http.Client) {
	optMenu := dropboxMenu()

	for optMenu != "Q" {
		switch optMenu {
		case "0":
			listFilesDropboxClient(client, tokenSesion)
		case "1":
			//TODO: Implement upload menu
		case "2":
			downloadFileDropboxClient(client, tokenSesion)
		case "3":
			createDropboxFolderClient(client, tokenSesion)
		default:
			fmt.Println("Opción incorrecta!")
		}
		optMenu = dropboxMenu()
	}
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
	if err != nil {
		fmt.Println("¡Ha habido un error en la petición!")

	} else {
		// Check if we have files
		if len(filesDropbox.Entries) > 0 {
			fmt.Println("Estos son los ficheros disponibles en Dropbox")

			for i, file := range filesDropbox.Entries {
				fmt.Printf("%d- %s \n", i+1, file.Name)
				fmt.Println("---")
			}
		} else {
			fmt.Println("¡No tienes ficheros subidos en la plataforma!")
		}
	}

}

func downloadFileDropboxClient(client *http.Client, token string) {
	listFilesDropboxClient(client, token)

	fmt.Printf("Introduce el nombre del archivo a descargar: ")
	var filename string
	fmt.Scanf("%s\n", &filename)

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

	if downloadedFile.Downloaded == true {
		fmt.Println("File name " + downloadedFile.Filename)
		// Compare file checksum
		checksumFile := sha256.Sum256(downloadedFile.Content)
		slice := checksumFile[:]
		checksumString := encode64(slice)

		if checksumString == downloadedFile.Checksum {
			// Save file
			err = ioutil.WriteFile("./downloads/"+downloadedFile.Filename, downloadedFile.Content, 0644)
			chk(err)
			fmt.Println("Fichero descargado correctamente en el directorio 'downloads'")

		} else {
			// Corrupted file
			fmt.Println("¡Error! El archivo no concuerda. Puede que el fichero esté corrompido")
		}
	} else {
		// Error downloading file
		fmt.Println("¡Error al descargar el fichero!")
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
