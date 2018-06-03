package main

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
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

func listFiles(client *http.Client) {
	req, err := http.NewRequest("GET", "https://localhost:10443/files", nil)
	chk(err)
	req.Header.Set("Authorization", "Bearer "+tokenSesion)

	resp, err := client.Do(req)
	chk(err)

	fileList := make([]UserFile, 0)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	json.Unmarshal(body, &fileList)

	// Check if we have files
	if len(fileList) > 0 {
		fmt.Println("\n Estos son los ficheros disponibles:\n")
		fmt.Println("\t ID - Filename")
		fmt.Println("\t-------------------------------------")

		for _, file := range fileList {
			fmt.Printf("\t  %s - %s \n", file.Id, file.Filename)
		}
		fmt.Printf("\n\n")
	} else {
		fmt.Println("¡No tienes ficheros subidos en la plataforma!")
	}
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
	var fileinfo fileInfoStruct
	fileinfo.filename = fileData.name
	fileinfo.extension = fileData.extension

	// -- divide file in parts
	file, err := os.Open(fileData.filepath)
	chk(err)
	defer file.Close()

	contentInBytes, err := ioutil.ReadAll(file)
	chk(err)
	// file checksum
	fileChecksumInBytes := md5.Sum(contentInBytes)
	fileinfo.checksum = hex.EncodeToString(fileChecksumInBytes[:])
	// file size
	stat, err := file.Stat()
	chk(err)
	fileinfo.size = stat.Size()
	// file packages
	fileParts := split(contentInBytes, MAX_PACKAGE_SIZE)
	// --

	var filePartsIds []string

	for index, part := range fileParts {

		// file part checksum
		fmt.Println("Comprobando paquete " + strconv.Itoa(index+1) + "/" + strconv.Itoa(len(fileParts)))

		checksumInBytes := md5.Sum(part)
		partChecksum := hex.EncodeToString(checksumInBytes[:])

		fileExists, partId, err := checkPackageExists(client, "https://localhost:10443/files/checkPackage", partChecksum)
		chk(err)
		if fileExists == false {

			// file part upload
			fmt.Print("Enviando paquete " + strconv.Itoa(index+1) + "... ")

			extraParams := map[string]string{
				"filename": fileData.name,
				"index":    strconv.Itoa(index + 1),
				"checksum": partChecksum,
			}

			// cypher part
			key, err := base32.StdEncoding.DecodeString(userSecretKey)
			chk(err)
			partContentEncrypted := encrypt(part, key)

			// send part
			uploadedPartId, err := filePartUpload(client, "https://localhost:10443/files/uploadPackage", extraParams, "file", partContentEncrypted, index+1)
			chk(err)
			partId = uploadedPartId
		}
		filePartsIds = append(filePartsIds, strconv.Itoa(partId))
	}

	fileinfo.packageIds = filePartsIds

	// file data saved
	fmt.Print("Guardando archivo " + fileinfo.filename + " ... ")

	_, err = saveFileInfo(client, "https://localhost:10443/files/saveFile", fileinfo)
	chk(err)
}

func downloadFile(client *http.Client) {
	var fileId string
	fmt.Printf("-> Introduce el ID del fichero a descargar: ")
	fmt.Scanf("%s\n", &fileId)

	fmt.Print("\nDescargando archivo... ")

	req, err := http.NewRequest("GET", "https://localhost:10443/files/download?file="+fileId, nil)
	chk(err)
	req.Header.Set("Authorization", "Bearer "+tokenSesion)

	resp, err := client.Do(req)
	chk(err)

	bodyResponse, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var downloadFile downloadFileStruct
	json.Unmarshal(bodyResponse, &downloadFile)

	if downloadFile.Ok == true {
		// decypher file
		key, err := base32.StdEncoding.DecodeString(userSecretKey)
		chk(err)
		fileContentDecrypted := decrypt(downloadFile.FileContent, key)

		saveFile(fileContentDecrypted, downloadFile.FileName)
	}

	fmt.Println(downloadFile.Msg + "\n")
}

func deleteFile(client *http.Client) {
	var fileId string
	fmt.Printf("-> Introduce el ID del fichero a borrar: ")
	fmt.Scanf("%s\n", &fileId)

	fmt.Print("\nBorrando archivo... ")

	req, err := http.NewRequest("DELETE", "https://localhost:10443/files/delete?file="+fileId, nil)
	chk(err)
	req.Header.Set("Authorization", "Bearer "+tokenSesion)

	resp, err := client.Do(req)
	chk(err)

	bodyResponse, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var deleteFile deleteFileStruct
	json.Unmarshal(bodyResponse, &deleteFile)

	fmt.Println(deleteFile.Msg + "\n")
}
