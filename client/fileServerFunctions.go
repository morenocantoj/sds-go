package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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

	// User File Data (saved on DB in case of success)
	// user_file_data := map[string]string{
	// 	"name":      fileData.name,
	// 	"extension": fileData.extension,
	// }

	// -- divide file in parts
	file, err := os.Open(fileData.filepath)
	chk(err)
	defer file.Close()

	contentInBytes, err := ioutil.ReadAll(file)
	chk(err)
	// fileChecksumInBytes := md5.Sum(contentInBytes)
	// fileChecksumString := hex.EncodeToString(fileChecksumInBytes[:])

	fileParts := split(contentInBytes, MAX_PACKAGE_SIZE)
	// --

	var filePartsIds []int

	for index, part := range fileParts {

		// file part checksum
		fmt.Println("Comprobando paquete " + strconv.Itoa(index+1) + "/" + strconv.Itoa(len(fileParts)))

		checksumInBytes := md5.Sum(part)
		partChecksum := hex.EncodeToString(checksumInBytes[:])

		fileExists, partId, err := checkFileExists(client, "https://localhost:10443/files/check", partChecksum)
		chk(err)
		if fileExists == false {

			// file part upload
			fmt.Println("Enviando paquete " + strconv.Itoa(index+1) + " ...")

			extraParams := map[string]string{
				"filename": fileData.name,
				"index":    strconv.Itoa(index + 1),
				"checksum": partChecksum,
			}

			uploadedPartId, err := filePartUpload(client, "https://localhost:10443/files/upload", extraParams, "file", part, index+1)
			chk(err)
			partId = uploadedPartId
		}
		filePartsIds = append(filePartsIds, partId)
	}

	// TODO: Enviar datos del archivo
	fmt.Println("--------------")
	fmt.Println(filePartsIds)

}

func downloadFile() {
	fmt.Println("Falta por implementar!!")

	var fileId string
	fmt.Printf("Introduce el id del fichero a descargar: ")
	fmt.Scanf("%s\n", &fileId)
}
