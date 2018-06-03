package main

import (
	"io/ioutil"
	"path/filepath"
)

func readFile(inputFile string) (fileStruct, error) {
	var data fileStruct
	var fileAbsPath string
	isAbsolute := filepath.IsAbs(inputFile)
	if isAbsolute {
		fileAbsPath = inputFile
	} else {
		folderPath, err := filepath.Abs("./")
		if err != nil {
			return data, err
		}
		fileAbsPath = folderPath + "/" + inputFile
	}

	// lectura completa de ficheros (precaucion! copia todo el fichero a memoria)
	file, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return data, err
	}

	var filename = filepath.Base(fileAbsPath)
	var extension = filepath.Ext(filename)

	// Object file
	data.name = filename
	data.extension = extension
	data.filepath = fileAbsPath
	data.content = file

	return data, err
}

func saveFile(fileContent []byte, filename string) {
	// TODO: Check if files folder exists (if not create it)
	var dst = "./downloads/" + filename

	err := ioutil.WriteFile(dst, fileContent, 0666)
	chk(err)
}
