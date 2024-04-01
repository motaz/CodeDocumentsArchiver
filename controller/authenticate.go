package controller

import (
	"CodeDocumentsArchiver/archiverdata"
	"bytes"
	"encoding/json"
	"errors"

	"io/ioutil"
	"net/http"
	"time"

	"github.com/motaz/codeutils"
)

type LoginRequest struct {
	User     string `json:"username"`
	Password string `json:"password"`
	Domain   string `json:"domain"`
	Key      string `json:"key"`
}

type SessionRequest struct {
	SessionID string `json:"sessionid"`
	Key       string `json:"key"`
}

func callCodeA(url string, request []byte) (result []byte, err error) {

	timeout := time.Duration(30 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	response, err := client.Post(url, "application/json", bytes.NewBuffer(request))

	if err == nil {

		result, _ = ioutil.ReadAll(response.Body)
	}

	return
}

type LoginResult struct {
	Success   bool   `json:"success"`
	Domain    string `json:"domain"`
	SessionID string `json:"sessionid"`
	Message   string `json:"message"`
	ErrorCode int    `json:"errorcode"`
	UserID    int
}

func getAuthenticationURL(domain string) (url string) {

	url = codeutils.GetConfigValue(domain+".ini", "authurl")
	return
}

func Login(domain, authdomain, user, password, key string, storeSession, longSession bool) (result LoginResult) {

	var userInfo archiverdata.UserType

	url := getAuthenticationURL(domain)
	var loginrequest LoginRequest
	if authdomain != "default" {
		loginrequest.Domain = authdomain
	}
	loginrequest.User = user
	loginrequest.Password = password
	loginrequest.Key = key

	jsonValue, _ := json.Marshal(loginrequest)
	println(string(jsonValue))

	resultStr, err := callCodeA(url+"CheckLogin", jsonValue)
	println(string(resultStr))
	if err == nil {
		json.Unmarshal(resultStr, &result)
		if result.Success {
			if loginrequest.Domain != "" {
				user = loginrequest.Domain + "/" + user
			}
			result.Success, userInfo, result.Message = checkLocalUser(domain, user)
			if !result.Success {

				auto := codeutils.GetConfigValue(domain+".ini", "autoaddusers")
				if auto == "yes" {
					userInfo.UserID, err = InsertNewUser(domain, user)
					userInfo.UserName = loginrequest.User
					userInfo.IsAdmin = false

					result.Success = err == nil
				} else {
					result.Success = false
					result.ErrorCode = 1
					result.Message = "User " + user + " does not exist in Documents system"
				}
			}
		} else {
			result.Success = false
			result.ErrorCode = 2

		}
		if result.Success {
			var session archiverdata.SessionInfoType
			session.SessionID = result.SessionID
			session.UserID = userInfo.UserID
			session.UserName = user
			session.IsAdmin = userInfo.IsAdmin
			session.KeepLogin = longSession
			if storeSession {
				err = InsertSession(domain, session, key)

			}
		}

	}

	return
}

type SessionResult struct {
	LoginResult
	UserName    string `json:"username"`
	SessionInfo string `json:"sessioninfo"`
	SessionID   string `json:"-"`
	IsAdmin     bool   `json:"isadmin"`
	UserID      int
}

func CheckSession(domain, sessionID, key string) (result SessionResult) {

	session, err := GetSession(domain, sessionID, key)
	if err == nil {
		result.Success = true
		result.UserID = session.UserID
		result.UserName = session.UserName
		result.IsAdmin = session.IsAdmin
		result.SessionInfo = session.SessionInfo

	} else {
		result.Success = false
		result.ErrorCode = 5
		result.Message = err.Error()
	}
	return
}

func checkLocalUser(domain, username string) (isValid bool, userInfo archiverdata.UserType, message string) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)
	var err error
	isValid, userInfo, err = archiverdata.GetUserInfoByName(databasename, username)

	if err != nil {
		message = err.Error()
	}
	return
}

func InsertNewUser(domain, username string) (userID int, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	var user archiverdata.UserType
	user.UserName = username
	isExist, _, _ := archiverdata.GetUserInfoByName(databasename, username)
	if !isExist {
		_, userID, err = archiverdata.InsertNewUser(databasename, user)

	} else {
		err = errors.New("User " + username + " is already exists")
	}

	return
}

func InsertSession(domain string, session archiverdata.SessionInfoType, key string) (err error) {

	session.SessionID = GetMD5Hash(session.SessionID + key)
	err = archiverdata.InsertSession(session)

	return
}

func GetSession(domain string, sessionID, key string) (session archiverdata.SessionInfoType, err error) {

	session, err = archiverdata.GetSession(GetMD5Hash(sessionID + key))

	return
}

func GetUserByEmail(domain, email string) (user archiverdata.UserType, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	user, err = archiverdata.GetUserByEmail(databasename, email)
	return
}
