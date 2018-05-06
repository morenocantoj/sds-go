package main

import (
	"io/ioutil"
	"log"
)

type fileStruct struct {
	name    string
	content []byte
}

func saveFile(fileData fileStruct) (bool, error) {

	var dst = "./files/" + fileData.name

	err := ioutil.WriteFile(dst, fileData.content, 0666)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	return true, nil
}
