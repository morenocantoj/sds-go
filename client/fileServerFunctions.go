package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func check(e error) {
	if e != nil {
		// TODO: show user an error
		panic(e)
	}
}

func listFiles() {
	fmt.Println("Falta por implementar!!")

	var filename string
	fmt.Printf("Introduce el fichero a subir: ")
	fmt.Scanf("%s\n", &filename)
}

func uploadFile(client *http.Client) {
	var filename string
	fmt.Printf("Introduce el archivo a subir (P.ej.: /Users/username/Desktop/file.txt): ")
	fmt.Scanf("%s\n", &filename)

	fileData, err := readFile(filename)
	if err != nil {
		fmt.Printf("ERROR!! No se encuentra el archivo introducido\n\n")
		return
	}

	// Realizamos la petición
	extraParams := map[string]string{
		"name":      fileData.name,
		"extension": fileData.extension,
		"user":      "n",
	}
	request, err := newfileUploadRequest("https://localhost:10443/files/upload", extraParams, "file", fileData.filepath)
	if err != nil {
		fmt.Printf("ERROR!! Ha fallado la comunicación con el servidor\n\n")
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("ERROR!! Ha fallado la comunicación con el servidor\n\n")
		return
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ERROR!! Ha fallado la comunicación con el servidor\n\n")
			return
		}
		resp.Body.Close()

		var data loginStruct
		json.Unmarshal(body, &data)
		fmt.Printf("\n%v\n\n", data.Msg)
	}
}

func downloadFile() {
	fmt.Println("Falta por implementar!!")

	var fileId string
	fmt.Printf("Introduce el id del fichero a descargar: ")
	fmt.Scanf("%s\n", &fileId)
}
