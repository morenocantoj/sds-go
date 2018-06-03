package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

// función para cifrar (con AES en este caso), adjunta el IV al principio
func encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)    // reservamos espacio para el IV al principio
	blk, err := aes.NewCipher(key)      // cifrador en bloque (AES), usa key
	chk(err)                            // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)    // ciframos los datos
	return
}

// función para descifrar (con AES en este caso)
func decrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)     // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)       // cifrador en bloque (AES), usa key
	chk(err)                             // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])     // desciframos (doble cifrado) los datos
	return
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

func checkPackageExists(client *http.Client, uri string, checksum string) (bool, int, error) {
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

func saveFileInfo(client *http.Client, uri string, fileinfo fileInfoStruct) (bool, error) {
	params := map[string]string{
		"filename":   fileinfo.filename,
		"extension":  fileinfo.extension,
		"checksum":   fileinfo.checksum,
		"size":       strconv.Itoa(int(fileinfo.size)),
		"packageIds": strings.Join(fileinfo.packageIds, ","),
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

	var saveFileResponse saveFileStruct
	err = json.Unmarshal(bodyResponse, &saveFileResponse)
	chk(err)
	fmt.Printf("%v\n\n", saveFileResponse.Msg)
	defer r.Body.Close()

	return saveFileResponse.Ok, err
}
