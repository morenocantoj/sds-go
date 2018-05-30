package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
)

type fileStruct struct {
	name      string
	extension string
	content   []byte
}

func saveFile(fileData fileStruct, file_uuid string) (string, error) {
	// TODO: Check if files folder exists (if not create it)
	var dst = "./files/" + file_uuid + fileData.extension

	fmt.Printf("%x", md5.Sum(fileData.content))
	err := ioutil.WriteFile(dst, fileData.content, 0666)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return file_uuid, nil
}
