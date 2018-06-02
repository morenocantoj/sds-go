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

func getFileId(user_file_id int, version int) (int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getFileId", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT id FROM files WHERE user_file_id = ? AND version = ?", user_file_id, version)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	chk(err)
	loginfo("getFileId", "Obteniendo id de una version de un archivo de un usuario", "db.QueryRow", "trace", nil)

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

func getFileChecksum(user_file_id int, version int) (string, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getFileChecksum", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT checksum FROM files WHERE user_file_id = ? AND version = ?", user_file_id, version)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return "", nil
	}
	chk(err)
	loginfo("getFileChecksum", "Obteniendo id de una version de un archivo de un usuario", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var fileChecksum = sqlResponse.String

		return fileChecksum, nil

	} else {
		return "", errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return "", errors.New("SQL Error: something has gone wrong")
}

func getUserFileName(user_file_id int) (string, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getUserFileName", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT filename FROM user_files WHERE id = ?", user_file_id)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return "", nil
	}
	chk(err)
	loginfo("getUserFileName", "Obteniendo el nombre de un archivo", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var fileName = sqlResponse.String

		return fileName, nil

	} else {
		return "", errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return "", errors.New("SQL Error: something has gone wrong")
}

func getFilePackages(file_id int) (map[int]int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getFilePackages", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	rows, err := db.Query("SELECT package_id, package_index FROM file_packages WHERE file_id = ?", file_id)
	chk(err)
	loginfo("getFilePackages", "Obteniendo los paquetes de un archivo de un usuario", "db.QueryRow", "trace", nil)
	defer rows.Close()

	packages := make(map[int]int)

	for rows.Next() {
		var (
			package_id    int
			package_index int
		)
		if err := rows.Scan(&package_id, &package_index); err != nil {
			return nil, err
		}
		packages[package_index] = package_id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return packages, nil
}

func getPackageUuid(package_id int) (string, int, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getPackageUuid", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var (
		uuid           string
		upload_user_id int
	)
	row := db.QueryRow("SELECT uuid, upload_user_id FROM packages WHERE id = ?", package_id)
	err = row.Scan(&uuid, &upload_user_id)
	if err == sql.ErrNoRows {
		return "", -1, nil
	}
	chk(err)
	loginfo("getPackageUuid", "Obteniendo uuid de un paquete", "db.QueryRow", "trace", nil)

	return uuid, upload_user_id, nil

	defer db.Close()
	return "", -1, errors.New("SQL Error: something has gone wrong")
}

func getUserSecretKeyById(userId int) (string, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getUserSecretKeyById", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT secret_key FROM users WHERE id = ?", userId)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return "", nil
	}
	chk(err)
	loginfo("getUserSecretKeyById", "Obteniendo la clave secreta del usuario para cifrar archivos", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var secretKey = sqlResponse.String
		return secretKey, nil

	} else {
		return "", errors.New("SQL Response: response is not valid")
	}

	defer db.Close()
	return "", errors.New("SQL Error: something has gone wrong")
}
