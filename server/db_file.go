package main

import (
	"database/sql"
	"errors"
	"strconv"
)

type user_file struct {
	userId    int
	filename  string
	extension string
}

type file_version struct {
	user_file_id int
	file_id      int
	version_num  int
}

type file struct {
	uuid     string
	checksum string
}

func checkFileExistsForUser(userId int, filename string) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkFileExistsForUser", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT id FROM user_files WHERE userId = ? AND filename = ?", userId, filename)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	chk(err)
	loginfo("checkFileExistsForUser", "Comprobado si existe un archivo igual para dicho usuario", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var fileId = sqlResponse.String

		if fileId != "" {
			id, err := strconv.Atoi(fileId)
			if err != nil {
				return -1, err
			}
			return id, nil
		} else {
			return -1, nil
		}

	} else {
		return -1, errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return -1, errors.New("SQL Error: something has gone wrong")
}

func insertUserFile(data user_file) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("insertUserFile", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("INSERT INTO user_files (userId, filename, extension) VALUES (?, ?, ?)", data.userId, data.filename, data.extension)
	chk(err)
	loginfo("insertUserFile", "Insertando un archivo igual para un usuario", "db.QueryRow", "trace", nil)

	insertedId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	id := int(insertedId)
	if err != nil {
		return -1, err
	}
	return id, nil

	defer db.Close()
	return -1, errors.New("SQL Error: something has gone wrong")
}

func checkUserFileLastVersion(userFileId int) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkUserFileLastVersion", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT max(version_num) as lastVersion FROM file_versions WHERE user_file_id = ?", userFileId)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows || sqlResponse.String == "" {
		return -1, nil
	}
	chk(err)
	loginfo("checkUserFileLastVersion", "Comprobado y devuelto la última version de un archivo de un usuario", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var lastVersion = sqlResponse.String

		last, err := strconv.Atoi(lastVersion)
		if err != nil {
			return -1, err
		}

		if last > 0 {
			return last, nil
		} else {
			return 0, nil
		}

	} else {
		return -1, errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return -1, errors.New("SQL Error: something has gone wrong")
}

func insertFileVersion(data file_version) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("insertFileVersion", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("INSERT INTO file_versions (user_file_id, file_id, version_num) VALUES (?,?,?)", data.user_file_id, data.file_id, data.version_num)
	err = row.Scan(&sqlResponse)
	if err != sql.ErrNoRows {
		chk(err)
	}
	loginfo("insertFileVersion", "Insertando una nueva version de un archivo de un usuario", "db.QueryRow", "trace", nil)

	if err == sql.ErrNoRows || sqlResponse.String == "" {
		return -1, nil
	}

	if sqlResponse.Valid {

		var insertedId = sqlResponse.String

		id, err := strconv.Atoi(insertedId)
		if err != nil {
			return -1, err
		}

		return id, nil

	} else {
		return -1, errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return -1, errors.New("SQL Error: something has gone wrong")
}

func checkLastFileVersionHasUpdates(lastVersionFileId int, lastVersionNum int, newChecksum string) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkLastFileVersionHasUpdates", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var fileId sql.NullString
	err = db.QueryRow("SELECT file_id FROM file_versions WHERE user_file_id = ? AND version_num = ?", lastVersionFileId, lastVersionNum).Scan(&fileId)
	if err == sql.ErrNoRows {
		return false, nil
	}
	chk(err)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT checksum FROM files WHERE id = ?", fileId)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return false, nil
	}
	chk(err)
	loginfo("checkLastFileVersionHasUpdates", "Comprobado si la anterior version del archivo es igual al nuevo archivo subido", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var lastChecksum = sqlResponse.String

		if lastChecksum != newChecksum {
			return true, nil
		} else {
			return false, nil
		}

	} else {
		return false, errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return false, errors.New("SQL Error: something has gone wrong")
}

func checkFileExistsInDatabase(checksum string) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkFileExistsInDatabase", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT id FROM files WHERE checksum = ?", checksum)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	chk(err)
	loginfo("checkFileExistsInDatabase", "Comprobado si existe un archivo igual almacenado en la base de datos", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var fileId = sqlResponse.String

		if fileId != "" {
			id, err := strconv.Atoi(fileId)
			if err != nil {
				return -1, err
			}
			return id, nil
		} else {
			return -1, nil
		}

	} else {
		return -1, errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return -1, errors.New("SQL Error: something has gone wrong")
}

func insertFileInDatabase(data file) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("insertFileInDatabase", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("INSERT INTO files (uuid, checksum) VALUES (?,?)", data.uuid, data.checksum)
	chk(err)
	loginfo("insertFileInDatabase", "Insertando un nuevo archivo en la base de datos", "db.QueryRow", "trace", nil)

	insertedId, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	id := int(insertedId)
	if err != nil {
		return -1, err
	}
	return id, nil

	defer db.Close()
	return -1, errors.New("SQL Error: something has gone wrong")
}
