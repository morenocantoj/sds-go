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

type file struct {
	user_file_id int
	packages_num int
	checksum     string
	version      int
	size         int
}

type file_package struct {
	file_id       int
	package_id    int
	package_index int
}

type packageFile struct {
	uuid           string
	checksum       string
	upload_user_id int
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
	row := db.QueryRow("SELECT max(version) as lastVersion FROM files WHERE user_file_id = ?", userFileId)
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

func insertFile(data file) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("insertFile", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("INSERT INTO files (user_file_id, packages_num, checksum, version, size) VALUES (?,?,?,?,?)", data.user_file_id, data.packages_num, data.checksum, data.version, data.size)
	chk(err)
	loginfo("insertFile", "Insertando un archivo de un usuario", "db.QueryRow", "trace", nil)

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

func checkLastFileVersionHasUpdates(userFileId int, lastVersionNum int, newChecksum string) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkLastFileVersionHasUpdates", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	err = db.QueryRow("SELECT checksum FROM files WHERE user_file_id = ? AND version = ?", userFileId, lastVersionNum).Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return true, nil
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

func checkPackageExistsInDatabase(checksum string) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkPackageExistsInDatabase", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT id FROM packages WHERE checksum = ?", checksum)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	chk(err)
	loginfo("checkPackageExistsInDatabase", "Comprobado si existe un archivo igual almacenado en la base de datos", "db.QueryRow", "trace", nil)

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

func insertPackageInDatabase(data packageFile) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("insertPackageInDatabase", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("INSERT INTO packages (uuid, checksum, upload_user_id) VALUES (?,?,?)", data.uuid, data.checksum, data.upload_user_id)
	chk(err)
	loginfo("insertPackageInDatabase", "Insertando un nuevo paquete en la base de datos", "db.QueryRow", "trace", nil)

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

func insertFilePackage(data file_package) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("insertFilePackage", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("INSERT INTO file_packages (file_id, package_id, package_index) VALUES (?,?,?)", data.file_id, data.package_id, data.package_index)
	chk(err)
	loginfo("insertFilePackage", "Relacionando un paquete con un archivo en la base de datos", "db.QueryRow", "trace", nil)

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
