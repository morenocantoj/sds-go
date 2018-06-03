package main

// fileOperations.go
type fileInfoStruct struct {
	filename   string
	extension  string
	packageIds []string
	checksum   string
	size       int64
}

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

type saveFileStruct struct {
	Ok  bool
	Msg string
}

type downloadFileStruct struct {
	Ok           bool
	Msg          string
	FileChecksum string
	FileContent  []byte
	FileName     string
}

type deleteFileStruct struct {
	Ok  bool
	Msg string
}

// client.go
type loginStruct struct {
	Ok        bool
	TwoFa     bool
	Msg       string
	Token     string
	SecretKey string
}

type twoFactorStruct struct {
	Ok    bool
	Token string
}

type registerStruct struct {
	Ok  bool
	Msg string
}

type JwtToken struct {
	Token string `json:"token"`
}

type UserFile struct {
	Id       string
	Filename string
}

type createDropboxFolder struct {
	Created bool
	Msg     string
}

type DropboxDownload struct {
	Downloaded bool
	Content    []byte
	Filename   string
	Checksum   string
}

type fileEnumDropboxStruct struct {
	Tag             string      `json:".tag"`
	Name            string      `json:"name"`
	Id              string      `json:"id"`
	Client_modified string      `json:"client_modified"`
	Server_modified string      `json:"server_modified"`
	Rev             string      `json:"rev"`
	Size            int         `json:"size"`
	Path_lower      string      `json:"path_lower"`
	Path_display    string      `json:"path_display"`
	Sharing_info    interface{} `json:"sharing_info"`
	Property_groups interface{} `json:"property_groups"`
	Shared          bool        `json:"has_explicit_shared_members"`
	Content_hash    string      `json:"content_hash"`
}

type fileListDropbox struct {
	Entries  []fileEnumDropboxStruct `json:"entries"`
	Cursor   string                  `json:"cursor"`
	Has_more bool                    `json:"has_more"`
}

type uploadFileDropbox struct {
	Uploaded bool
	Msg      string
}

type tokenValid struct {
	Code int
}
