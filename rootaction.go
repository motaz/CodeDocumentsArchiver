package main

import (
	"CodeDocumentsArchiver/archiverdata"
	"CodeDocumentsArchiver/controller"
	"CodeDocumentsArchiver/services"

	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"time"
)

type HeaderType struct {
	Username string
	Domain   string
	Title    string
	IsAdmin  bool
	Count    int
}

type SectionBox struct {
	SelectedID int
	Sections   []archiverdata.DocumentSection
}

type UploadFormType struct {
	Header HeaderType
	Key    string
	Today  string
	Box    SectionBox
	Text   string

	Message   string
	Class     string
	Page      int
	NextPage  int
	PrePage   int
	Documents []archiverdata.DocumentType
}

func checkSession(w http.ResponseWriter, req *http.Request) (success bool, userInfo archiverdata.UserType) {

	session := getCookie(w, req, "session")
	success = session != ""
	if success {
		domain := services.GetDomain(req)
		sessionResult := controller.CheckSession(domain, session, req.UserAgent()+"cda-98-")
		success = sessionResult.Success
		userInfo.UserName = sessionResult.UserName
		userInfo.UserID = sessionResult.UserID
		userInfo.IsAdmin = sessionResult.IsAdmin
	}
	if !success {
		RedirectToLogin(w, req)
	}
	return
}

func RedirectToLogin(w http.ResponseWriter, req *http.Request) {

	page := req.RequestURI
	http.Redirect(w, req, "/cda/login?page="+page, http.StatusTemporaryRedirect)

}

func viewUpload(w http.ResponseWriter, req *http.Request) {

	isValid, userInfo := checkSession(w, req)
	if isValid {
		domain := services.GetDomain(req)
		searchkey := req.FormValue("searchkey")
		var uform UploadFormType
		uform.Header.Title = getConfigurationParameter(domain, "title")
		uform.Header.Domain = domain
		if req.FormValue("upload") != "" {
			upload(w, req, &uform, userInfo.UserID)
		}
		uform.Page = 1

		uform.NextPage = uform.Page + 1
		if req.FormValue("page") != "" {
			apage := req.FormValue("page")
			uform.Page, _ = strconv.Atoi(apage)
		}
		uform.PrePage = uform.Page - 1
		uform.NextPage = uform.Page + 1

		uform.Key = searchkey
		today := time.Now()
		uform.Today = today.Format("2006-01-02")
		uform.Text = strings.Trim(req.FormValue("text"), "")
		uform.Box.Sections, _ = controller.GetDocumentTypes(domain)
		uform.Header.Count = controller.GetDocumentsCount(domain)
		uform.Header.Username = userInfo.UserName
		uform.Header.IsAdmin = userInfo.IsAdmin
		if req.FormValue("filter") != "" && req.FormValue("sectionid") != "-1" {
			uform.Box.SelectedID, _ = strconv.Atoi(req.FormValue("sectionid"))
			var err error
			uform.Documents, err = controller.FilterBySection(domain, uform.Box.SelectedID)
			uform.Page = 1
			if err != nil {
				uform.Class = "errormessage"
				uform.Message = err.Error()
			}
		} else if uform.Text == "" {
			var err error
			uform.Documents, err = controller.RetreiveLastDocuments(domain, uform.Page)
			if err != nil {
				uform.Class = "errormessage"
				uform.Message = err.Error()
			}
			if len(uform.Documents) == 0 {

				uform.NextPage = 0
			}
		} else {
			var err error
			uform.Documents, err = controller.SearchDocuments(domain, uform.Text)
			uform.PrePage = 0
			uform.NextPage = 0
			if err != nil {
				uform.Class = "errormessage"
				uform.Message = err.Error()
			}
		}

		for i := 0; i < len(uform.Documents); i++ {
			uform.Documents[i].ShowEdit = uform.Documents[i].UserID == userInfo.UserID || userInfo.IsAdmin
		}

		err := mytemplate.ExecuteTemplate(w, "index.html", uform)

		if err != nil {
			w.Write([]byte("Error: " + err.Error()))
		}
	}

}

type LoginData struct {
	Domain      string
	Username    string
	AuthDomain  string
	AuthDomains []string
}

func Login(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("content-type", "text/html")
	var data LoginData
	data.Domain = services.GetDomain(req)
	data.Username = req.FormValue("username")
	data.AuthDomain = req.FormValue("authdomain")

	if req.FormValue("submitlogin") != "" {
		remember := req.FormValue("remember") == "1"

		result := controller.Login(data.Domain, data.AuthDomain, data.Username,
			req.FormValue("password"), req.UserAgent()+"cda-98-", true, remember)
		if result.Success {
			setCookies(w, req, "session", result.SessionID, remember)
			if req.FormValue("page") == "" {
				fmt.Fprint(w, "<script>document.location='/cda/';</script>")
			} else {
				fmt.Fprintf(w, "<script>document.location='%s';</script>", req.FormValue("page"))
			}

		} else {
			fmt.Fprintf(w, "<p id=errormessage>Error: %s</p>", result.Message)
		}
	}

	thereIsUser, err := controller.ThereIsUser(data.Domain)
	auto := getConfigurationParameter(data.Domain, "autoaddusers")
	authDomains := getConfigurationParameter(data.Domain, "authdomains")
	data.AuthDomains = strings.Split(authDomains, ",")

	if auto == "yes" || (err == nil && thereIsUser) {
		err := mytemplate.ExecuteTemplate(w, "login.html", data)
		if err != nil {

			fmt.Fprint(w, err.Error())
		}
	} else {
		if req.FormValue("addadmin") != "" {
		} else {
			NewAdminUser(w, req)
		}
	}
}

type DomainType struct {
	Domain string
}

func NewAdminUser(w http.ResponseWriter, req *http.Request) {

	var domainData DomainType
	domainData.Domain = services.GetDomain(req)
	err := mytemplate.ExecuteTemplate(w, "newadmin.html", domainData)
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
}

func DownloadAttachment(w http.ResponseWriter, req *http.Request) {

	isValid, _ := checkSession(w, req)
	if isValid {
		revisionID := req.FormValue("id")
		domain := services.GetDomain(req)

		closer, doc, err := controller.GetAttachment(domain, revisionID)

		if err != nil {
			w.Write([]byte("Error: " + err.Error()))
		} else {
			w.Header().Set("Content-Disposition", "filename="+doc.FileName+";")
			_, err = io.Copy(w, closer)

			if err != nil {
				w.Write([]byte("Error copying file: " + err.Error()))
			}
		}

	}
}

type ViewDocumentType struct {
	Header        HeaderType
	ID            string
	DocumentDate  string
	FileName      string
	Description   string
	DocUsername   string
	InsertionTime time.Time
	UpdatedTime   time.Time
	Box           SectionBox
	SectionName   string
	Removed       bool
	ShowEdit      bool
	Message       string
	Class         string
	FileMD5       string
	History       []archiverdata.HistoryType
}

func doUpateInfo(userID int, w http.ResponseWriter, req *http.Request) (message, class string) {

	domain := services.GetDomain(req)
	doc, err := controller.GetDocumentByRevisionID(domain, req.FormValue("documentid"))
	if err == nil {
		doc.Description = req.FormValue("description")
		doc.DocumentDate, _ = time.Parse("2006-01-02", req.FormValue("documentdate"))

		doc.SectionID, _ = strconv.Atoi(req.FormValue("sectionid"))
		domain := services.GetDomain(req)
		_, err = controller.ModifyDocumentInfo(domain, doc, userID)

		if err != nil {
			class = "errormessage"
			message = err.Error()
		} else {
			class = "infomessage"
			message = "Document informationhas updated"
		}
	}
	return
}

func ViewDocument(w http.ResponseWriter, req *http.Request) {

	isValid, userInfo := checkSession(w, req)
	if isValid {
		revisionID := req.FormValue("documentid")
		if revisionID == "" {
			revisionID = req.FormValue("id")
		}
		var docForm ViewDocumentType
		domain := services.GetDomain(req)
		docForm.Header.Domain = domain
		docForm.Header.Username = userInfo.UserName
		docForm.Header.Title = getConfigurationParameter(domain, "title")

		doc, err := controller.GetDocumentByRevisionID(domain, revisionID)
		if err != nil {
			fmt.Fprintf(w, "Error %v", err.Error())
		} else {
			docForm.ShowEdit = userInfo.UserID == doc.UserID || userInfo.IsAdmin

			if docForm.ShowEdit {
				if req.FormValue("remove") != "" {

					docForm.Message, docForm.Class = doTempRemoveDocument(userInfo.UserID, w, req)
				}
			}
			if req.FormValue("updateinfo") != "" {

				docForm.Message, docForm.Class = doUpateInfo(userInfo.UserID, w, req)

			}

			if req.FormValue("uploaddodument") != "" {

				revisionID, docForm.Message, docForm.Class = reUpload(w, req, userInfo.UserID)
			}

			if req.FormValue("restore") != "" {
				docForm.Message, docForm.Class = doRestoreDocument(userInfo.UserID, w, req)
			}
			docForm.History = controller.GetHistory(domain, doc.ID)

			doc, err = controller.GetDocumentByRevisionID(domain, revisionID)

			if err == nil {

				docForm.DocumentDate = doc.DocumentDate.Format("2006-01-02")
				docForm.Description = doc.Description
				docForm.FileName = doc.FileName
				docForm.ID = doc.RevisionID
				docForm.DocUsername = controller.GetUserByID(domain, doc.UserID).UserName
				docForm.InsertionTime = doc.InsertionTime
				docForm.UpdatedTime = doc.UpdatedTime

				if len(docForm.History) > 0 {
					docForm.FileMD5 = docForm.History[0].FileMD5
				}

				docForm.Box.Sections, err = controller.GetDocumentTypes(domain)
				docForm.Box.SelectedID = doc.SectionID
				docForm.SectionName = controller.GetSectionName(domain, doc.SectionID)
				docForm.Removed = doc.Removed

				err := mytemplate.ExecuteTemplate(w, "viewdocument.html", docForm)

				if err != nil {
					w.Write([]byte("Error: " + err.Error()))
				}
			}

		}
	}

}

func doTempRemoveDocument(userID int, w http.ResponseWriter, req *http.Request) (message, class string) {

	domain := services.GetDomain(req)
	doc, err := controller.GetDocumentByRevisionID(domain, req.FormValue("documentid"))
	if err == nil {

		_, err = controller.TempRemoveDocument(domain, doc, userID)
		if err != nil {
			class = "errormessage"
			message = err.Error()
		} else {
			class = "infomessage"
			message = "Document has been removed temporarly"
		}
	} else {
		class = "errormessage"
		message = err.Error()
	}
	return
}

func doRestoreDocument(userID int, w http.ResponseWriter, req *http.Request) (message, class string) {

	domain := services.GetDomain(req)
	doc, err := controller.GetDocumentByRevisionID(domain, req.FormValue("documentid"))
	if err == nil {

		_, err = controller.RestoreDocument(domain, doc, userID)
		if err != nil {
			class = "errormessage"
			message = err.Error()
		} else {
			class = "infomessage"
			message = "Document has been restored"
		}

	} else {
		class = "errormessage"
		message = err.Error()
	}

	return
}

func Logout(w http.ResponseWriter, req *http.Request) {

	expiration := time.Now()
	cookie := http.Cookie{Name: "session", Value: "-", Expires: expiration}
	http.SetCookie(w, &cookie)

	http.Redirect(w, req, "/cda", 307)
}

func History(w http.ResponseWriter, req *http.Request) {

	domain := services.GetDomain(req)

	w.Header().Set("content-type", "text/html")
	count, err := controller.ProcessMissingHistory(domain)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		fmt.Fprintf(w, "%d inserted", count)
	}
}
