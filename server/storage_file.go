package main

import (
	"crypto/aes"
	"crypto/cipher"
	"io/ioutil"
	"os"
)

// función para cifrar (con AES en este caso), adjunta el IV al principio
func encrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)+16)    // reservamos espacio para el IV al principio
	blk, err := aes.NewCipher(key)      // cifrador en bloque (AES), usa key
	chk(err)                            // comprobamos el error
	ctr := cipher.NewCTR(blk, out[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out[16:], data)    // ciframos los datos
	return
}

// función para descifrar (con AES en este caso)
func decrypt(data, key []byte) (out []byte) {
	out = make([]byte, len(data)-16)     // la salida no va a tener el IV
	blk, err := aes.NewCipher(key)       // cifrador en bloque (AES), usa key
	chk(err)                             // comprobamos el error
	ctr := cipher.NewCTR(blk, data[:16]) // cifrador en flujo: modo CTR, usa IV
	ctr.XORKeyStream(out, data[16:])     // desciframos (doble cifrado) los datos
	return
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

func saveFile(fileData filePartStruct, file_uuid string) (string, error) {
	// TODO: Check if files folder exists (if not create it)
	var dst = "./files/" + file_uuid //+ fileData.extension

	err := ioutil.WriteFile(dst, fileData.content, 0666)
	chk(err)

	return file_uuid, nil
}

func readFile(filename string) ([]byte, error) {
	var src = "./files/" + filename

	// lectura completa de ficheros (precaucion! copia todo el fichero a memoria)
	file, err := ioutil.ReadFile(src)
	chk(err)

	return file, nil
}

func deleteFile(filename string) (bool, error) {
	var src = "./files/" + filename

	err := os.Remove(src)
	chk(err)

	return true, nil
}
