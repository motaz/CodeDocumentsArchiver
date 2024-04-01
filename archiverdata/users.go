package archiverdata

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type UserType struct {
	UserID   int
	UserName string
	IsAdmin  bool
	Emails   []string
}

func GetUsers(databasename string) (users []UserType, err error) {

	stmt := `SELECT id, username, isAdmin FROM ` + databasename + `.users`

	rows, err := dbconn.Query(stmt)
	if err != nil {
		writeLog("Error in GetUsers: " + err.Error())

	} else {
		defer rows.Close()

		for rows.Next() {
			var user UserType

			err = rows.Scan(&user.UserID, &user.UserName, &user.IsAdmin)

			if err == nil {
				users = append(users, user)
			} else {
				writeLog("Error in GetUsers scan: " + err.Error())
			}
		}
	}

	return
}

func GetUserEmails(databasename string, userID int) (emails []string, err error) {

	stmt := `select email from ` + databasename + `.emails where userID=?`

	rows, err := dbconn.Query(stmt, userID)
	if err != nil {
		writeLog("Error in GetUserEmails: " + err.Error())
	} else {
		defer rows.Close()

		for rows.Next() {
			var email string
			err = rows.Scan(&email)
			if err == nil {
				emails = append(emails, email)
			} else {
				writeLog(err.Error())
			}
		}
	}

	return
}

func GetUserByEmail(databasename string, email string) (user UserType, err error) {

	stmt := `select userID from ` + databasename + `.emails where lower(email)=?`

	row := dbconn.QueryRow(stmt, strings.ToLower(email))
	err = row.Err()
	if err == nil {

		var userID int
		err = row.Scan(&userID)
		if err == nil {
			user, err = GetUserByid(databasename, userID)
		}
	}

	return
}

func readUserInfo(databasename string, row *sql.Row) (user UserType, err error) {

	err = row.Scan(&user.UserID, &user.UserName, &user.IsAdmin)
	if err == nil {
		user.Emails, err = GetUserEmails(databasename, user.UserID)
	} else {
		writeLog("readUserInfo: " + err.Error())
	}
	return
}

func GetUserByid(databasename string, id int) (user UserType, err error) {

	stmt := `SELECT id, username, isAdmin FROM ` + databasename + `.users
			 WHERE id = ?`

	row := dbconn.QueryRow(stmt, id)
	user, err = readUserInfo(databasename, row)

	return
}

func UpdateUserInfo(databasename string, userID int, userName string, isAdmin bool) (success bool, err error) {

	st := `UPDATE ` + databasename + `.users SET username=?, isAdmin=? WHERE id=?`

	_, err = dbconn.Exec(st, userName, isAdmin, userID)
	success = err == nil
	if !success {
		writeLog("Error in UpdateUserInfo: " + err.Error())
	}

	return
}

func InsertNewUser(databasename string, userInfo UserType) (success bool, userID int, err error) {

	st := `insert into ` + databasename + `.users (ID, Username, IsAdmin) values (?, ?, ?)`

	result, err := dbconn.Exec(st, userInfo.UserID, userInfo.UserName, userInfo.IsAdmin)
	success = err == nil
	if success {
		lastId, _ := result.LastInsertId()
		userID = int(lastId)

		if len(userInfo.Emails) > 0 {
			for _, email := range userInfo.Emails {
				InsertUserEmail(databasename, userInfo.UserID, email)
			}
		}
	} else {
		writeLog("Error in InsertNewUser: " + err.Error())
	}

	return
}

func InsertUserEmail(databasename string, userID int, email string) (success bool, err error) {

	st := `insert into ` + databasename + `.emails (UserID, Email) values (?, ?)`

	_, err = dbconn.Exec(st, userID, strings.ToLower(email))
	success = err == nil
	if !success {
		writeLog("Error in InsertUserEmail: " + err.Error())
	}

	return
}

func GetUserInfoByName(databasename string, username string) (found bool, userInfo UserType, err error) {

	st := `SELECT id, username, isAdmin FROM ` + databasename + `.users WHERE username=?`

	row := dbconn.QueryRow(st, username)

	userInfo, err = readUserInfo(databasename, row)
	found = err == nil

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			err = errors.New("User: " + username + " does not exist in Documents system")
		}
	}

	return
}
