package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
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

		// Get actual user ID
		bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
		if err != nil {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el archivo")
		}
		userId := getUserIdFromToken(bearerToken)

		// Save user file register on DB
		userFileId, err := checkFileExistsForUser(userId, data.name)
		if err != nil {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el archivo")
		}
		if userFileId == -1 {
			var newUserFile user_file
			newUserFile.userId = userId
			newUserFile.filename = data.name
			newUserFile.extension = data.extension

			insertedUserFileId, err := insertUserFile(newUserFile)
			if err != nil {
				log.Fatal(err)
				response(w, false, "[server] Se ha producido un error al subir el archivo")
			}
			userFileId = insertedUserFileId
		}

		// Save file
		var newFile file

		uid, err := uuid.NewV4()
		if err != nil {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el archivo")
		}
		newFile.uuid = uid.String()

		checksumInBytes := md5.Sum(data.content)
		newFile.checksum = hex.EncodeToString(checksumInBytes[:])

		fileId, err := checkFileExistsInDatabase(newFile.checksum)
		if err != nil {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el archivo")
		}
		if fileId == -1 {
			// Save file on DB Storage
			saved_uuid, err := saveFile(data, newFile.uuid)
			if err != nil || saved_uuid == "" {
				log.Fatal(err)
				response(w, false, "[server] Se ha producido un error al subir el archivo")
			}
			fmt.Printf("Archivo \"%v\" subido correctamente\n", data.name)

			// Save file on DB
			insertedFileId, err := insertFileInDatabase(newFile)
			if err != nil {
				log.Fatal(err)
				response(w, false, "[server] Se ha producido un error al subir el archivo")
			}
			fileId = insertedFileId
		}

		// Save file version on DB
		lastFileVersion, err := checkUserFileLastVersion(userFileId)
		if err != nil {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el archivo")
		}

		hasUpdates := true
		var newFileVersion file_version
		newFileVersion.user_file_id = userFileId
		newFileVersion.file_id = fileId
		if lastFileVersion == -1 {
			newFileVersion.version_num = 1
		} else {
			newFileVersion.version_num = lastFileVersion + 1
			hasUpdates, err = checkLastFileVersionHasUpdates(userFileId, lastFileVersion, newFile.checksum)
			if err != nil {
				log.Fatal(err)
				response(w, false, "[server] Se ha producido un error al subir el archivo")
			}
		}

		if hasUpdates == true {
			_, err = insertFileVersion(newFileVersion)
			if err != nil {
				log.Fatal(err)
				response(w, false, "[server] Se ha producido un error al subir el archivo")
			}

			response(w, true, "[server] Archivo subido")
		} else {
			response(w, true, "[server] El archivo ya existe")
		}

	}
}
