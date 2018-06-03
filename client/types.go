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
