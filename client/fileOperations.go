package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func readFile(filename string) (string, error) {
	var fileAbsPath string
	isAbsolute := filepath.IsAbs(filename)
	if isAbsolute {
		fileAbsPath = filename
	} else {
		folderPath, err := filepath.Abs("./")
		if err != nil {
			return "", err
		}
		fileAbsPath = folderPath + "/" + filename
	}
	return fileAbsPath, nil
	// lectura completa de ficheros (precaucion! copia todo el fichero a memoria)
	// file, err := ioutil.ReadFile(fileAbsPath)
	// return file, err
}

// FIXME: Move to server
func saveFile(filePath string) (bool, error) {

	from, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	defer from.Close()

	var filename = filepath.Base(filePath)
	var dst = "../server/files/" + filename

	to, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	return true, nil
}
