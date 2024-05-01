// CodeDocumentsArchiver project main.go
package main

import (
	"CodeDocumentsArchiver/controller"
	"CodeDocumentsArchiver/services"
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

var mytemplate *template.Template

const VERSION = "1.2.11 r1-May"

//go:embed templates
var templatesFS embed.FS

//go:embed resources
var static embed.FS

func InitTemplate(embededTemplates embed.FS) (atemplate *template.Template, err error) {

	atemplate, err = template.ParseFS(embededTemplates, "templates/*.html")
	if err != nil {
		fmt.Println("error in InitTemplate: " + err.Error())
	}
	return
}

func main() {

	writeLog("Starting CDA: version: " + VERSION)
	var err error
	mytemplate, err = InitTemplate(templatesFS)
	if err == nil {
		err = controller.InitDB()
		if err != nil {
			writeLog("Error in initialization : " + err.Error())
		} else {

			http.HandleFunc("/", redirectToIndex)
			http.HandleFunc("/cda", viewUpload)
			http.HandleFunc("/cda/", viewUpload)
			http.HandleFunc("/cda/login", Login)
			http.HandleFunc("/cda/logout", Logout)
			http.HandleFunc("/cda/download", DownloadAttachment)
			http.HandleFunc("/cda/document", ViewDocument)
			http.HandleFunc("/cda/UploadAttachment", services.UploadAttachment)

			http.Handle("/cda/resources/", http.StripPrefix("/cda/", http.FileServer(http.FS(static))))

			fmt.Println("Code Documents Archiver, Listening on port 10032")
			fmt.Println("http://localhost:10032")

			err = http.ListenAndServe(":10032", nil)
			if err != nil {
				message := "Error while listening: " + err.Error()
				fmt.Println(message)
				writeLog(message)
			}
		}
	} else {
		writeLog("Error loading templates: " + err.Error())
	}
}

func redirectToIndex(w http.ResponseWriter, req *http.Request) {

	http.Redirect(w, req, "/cda"+req.RequestURI, http.StatusTemporaryRedirect)
}
