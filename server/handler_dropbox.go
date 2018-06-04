package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

/**
 * Server Response Functions
 */
func responseCreateDropboxFolder(w io.Writer, created bool, msg string) {
	r := respCreateDropboxFolder{Created: created, Msg: msg}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func responseDownloadDropboxfile(w io.Writer, downloaded bool, content []byte, filename string, checksum string) {
	r := DropboxDownloadResponse{Downloaded: downloaded, Content: content, Filename: filename, Checksum: checksum}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

func responseListFilesDropbox(w io.Writer, fileList fileListDropbox) {
	rJSON, err := json.Marshal(&fileList)
	chk(err)
	w.Write(rJSON)
}

func responseUploadFileDropbox(w io.Writer, uploaded bool, msg string) {
	r := uploadFileDropboxStruct{Uploaded: uploaded, Msg: msg}
	rJSON, err := json.Marshal(&r)
	chk(err)
	w.Write(rJSON)
}

/**
 * DROPBOX HANDLERS FUNCTIONS
 */

func createDropboxFolder(w http.ResponseWriter, req *http.Request) {
	// Get user id by token and use it for create the new folder
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))
	loginfo("createDropboxFolder", "Usuario "+folderId+" crea carpeta personal", "Dropbox API", "info", nil)

	clientDropbox := &http.Client{}

	var jsonStr = []byte(`{
    "path": "/` + folderId + `",
    "autorename": false
		}`)

	body := bytes.NewBuffer(jsonStr)

	req, _ = http.NewRequest("POST", "https://api.dropboxapi.com/2/files/create_folder_v2", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)

	resp, err := clientDropbox.Do(req)
	chk(err)

	b, _ := ioutil.ReadAll(resp.Body)
	respDropbox := string(b)

	if strings.Contains(respDropbox, "metadata") && strings.Contains(respDropbox, "id") {
		// Folder created successfully
		responseCreateDropboxFolder(w, true, "¡Carpeta creada correctamente!")
	} else if strings.Contains(respDropbox, "error") {
		responseCreateDropboxFolder(w, false, "¡Carpeta personal ya existente!")
	} else {
		responseCreateDropboxFolder(w, false, "¡Error al crear carpeta personal!")
	}
}

func downloadFileDropbox(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))
	queryParams := req.URL.Query()
	filename := queryParams.Get("filename")

	loginfo("downloadfileDropbox", "Usuario "+folderId+" intenta bajar archivo "+filename, "Dropbox API", "info", nil)

	// Make request for dropbox
	clientDropbox := &http.Client{}
	fullPath := "/" + folderId + "/" + filename

	dropboxHeader := `{"path": "` + fullPath + `"}`

	req, err = http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
	chk(err)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)
	req.Header.Set("Dropbox-API-Arg", dropboxHeader)

	resp, err := clientDropbox.Do(req)
	chk(err)

	// Get body response
	b, _ := ioutil.ReadAll(resp.Body)
	contentFile := b

	jsonResult := DropboxDownloadStruct{}
	jsonString := resp.Header.Get("dropbox-api-result")

	err = json.Unmarshal([]byte(jsonString), &jsonResult)
	if err != nil {
		// Fail downloading file
		responseDownloadDropboxfile(w, false, contentFile, "err", "err")

	} else {
		// Correct response

		// Checksum
		checksumFile := sha256.Sum256(contentFile)
		slice := checksumFile[:]
		checksumString := encode64(slice)

		responseDownloadDropboxfile(w, true, contentFile, filename, checksumString)
	}
}

//Lists all dropbox files by user
func listFilesDropbox(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))
	loginfo("listFilesDropbox", "Usuario "+folderId+" intenta listar sus archivos", "Dropbox API", "info", nil)

	// Make dropbox Request
	clientDropbox := &http.Client{}

	var jsonStr = []byte(`{
		"path": "/` + folderId + `",
		"recursive": false,
		"include_media_info": false,
    "include_deleted": false,
    "include_has_explicit_shared_members": false,
    "include_mounted_folders": true
		}`)

	body := bytes.NewBuffer(jsonStr)

	req, err = http.NewRequest("POST", "https://api.dropboxapi.com/2/files/list_folder", body)
	chk(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)

	resp, err := clientDropbox.Do(req)
	chk(err)

	b, err := ioutil.ReadAll(resp.Body)
	chk(err)

	var filesDropbox fileListDropbox
	err = json.Unmarshal(b, &filesDropbox)
	chk(err)

	responseListFilesDropbox(w, filesDropbox)
}

func uploadFileDropbox(w http.ResponseWriter, req *http.Request) {
	// Get user params
	bearerToken, err := GetBearerToken(req.Header.Get("Authorization"))
	chk(err)
	folderId := strconv.Itoa(getUserIdFromToken(bearerToken))

	loginfo("uploadFileDropbox", "Usuario "+folderId+" intenta subir un archivo a Dropbox", "Dropbox API", "info", nil)

	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	chk(err)

	var fileData []byte
	var checksum string
	var filename string

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(req.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			chk(err)

			slurp, err := ioutil.ReadAll(p)
			chk(err)

			switch p.FormName() {
			case "file":
				// Send file to dropbox
				fileData = slurp
			case "filename":
				filename = string(slurp)
			case "checksum":
				checksum = string(slurp)
			default:
			}
		}

		// File checksum
		checksumFile := sha256.Sum256(fileData)
		slice := checksumFile[:]
		checksumString := encode64(slice)

		if checksumString != checksum {
			// Not the same file
			responseUploadFileDropbox(w, false, "¡Error! El fichero recibido no concuerda con el enviado")

		} else {
			// Upload file to dropbox
			// Make request for dropbox
			clientDropbox := &http.Client{}
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			reader := bytes.NewReader(fileData)

			part, err := writer.CreateFormFile("prueba.txt", "prueba.txt")
			chk(err)

			_, err = io.Copy(part, reader)
			chk(err)

			err = writer.Close()
			chk(err)

			path := "/" + folderId + "/" + filename
			dropboxHeader := `{"path": "` + path + `", "mode": "add", "autorename": true, "mute": false}`

			req, err = http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", body)

			chk(err)
			req.Header.Add("Content-Type", writer.FormDataContentType())
			req.Header.Add("Content-Type", "application/octet-stream")
			req.Header.Set("Authorization", "Bearer "+DROPBOX_TOKEN)
			req.Header.Set("Dropbox-API-Arg", dropboxHeader)

			resp, err := clientDropbox.Do(req)
			chk(err)

			// Get body response
			b, _ := ioutil.ReadAll(resp.Body)
			respDropbox := string(b)

			if strings.Contains(respDropbox, "name") && strings.Contains(respDropbox, "content_hash") && strings.Contains(respDropbox, "id") {
				responseUploadFileDropbox(w, true, "¡Fichero subido a dropbox correctamente!")
			} else {
				responseUploadFileDropbox(w, false, "¡Error al subir el fichero! Inténtalo de nuevo")
			}
		}

	}
}
