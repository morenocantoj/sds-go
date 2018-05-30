package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"

	uuid "github.com/satori/go.uuid"
)

type fileStruct struct {
	name      string
	extension string
	content   []byte
}

func saveFile(fileData fileStruct) (bool, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	filename := uid.String()

	// TODO: Check if files folder exists (if not create it)
	var dst = "./files/" + filename + fileData.extension //fileData.name

	fmt.Printf("%x", md5.Sum(fileData.content))
	err = ioutil.WriteFile(dst, fileData.content, 0666)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	return true, nil
}
