package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

func handlerFileUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		io.WriteString(w, "Only POST is supported!")
		return
	}

	var data fileStruct

	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(req.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			slurp, err := ioutil.ReadAll(p)
			if err != nil {
				log.Fatal(err)
			}

			switch p.FormName() {
			case "file":
				data.content = slurp
				// Check if is file content
				contentType := p.Header.Get("Content-type")
				if contentType == "application/octet-stream" {
				}
			case "name":
				data.name = string(slurp)
			case "extension":
				data.extension = string(slurp)
			case "user":
			default:
			}

		}

		isSaved, err := saveFile(data)
		if err != nil || isSaved == false {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el archivo")
		}
		fmt.Printf("Archivo \"%v\" subido correctamente\n", data.name)
		response(w, true, "[server] Archivo subido")
	}
}
