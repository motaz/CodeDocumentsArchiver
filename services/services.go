package services

import (
	"CodeDocumentsArchiver/archiverdata"
	"CodeDocumentsArchiver/controller"
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"io/ioutil"
	"net/http"
	"strings"
)

type UploadInputType struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Title    string `json:"title"`
	FileName string `json:"filename"`
	Contents string `json:"contents"`
	Domain   string `json:"domain"`
}

type ResultMessage struct {
	Status  int    `json:"-"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func setError(result *ResultMessage, status int, message string) {
	result.Status = status
	result.Message = message
	result.Success = false
}

func GetDomain(req *http.Request) (domain string) {

	domain = req.Host
	if strings.Contains(domain, "www.") {
		domain = domain[strings.Index(domain, "www.")+4:]
	}
	if strings.Contains(domain, ".") {

		domain = domain[:strings.Index(domain, ".")]
	}
	if strings.Contains(domain, "[::]") || strings.Contains(domain, "[::1]") ||
		strings.Contains(domain, "[::0]") {
		domain = "localhost"

	}
	if strings.Contains(domain, ":") {
		domain = domain[:strings.Index(domain, ":")]
	}
	return
}

func UploadAttachment(w http.ResponseWriter, req *http.Request) {

	body, _ := ioutil.ReadAll(req.Body)

	var uploadinput UploadInputType
	json.Unmarshal(body, &uploadinput)
	domain := uploadinput.Domain
	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	authDomain := ""
	if strings.Contains(uploadinput.Username, "/") {
		authDomain = uploadinput.Username[:strings.Index(uploadinput.Username, "/")-1]
	}

	var resultMessage ResultMessage
	if uploadinput.Username != "" && uploadinput.Password != "" && uploadinput.Email != "" {
		result := controller.Login(domain, authDomain, uploadinput.Username, uploadinput.Password, "", false, false)
		resultMessage.Success = result.Success
		resultMessage.Message = result.Message
		if result.Success {

			user, err := archiverdata.GetUserByEmail(databasename, uploadinput.Email)
			if err == nil {

				resultMessage.Status = http.StatusOK
				resultMessage.Success = true
				sDec, _ := base64.StdEncoding.DecodeString(uploadinput.Contents)
				bR := bytes.NewReader(sDec)
				r := bufio.NewReader(bR)
				var doc archiverdata.DocumentType
				var info archiverdata.DocumentInfoType

				doc.Description = uploadinput.Title
				doc.UserID = user.UserID
				doc.FileName = uploadinput.FileName
				doc.InsertionTime = time.Now()
				doc.DocumentDate = time.Now()
				if len(uploadinput.Contents) < 100 {
					resultMessage.Message = "Empty attachment"
				} else {
					info.IsPublic = false

					_, err := controller.InsertNewAttachment(domain, doc, uploadinput.FileName,
						r, true, info)
					if err != nil {
						setError(&resultMessage, http.StatusInternalServerError, err.Error())
					} else {
						resultMessage.Message = "Uploaded: " + uploadinput.FileName
					}
				}

			} else {
				if strings.Contains(err.Error(), "does not") {
					setError(&resultMessage, http.StatusNotFound, err.Error())
				} else {
					setError(&resultMessage, http.StatusInternalServerError, err.Error())
				}

			}

		} else {
			setError(&resultMessage, http.StatusUnauthorized, result.Message)

		}

	} else {
		setError(&resultMessage, http.StatusBadRequest, "Empty data")
	}
	data, _ := json.Marshal(resultMessage)
	w.WriteHeader(resultMessage.Status)

	w.Write(data)
	fmt.Printf("%+v", resultMessage)
}
