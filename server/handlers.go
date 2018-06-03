package main

import (
	"database/sql"
	"encoding/base32"
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

func chkErrorPackageUpload(err error, w http.ResponseWriter) {
	if err != nil {
		log.Fatal(err)
		response(w, false, "[server] Se ha producido un error al subir el paquete")
	}
}

func chkErrorFileSave(err error, w http.ResponseWriter) {
	if err != nil {
		log.Fatal(err)
		response(w, false, "[server] Se ha producido un error al guardar el archivo")
	}
}

func handlerPackageUpload(w http.ResponseWriter, req *http.Request) {
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
		chkErrorPackageUpload(err, w)
		userId := getUserIdFromToken(bearerToken)

		// File data
		uid, err := uuid.NewV4()
		chkErrorPackageUpload(err, w)

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
		chkErrorPackageUpload(err, w)

		r := uploadPackageResponse{Ok: true, Id: insertedPackageId, Msg: "Listo"}
		rJSON, err := json.Marshal(&r)
		chkErrorPackageUpload(err, w)
		w.Write(rJSON)
	}
}

func handlerPackageCheck(w http.ResponseWriter, req *http.Request) {
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

func handlerFileSave(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		io.WriteString(w, "Only POST is supported!")
		return
	}
	var fileinfo fileInfoStruct

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
			case "filename":
				fileinfo.filename = string(slurp)
			case "extension":
				fileinfo.extension = string(slurp)
			case "checksum":
				fileinfo.checksum = string(slurp)
			case "size":
				filesize := string(slurp)
				fileinfo.size, _ = strconv.Atoi(filesize)
			case "packageIds":
				packageIdsInString := string(slurp)
				fileinfo.packageIds = strings.Split(packageIdsInString, ",")
			default:
			}

		}
	}

	r := saveFileResponse{Ok: false, Msg: ""}

	// Get actual user ID
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chkErrorFileSave(err, w)
	userId := getUserIdFromToken(bearerToken)

	// Save user file register on DB
	userFileId, err := checkFileExistsForUser(userId, fileinfo.filename)
	chkErrorFileSave(err, w)
	if userFileId == -1 {
		var newUserFile user_file
		newUserFile.userId = userId
		newUserFile.filename = fileinfo.filename
		newUserFile.extension = fileinfo.extension

		insertedUserFileId, err := insertUserFile(newUserFile)
		chkErrorFileSave(err, w)
		userFileId = insertedUserFileId
	}

	// Save file version on DB
	lastFileVersion, err := checkUserFileLastVersion(userFileId)
	chkErrorFileSave(err, w)

	hasUpdates, err := checkLastFileVersionHasUpdates(userFileId, lastFileVersion, fileinfo.checksum)
	chkErrorFileSave(err, w)

	if hasUpdates == true {
		var newFile file
		newFile.user_file_id = userFileId
		newFile.packages_num = len(fileinfo.packageIds)
		newFile.checksum = fileinfo.checksum
		if lastFileVersion == -1 {
			newFile.version = 1
		} else {
			newFile.version = lastFileVersion + 1
		}
		newFile.size = fileinfo.size

		insertedFileId, err := insertFile(newFile)
		chkErrorFileSave(err, w)

		for index, packageIdInString := range fileinfo.packageIds {
			packageId, err := strconv.Atoi(packageIdInString)
			chkErrorFileSave(err, w)

			var newFilePackage file_package
			newFilePackage.file_id = insertedFileId
			newFilePackage.package_id = packageId
			newFilePackage.package_index = index + 1

			_, err = insertFilePackage(newFilePackage)
			chkErrorFileSave(err, w)

			r.Ok = true
			r.Msg = "Listo"
		}
	} else {
		r.Ok = true
		r.Msg = "El archivo ya existe"
	}

	rJSON, err := json.Marshal(&r) // codificamos en JSON
	chk(err)                       // comprobamos error
	w.Write(rJSON)
}

func handlerFileList(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	userId := strconv.Itoa(getUserIdFromToken(bearerToken))

	// Open db
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("handlerFileList", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	rows, err := db.Query("SELECT id, filename FROM user_files WHERE userId = ?", userId)
	chk(err)

	// Get column names
	columns, err := rows.Columns()
	chk(err)

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var files []fileEnumStruct
	// Fetch rows
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		chk(err)

		// Create a file struct and append
		file := fileEnumStruct{Id: string(values[0]), Filename: string(values[1])}
		files = append(files, file)
	}
	defer db.Close()
	fmt.Printf("Slice: %v\n", files)

	responseFilesList(w, files)
}

func handlerFileDownload(w http.ResponseWriter, req *http.Request) {
	response := downloadFileResponse{Ok: false, Msg: "", FileContent: nil, FileName: "", Checksum: ""}
	if req.Method != "GET" {
		response.Msg = "Only GET is supported!"
		rJSON, err := json.Marshal(&response) // codificamos en JSON
		chk(err)                              // comprobamos error
		w.Write(rJSON)
		return
	}

	userFileIdInString := req.URL.Query().Get("file")
	userFileId, err := strconv.Atoi(userFileIdInString)

	// Get actual user ID
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chkErrorPackageUpload(err, w)
	userId := getUserIdFromToken(bearerToken)

	lastFileVersion, err := checkUserFileLastVersion(userFileId)
	chk(err)
	if lastFileVersion == -1 {
		response.Msg = "ERROR! No se encuentra el archivo introducido"
		rJSON, err := json.Marshal(&response) // codificamos en JSON
		chk(err)                              // comprobamos error
		w.Write(rJSON)
		return
	}

	fileId, err := getFileId(userFileId, lastFileVersion)
	chk(err)

	packageIds, err := getFilePackages(fileId)
	chk(err)

	var fileContent []byte
	//packages := make(map[int][]byte)

	for _, packageId := range packageIds {
		packageUuid, uploadUserId, err := getPackageUuid(packageId)
		chk(err)

		secretKey, err := getUserSecretKeyById(uploadUserId)
		chk(err)

		packageContent, err := readFile(packageUuid)
		chk(err)

		// decrypt package with user's secret key who's uploaded
		key, err := base32.StdEncoding.DecodeString(secretKey)
		chk(err)
		packageContentDecryted := decrypt(packageContent, key)

		fileContent = append(fileContent, packageContentDecryted...)
		//packages[packageIndex] = packageContent
	}

	// encrypt file content with requested user's secret key
	userSecretKey, err := getUserSecretKeyById(userId)
	chk(err)
	userKey, err := base32.StdEncoding.DecodeString(userSecretKey)
	chk(err)
	fileContentEncrypted := encrypt(fileContent, userKey)

	fileChecksum, err := getFileChecksum(userFileId, lastFileVersion)
	chk(err)
	fileName, err := getUserFileName(userFileId)
	chk(err)

	response.Ok = true
	response.Msg = "Listo"
	response.FileContent = fileContentEncrypted
	response.FileName = fileName
	response.Checksum = fileChecksum

	rJSON, err := json.Marshal(&response) // codificamos en JSON
	chk(err)                              // comprobamos error
	w.Write(rJSON)
}

func handlerFileDelete(w http.ResponseWriter, req *http.Request) {
	response := downloadFileResponse{Ok: false, Msg: ""}
	if req.Method != "DELETE" {
		response.Msg = "Only DELETE is supported!"
		rJSON, err := json.Marshal(&response) // codificamos en JSON
		chk(err)                              // comprobamos error
		w.Write(rJSON)
		return
	}

	userFileIdInString := req.URL.Query().Get("file")
	userFileId, err := strconv.Atoi(userFileIdInString)

	fileVersionIds, err := checkUserFileLastVersion(userFileId)
	chk(err)
	if fileVersionIds == -1 {
		response.Msg = "ERROR! No se encuentra el archivo introducido"
		rJSON, err := json.Marshal(&response) // codificamos en JSON
		chk(err)                              // comprobamos error
		w.Write(rJSON)
		return
	}

	response.Ok = true
	response.Msg = "Listo"

	rJSON, err := json.Marshal(&response) // codificamos en JSON
	chk(err)                              // comprobamos error
	w.Write(rJSON)
}
