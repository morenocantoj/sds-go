package main

import (
	"io/ioutil"
)

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

func saveFile(fileData filePartStruct, file_uuid string) (string, error) {
	// TODO: Check if files folder exists (if not create it)
	var dst = "./files/" + file_uuid //+ fileData.extension

	err := ioutil.WriteFile(dst, fileData.content, 0666)
	chk(err)

	return file_uuid, nil
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}
