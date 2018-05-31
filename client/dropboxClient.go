package main

import (
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
			//TODO: Implement list files menu
		case "1":
			//TODO: Implement upload menu
		case "2":
			//TODO: Implement download menu
		case "3":
			createDropboxFolderClient(client, tokenSesion)
		default:
			fmt.Println("Opción incorrecta!")
		}
		optMenu = dropboxMenu()
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
