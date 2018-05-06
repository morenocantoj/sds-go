package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type fileStruct struct {
	name     string
	filepath string
	content  []byte
}

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

	// Object file
	data.name = filename
	data.filepath = fileAbsPath
	data.content = file
	return data, err
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
