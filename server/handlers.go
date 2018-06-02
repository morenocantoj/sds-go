package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

type checkFileResponse struct {
	Ok  bool
	Id  int
	Msg string
}

type uploadPackageResponse struct {
	Ok  bool
	Id  int
	Msg string
}

func chkErrorFileUpload(err error, w http.ResponseWriter) {
	if err != nil {
		log.Fatal(err)
		response(w, false, "[server] Se ha producido un error al subir el paquete")
	}
}

func handlerFileUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		io.WriteString(w, "Only POST is supported!")
		return
	}

	var data filePartStruct

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
				fmt.Println("[server] File Part received: " + p.FileName())
			case "filename":
				data.filename = string(slurp)
			case "index":
				index, err := strconv.Atoi(string(slurp))
				chk(err)
				data.index = index
			case "checksum":
				data.checksum = string(slurp)
			default:
			}

		}

		// Get actual user ID
		bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
		chkErrorFileUpload(err, w)
		userId := getUserIdFromToken(bearerToken)

		// File data
		uid, err := uuid.NewV4()
		chkErrorFileUpload(err, w)

		var newFile packageFile
		newFile.uuid = uid.String()
		newFile.checksum = data.checksum
		newFile.upload_user_id = userId

		// Save package on DB Storage
		saved_uuid, err := saveFile(data, newFile.uuid)
		if err != nil || saved_uuid == "" {
			log.Fatal(err)
			response(w, false, "[server] Se ha producido un error al subir el paquete")
		}
		fmt.Printf("Paquete %v de archivo \"%v\" subido correctamente\n", data.index, data.filename)

		// Save file on DB
		insertedPackageId, err := insertPackageInDatabase(newFile)
		chkErrorFileUpload(err, w)
		fmt.Println("-------")
		fmt.Println(insertedPackageId)
		r := uploadPackageResponse{Ok: true, Id: insertedPackageId, Msg: "Subido"}
		rJSON, err := json.Marshal(&r)
		chkErrorFileUpload(err, w)
		w.Write(rJSON)
	}
}

func handlerFileCheck(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		io.WriteString(w, "Only POST is supported!")
		return
	}
	var checksum string

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
			case "checksum":
				checksum = string(slurp)
			default:
			}

		}
	}

	partId := -1
	r := checkFileResponse{Ok: false, Id: partId, Msg: ""}

	partId, err = checkPackageExistsInDatabase(checksum)
	chk(err)
	if partId != -1 {
		r.Ok = true
		r.Id = partId
		r.Msg = "El archivo existe"
	}

	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)
}

// --------
// -------
// FIXME: delete! it's OLD
func handlerFileUploadOld(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		io.WriteString(w, "Only POST is supported!")
		return
	}

	var data fileStruct

	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}
	var fileParts [][]byte

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
				fileParts = append(fileParts, slurp)
				fmt.Println(p.FileName())
			case "name":
				data.name = string(slurp)
			case "extension":
				data.extension = string(slurp)
			default:
			}

		}

		for index, part := range fileParts {
			fmt.Println("PART " + strconv.Itoa(index))
			fmt.Println(string(part[:]))
		}

		// // Get actual user ID
		// bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
		// if err != nil {
		// 	log.Fatal(err)
		// 	response(w, false, "[server] Se ha producido un error al subir el archivo")
		// }
		// userId := getUserIdFromToken(bearerToken)
		//
		// // Save user file register on DB
		// userFileId, err := checkFileExistsForUser(userId, data.name)
		// if err != nil {
		// 	log.Fatal(err)
		// 	response(w, false, "[server] Se ha producido un error al subir el archivo")
		// }
		// if userFileId == -1 {
		// 	var newUserFile user_file
		// 	newUserFile.userId = userId
		// 	newUserFile.filename = data.name
		// 	newUserFile.extension = data.extension
		//
		// 	insertedUserFileId, err := insertUserFile(newUserFile)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		response(w, false, "[server] Se ha producido un error al subir el archivo")
		// 	}
		// 	userFileId = insertedUserFileId
		// }
		//
		// // Save file
		// var newFile file
		//
		// uid, err := uuid.NewV4()
		// if err != nil {
		// 	log.Fatal(err)
		// 	response(w, false, "[server] Se ha producido un error al subir el archivo")
		// }
		// newFile.uuid = uid.String()
		//
		// checksumInBytes := md5.Sum(data.content)
		// newFile.checksum = hex.EncodeToString(checksumInBytes[:])
		//
		// fileId, err := checkFileExistsInDatabase(newFile.checksum)
		// if err != nil {
		// 	log.Fatal(err)
		// 	response(w, false, "[server] Se ha producido un error al subir el archivo")
		// }
		// if fileId == -1 {
		// 	// Save file on DB Storage
		// 	saved_uuid, err := saveFile(data, newFile.uuid)
		// 	if err != nil || saved_uuid == "" {
		// 		log.Fatal(err)
		// 		response(w, false, "[server] Se ha producido un error al subir el archivo")
		// 	}
		// 	fmt.Printf("Archivo \"%v\" subido correctamente\n", data.name)
		//
		// 	// Save file on DB
		// 	insertedFileId, err := insertFileInDatabase(newFile)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		response(w, false, "[server] Se ha producido un error al subir el archivo")
		// 	}
		// 	fileId = insertedFileId
		// }
		//
		// // Save file version on DB
		// lastFileVersion, err := checkUserFileLastVersion(userFileId)
		// if err != nil {
		// 	log.Fatal(err)
		// 	response(w, false, "[server] Se ha producido un error al subir el archivo")
		// }
		//
		// hasUpdates := true
		// var newFileVersion file_version
		// newFileVersion.user_file_id = userFileId
		// newFileVersion.file_id = fileId
		// if lastFileVersion == -1 {
		// 	newFileVersion.version_num = 1
		// } else {
		// 	newFileVersion.version_num = lastFileVersion + 1
		// 	hasUpdates, err = checkLastFileVersionHasUpdates(userFileId, lastFileVersion, newFile.checksum)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		response(w, false, "[server] Se ha producido un error al subir el archivo")
		// 	}
		// }
		//
		// if hasUpdates == true {
		// 	_, err = insertFileVersion(newFileVersion)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 		response(w, false, "[server] Se ha producido un error al subir el archivo")
		// 	}
		//
		// 	response(w, true, "[server] Archivo subido")
		// } else {
		// 	response(w, true, "[server] El archivo ya existe")
		// }
		response(w, true, "[server] oki")
	}
}
