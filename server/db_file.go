package main

import (
	"database/sql"
	"errors"
	"strconv"
)

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

func getFilePackages(file_id int) ([]file_package, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getFilePackages", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	rows, err := db.Query("SELECT * FROM file_packages WHERE file_id = ?", file_id)
	chk(err)
	loginfo("getFilePackages", "Obteniendo los paquetes de un archivo de un usuario", "db.QueryRow", "trace", nil)
	defer rows.Close()

	var packages []file_package

	for rows.Next() {
		var filePackage file_package
		err := rows.Scan(&filePackage.id, &filePackage.file_id, &filePackage.package_id, &filePackage.package_index)
		chk(err)
		packages = append(packages, filePackage)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return packages, nil
}

func getPackage(package_id int) (packageFile, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getPackage", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var filePackage packageFile
	row := db.QueryRow("SELECT * FROM packages WHERE id = ?", package_id)
	err = row.Scan(&filePackage.id, &filePackage.uuid, &filePackage.checksum, &filePackage.timestamp, &filePackage.upload_user_id)
	if err == sql.ErrNoRows {
		return filePackage, nil
	}
	chk(err)
	loginfo("getPackage", "Obteniendo uuid de un paquete", "db.QueryRow", "trace", nil)

	return filePackage, nil

	defer db.Close()
	return filePackage, errors.New("SQL Error: something has gone wrong")
}

// FIXME: move to db_auth.go
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

func checkUserFileBelongsToUser(user_id int, user_file_id int) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("checkUserFileBelongsToUser", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	var sqlResponse sql.NullString
	row := db.QueryRow("SELECT userId FROM user_files WHERE id = ?", user_file_id)
	err = row.Scan(&sqlResponse)
	if err == sql.ErrNoRows {
		return false, nil
	}
	chk(err)
	loginfo("checkUserFileBelongsToUser", "Comprobado si el archivo pertence al usuario", "db.QueryRow", "trace", nil)

	if sqlResponse.Valid {

		var fileOwnerIdInString = sqlResponse.String
		fileOwnerId, err := strconv.Atoi(fileOwnerIdInString)
		chk(err)

		if fileOwnerId == user_id {
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

func getAllFileVersions(user_file_id int) ([]file, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getAllFileVersions", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	rows, err := db.Query("SELECT * FROM files WHERE user_file_id = ?", user_file_id)
	chk(err)
	loginfo("getAllFileVersions", "Obteniendo todas las versiones un archivo de un usuario", "db.QueryRow", "trace", nil)
	defer rows.Close()

	var fileVersions []file

	for rows.Next() {
		var version file
		err := rows.Scan(&version.id, &version.user_file_id, &version.packages_num, &version.checksum, &version.version, &version.size, &version.timestamp)
		chk(err)
		fileVersions = append(fileVersions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fileVersions, nil
}

func getOtherFilesUsingPackage(package_id int, file_id int) ([]file_package, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("getOtherFilesUsingPackage", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	rows, err := db.Query("SELECT * FROM file_packages WHERE package_id = ? AND file_id <> ?", package_id, file_id)
	chk(err)
	loginfo("getOtherFilesUsingPackage", "Obteniendo los paquetes de un archivo de un usuario", "db.QueryRow", "trace", nil)
	defer rows.Close()

	var packages []file_package

	for rows.Next() {
		var filePackage file_package
		err := rows.Scan(&filePackage.id, &filePackage.file_id, &filePackage.package_id, &filePackage.package_index)
		chk(err)
		packages = append(packages, filePackage)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return packages, nil
}

func deletePackageInDatabase(package_id int) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("deletePackage", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("DELETE FROM packages WHERE id = ?", package_id)
	chk(err)
	loginfo("deletePackage", "Borrando un paquete de BD", "db.QueryRow", "trace", nil)

	_, err = res.RowsAffected()
	chk(err)
	return true, nil

	defer db.Close()
	return false, errors.New("SQL Error: something has gone wrong")
}

func deleteFilePackages(file_id int) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("deleteFilePackages", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("DELETE FROM file_packages WHERE file_id = ?", file_id)
	chk(err)
	loginfo("deleteFilePackages", "Borrando la relación de un archivo con sus paquetes", "db.QueryRow", "trace", nil)

	_, err = res.RowsAffected()
	chk(err)
	return true, nil

	defer db.Close()
	return false, errors.New("SQL Error: something has gone wrong")
}

func deleteFileInDatabase(file_id int) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("deleteFile", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("DELETE FROM files WHERE id = ?", file_id)
	chk(err)
	loginfo("deleteFile", "Borrando un archivo de BD", "db.QueryRow", "trace", nil)

	_, err = res.RowsAffected()
	chk(err)
	return true, nil

	defer db.Close()
	return false, errors.New("SQL Error: something has gone wrong")
}

func deleteUserFile(user_file_id int) (bool, error) {
	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	chk(err)
	loginfo("deleteUserFile", "Conexión a MySQL abierta", "sql.Open", "trace", nil)

	res, err := db.Exec("DELETE FROM user_files WHERE id = ?", user_file_id)
	chk(err)
	loginfo("deleteUserFile", "Borrando la relación de un archivo con un usuario", "db.QueryRow", "trace", nil)

	_, err = res.RowsAffected()
	chk(err)
	return true, nil

	defer db.Close()
	return false, errors.New("SQL Error: something has gone wrong")
}
