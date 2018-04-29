package main

import (
	"io/ioutil"
	"path/filepath"
)

func readFile(filename string) ([]byte, error) {
	var fileAbsPath string
	isAbsolute := filepath.IsAbs(filename)
	if isAbsolute {
		fileAbsPath = filename
	} else {
		folderPath, err := filepath.Abs("./")
		if err != nil {
			return nil, err
		}
		fileAbsPath = folderPath + "/" + filename
	}
	// lectura completa de ficheros (precaucion! copia todo el fichero a memoria)
	file, err := ioutil.ReadFile(fileAbsPath)
	return file, err
}

// FIXME: Move to server
func saveFile(file []byte) (bool, error) {
	return false, nil
}
