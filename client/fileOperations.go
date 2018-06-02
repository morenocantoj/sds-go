package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type fileStruct struct {
	name      string
	extension string
	filepath  string
	content   []byte
}

type checkFileStruct struct {
	Ok  bool
	Id  int
	Msg string
}

const MAX_PACKAGE_SIZE = 4 * 1000 // 4MB

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

func filePartUpload(client *http.Client, uri string, params map[string]string, paramName string, partFile []byte, partIndex int) (int, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	reader := bytes.NewReader(partFile)

	part, err := writer.CreateFormFile(paramName, strconv.Itoa(partIndex))
	chk(err)

	_, err = io.Copy(part, reader)
	chk(err)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	chk(err)

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+tokenSesion)

	resp, err := client.Do(req)
	chk(err)

	bodyResponse, err := ioutil.ReadAll(resp.Body)
	chk(err)
	resp.Body.Close()

	var data checkFileStruct
	json.Unmarshal(bodyResponse, &data)
	fmt.Printf("%v\n", data.Msg)
	return data.Id, err
}

func checkFileExists(client *http.Client, uri string, checksum string) (bool, int, error) {
	params := map[string]string{
		"checksum": checksum,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+tokenSesion)

	r, err := client.Do(req)
	chk(err)

	bodyResponse, err := ioutil.ReadAll(r.Body)
	chk(err)

	var checkResponse checkFileStruct
	err = json.Unmarshal(bodyResponse, &checkResponse)
	chk(err)
	defer r.Body.Close()

	return checkResponse.Ok, checkResponse.Id, err
}

// ------------------------------
// ------------------------------
// ------------------------------
// FIXME: OLD --> delete on finish
// ------------------------------
// ------------------------------
// ------------------------------
// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contentInBytes, err := ioutil.ReadAll(file)
	fileParts := split(contentInBytes, MAX_PACKAGE_SIZE)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for index, part := range fileParts {
		fmt.Println("PART " + strconv.Itoa(index))
		fmt.Println(part)

		reader := bytes.NewReader(part)

		partWriter, err := writer.CreateFormFile(paramName, strconv.Itoa(index)) //filepath.Base(path))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(partWriter, reader)
	}

	// body := &bytes.Buffer{}
	// writer := multipart.NewWriter(body)
	// part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = io.Copy(part, reader)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+tokenSesion)
	return req, err
}
