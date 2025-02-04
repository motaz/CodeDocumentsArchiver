package controller

import (
	"CodeDocumentsArchiver/archiverdata"
	"bufio"
	"bytes"

	"strings"
	"time"

	"io"
)

func findUser(userID int, users []archiverdata.UserType) (username string) {

	for _, user := range users {
		if userID == user.UserID {
			username = user.UserName
			break
		}
	}
	return
}

func findSection(sectionID int, sections []archiverdata.DocumentSection) (sectionname string) {

	for _, section := range sections {
		if sectionID == section.SectionID {
			sectionname = section.SectionName
			break
		}
	}
	return
}

func fillLookups(domain string, documents []archiverdata.DocumentType) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	users, err := archiverdata.GetUsers(databasename)

	sections, err2 := archiverdata.GetSections(databasename)

	if err == nil && err2 == nil {
		for i := 0; i < len(documents); i++ {
			documents[i].UserName = findUser(documents[i].UserID, users)
			documents[i].SectionName = findSection(documents[i].SectionID, sections)
			documents[i].IsNew = documents[i].InsertionTime.After(time.Now().AddDate(0, 0, -3))

			documents[i].IsNewUpdated = documents[i].UpdatedTime.After(time.Now().AddDate(0, 0, -3))

		}

	}
}

func RetreiveLastDocuments(domain string, page int) (documents []archiverdata.DocumentType, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)
	lastDocuments, err := archiverdata.GetLastDocuments(databasename, page)
	if err == nil {
		for _, item := range lastDocuments {
			if !item.Removed {
				documents = append(documents, item)
			}
		}
		fillLookups(domain, documents)
	}
	return
}

func InsertNewAttachment(domain string, theDoc archiverdata.DocumentType,
	filename string, file *bufio.Reader, autoInsertionTime bool, info archiverdata.DocumentInfoType) (docID int64, err error) {

	buf := bytes.NewBuffer(nil)

	var copied int64
	copied, err = io.Copy(buf, file)

	if err == nil && copied > 0 {
		databasename := archiverdata.GetDatabaseNameFromDomain(domain)
		theDoc.Year = GetCurrentYear()
		if autoInsertionTime {
			docID, err = archiverdata.InsertDocumentInfo(databasename, &theDoc, info)
		} else {
			docID, err = archiverdata.ImportDocumentInfo(databasename, &theDoc, info)

		}
		if err == nil {
			_, err = archiverdata.InsertAttachement(databasename, theDoc.Year, docID, theDoc.RevisionID, filename, buf)
			if err == nil {
				archiverdata.InsertHistory(databasename, theDoc.Year, theDoc.RevisionID, "new", docID, theDoc.UserID)
			}
		}
	}

	return
}

func GetAttachment(domain string, revisionID string) (closer io.ReadCloser, doc archiverdata.DocumentType, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)
	doc, err = archiverdata.GetDocumentByRevisionID(databasename, revisionID)
	if err == nil {
		closer, _, err = archiverdata.GetAttachment(databasename, doc.Year, doc.RevisionID)
	} else {
		if strings.Contains(err.Error(), "not found") {
			var record archiverdata.HistoryType
			record, err = archiverdata.GetHistoryRecord(databasename, revisionID)
			if err == nil {
				closer, _, err = archiverdata.GetAttachment(databasename, record.Year, record.RevisionID)
			}
		}

	}
	return
}

func GetDocumentByRevisionID(domain, revisionID string) (doc archiverdata.DocumentType, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	doc, err = archiverdata.GetDocumentByRevisionID(databasename, revisionID)

	return
}

func ModifyDocumentInfo(domain string, doc archiverdata.DocumentType, userID int) (success bool, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	success, err = archiverdata.UpdateDocumentInfo(databasename, doc)
	if success {
		archiverdata.InsertHistory(databasename, doc.Year, doc.RevisionID, "updateinfo", doc.ID, userID)
	}

	return
}

func ModifyAttachment(domain string, theDoc *archiverdata.DocumentType,
	buf *bytes.Buffer, newRevision bool, year string) (success bool, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	success = false
	var event string
	if !newRevision {
		fileMD5 := GetFileMD5(buf)

		success, err = archiverdata.UpdateAttachment(databasename, theDoc.ID, theDoc.RevisionID,
			theDoc.Year, theDoc.FileName, fileMD5, buf)
		event = "updatefile"
	} else {

		theDoc.RevisionID = archiverdata.GetNewRevisionID(databasename)
		success, err = archiverdata.InsertAttachement(databasename, year, theDoc.ID,
			theDoc.RevisionID, theDoc.FileName, buf)
		if success {
			theDoc.Year = year

			archiverdata.UpdateDocumentInfo(databasename, *theDoc)
			event = "revision"
		}
	}
	if success {
		archiverdata.InsertHistory(databasename, year, theDoc.RevisionID, event,
			theDoc.ID, theDoc.UserID)
	}
	return
}

func TempRemoveDocument(domain string, doc archiverdata.DocumentType, userID int) (success bool, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	success, err = archiverdata.TempRemoveDocument(databasename, doc)
	if success {
		archiverdata.InsertHistory(databasename, doc.Year, doc.RevisionID, "remove", doc.ID, userID)
	}

	return
}

func RestoreDocument(domain string, doc archiverdata.DocumentType, userID int) (success bool, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	success, err = archiverdata.RestoreDocument(databasename, doc)
	if success {
		archiverdata.InsertHistory(databasename, doc.Year, doc.RevisionID, "restore", doc.ID, userID)
	}
	return
}

func SearchDocuments(domain, text string) (documents []archiverdata.DocumentType, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	documents, err = archiverdata.SearchDocuments(databasename, text)

	fillLookups(domain, documents)

	return
}

func GetDocumentsCount(domain string) (count int) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	count, _ = archiverdata.GetDocumentsCount(databasename)

	return
}

func FilterBySection(domain string, sectionID int) (documents []archiverdata.DocumentType, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	documents, err = archiverdata.GetFilteredDocuments(databasename, sectionID)
	fillLookups(domain, documents)

	return
}

func GetHistory(domain string, docID int64) (history []archiverdata.HistoryType) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	history, _ = archiverdata.GetHistory(databasename, docID)

	return
}

func ProcessMissingHistory(domain string) (count int, err error) {

	count = 0
	databasename := archiverdata.GetDatabaseNameFromDomain(domain)
	allDocuments, err := archiverdata.GetAllDocuments(databasename)
	if err == nil {
		for _, item := range allDocuments {
			history, err := archiverdata.GetHistory(databasename, item.ID)
			if err == nil {
				if len(history) == 0 {
					success, _ := archiverdata.InsertHistoryWithTime(databasename, item.Year, item.RevisionID, "new",
						item.ID, item.UserID, item.InsertionTime)
					if success {
						count++
					}
				}
			}
		}
	}
	return
}
