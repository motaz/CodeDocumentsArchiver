package main

import (
	"CodeDocumentsArchiver/archiverdata"
	"CodeDocumentsArchiver/controller"
	"CodeDocumentsArchiver/services"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"

	"net/http"
	"strconv"
	"time"
)

// upload logic
func upload(w http.ResponseWriter, r *http.Request, form *UploadFormType, userID int) {

	w.Header().Add("Content-Type", "text/html;charset=UTF-8")
	w.Header().Add("encoding", "UTF-8")
	if r.Method == "GET" {
		//crutime := time.Now().Unix()
		//h := md5.New()
		//o.WriteString(h, strconv.FormatInt(crutime, 10))

		//	t, _ := template.ParseFiles("upload.gtpl")
		//t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")

		if err != nil {
			w.Write([]byte("<h3>Error: " + err.Error() + "</h3>"))
			return
		}
		defer file.Close()
		filer := bufio.NewReader(file)

		var fileInfo archiverdata.DocumentType
		fileInfo.Description = r.FormValue("description")
		fileInfo.FileName = archiverdata.GetOnlyFile(handler.Filename)
		fileInfo.InsertionTime = time.Now()
		fileInfo.DocumentDate, err = time.Parse("2006-01-02", r.FormValue("documentdate"))
		if err != nil {
			form.Class = "errormessage"
			form.Message = err.Error()
		} else {
			fileInfo.SectionID, _ = strconv.Atoi(r.FormValue("sectionid"))
			fileInfo.Type = "document"
			fileInfo.UserID = userID
			domain := services.GetDomain(r)
			_, err = controller.InsertNewAttachment(domain, fileInfo, fileInfo.FileName, filer, true)
			if err == nil {
				form.Class = "infomessage"
				form.Message = "File: " + fileInfo.FileName + " Has uploaded"

			} else {
				form.Class = "errormessage"
				form.Message = err.Error()
			}
		}

		writeLog("Received : " + handler.Filename + ", from " + r.RemoteAddr)

	}

}

func reUpload(w http.ResponseWriter, r *http.Request, userID int) (revisionID, message, class string) {

	isValid, userInfo := checkSession(w, r)
	if isValid {
		domain := services.GetDomain(r)
		w.Header().Add("Content-Type", "text/html;charset=UTF-8")
		w.Header().Add("encoding", "UTF-8")
		if r.Method == "GET" {
			//crutime := time.Now().Unix()
			//h := md5.New()
			//o.WriteString(h, strconv.FormatInt(crutime, 10))

			//	t, _ := template.ParseFiles("upload.gtpl")
			//t.Execute(w, token)
		} else {
			r.ParseMultipartForm(32 << 20)
			file, handler, err := r.FormFile("uploadfile")

			if err != nil {
				class = "errormessage"
				message = err.Error()
				return
			}
			defer file.Close()
			filer := bufio.NewReader(file)
			buf := bytes.NewBuffer(nil)
			size, err := io.Copy(buf, filer)

			var fileInfo archiverdata.AttachmentRecordInputType
			id := r.FormValue("documentid")
			newRevision := r.FormValue("newrevision") == "1"
			var oldDoc archiverdata.DocumentType

			doc, err := controller.GetDocumentByRevisionID(domain, id)
			oldDoc = doc
			doc.FileName = archiverdata.GetOnlyFile(handler.Filename)
			doc.InsertionTime = time.Now()
			doc.UserID = userID
			if err != nil {
				class = "errormessage"
				message = err.Error()
				return
			}
			fileInfo.UserID = userInfo.UserID

			year := controller.GetCurrentYear()

			if year != doc.Year || oldDoc.FileName != doc.FileName {
				newRevision = true
			}
			if !newRevision {
				_, doc, err := controller.GetAttachment(domain, id)
				if err == nil && doc.AttachmentSize > 0 {
					dif := (size - doc.AttachmentSize) * 100 / doc.AttachmentSize
					delta := math.Abs(float64(dif))

					newRevision = delta > 20
					fmt.Println(doc.AttachmentSize, size, delta, newRevision)

				}
			}

			_, err = controller.ModifyAttachment(domain, &doc, buf, newRevision, year)
			revisionID = doc.RevisionID
			if err == nil {
				class = "infomessage"
				message = "Attachmetn File has updated"
				if newRevision {
					message += " into new revision"
				}

			} else {
				class = "errormessage"
				message = err.Error()
			}

			writeLog("Received : " + handler.Filename + ", from " + r.RemoteAddr)

		}
	} else {
		RedirectToLogin(w, r)

	}
	return
}
