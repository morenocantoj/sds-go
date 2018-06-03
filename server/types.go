package main

// db_file.go
type user_file struct {
	userId    int
	filename  string
	extension string
}

type file struct {
	id           int
	user_file_id int
	packages_num int
	checksum     string
	version      int
	size         int
	timestamp    string
}

type file_package struct {
	id            int
	file_id       int
	package_id    int
	package_index int
}

type packageFile struct {
	id             int
	uuid           string
	checksum       string
	upload_user_id int
	timestamp      string
}

// handlers.go
type fileInfoStruct struct {
	filename   string
	extension  string
	packageIds []string
	checksum   string
	size       int
}

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

type saveFileResponse struct {
	Ok  bool
	Msg string
}

type downloadFileResponse struct {
	Ok          bool
	Msg         string
	FileContent []byte
	FileName    string
	Checksum    string
}

type deleteFileResponse struct {
	Ok  bool
	Msg string
}

// storage.go
type filePartStruct struct {
	filename string
	index    int
	checksum string
	content  []byte
}

type fileStruct struct {
	name      string
	extension string
	content   []byte
}

// server.go

type JwtToken struct {
	Token string `json:"token"`
}

// respuesta del servidor
type resp struct {
	Ok  bool   // true -> correcto, false -> error
	Msg string // mensaje adicional
}

type respLogin struct {
	Ok        bool   // true -> correcto, false -> error
	TwoFa     bool   // Two Factor enabled
	Msg       string // mensaje adicional
	Token     string
	SecretKey string
}

type user struct {
	username string
	password string
}

type Exception struct {
	Message string `json:"message"`
}

type OtpToken struct {
	Token string
}

type twoFactorStruct struct {
	Ok    bool
	Token string
}

type fileEnumStruct struct {
	Id       string
	Filename string
}

type fileList []fileEnumStruct
