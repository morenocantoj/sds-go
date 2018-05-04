package main

import (
	"fmt"
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

func uploadFile() {
	fmt.Println("Falta por implementar!!")

	var filename string
	fmt.Printf("Introduce el archivo a subir (P.ej.: /Users/username/Desktop/file.txt): ")
	fmt.Scanf("%s\n", &filename)

	filePath, err := readFile(filename)
	if err != nil {
		fmt.Printf("ERROR!! No se encuentra el archivo introducido\n\n")
		return
	}

	// FIXME: Move to server
	isSaved, err := saveFile(filePath)
	if err != nil {
		fmt.Printf("ERROR!! Se ha producido un error almacenando el archivo\n\n")
		return
	}
	if isSaved {
		fmt.Printf("El archivo se ha subido correctamente\n\n")
	}
}

func downloadFile() {
	fmt.Println("Falta por implementar!!")

	var fileId string
	fmt.Printf("Introduce el id del fichero a descargar: ")
	fmt.Scanf("%s\n", &fileId)
}
