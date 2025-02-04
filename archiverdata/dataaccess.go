package archiverdata

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/motaz/codeutils"

	_ "github.com/go-sql-driver/mysql"
)

const SELECT_FROM_DOCUMENTS = `SELECT id, revisionID, filename, insertionTime, 
		documentDate, updatedTime, userID, sectionID, year, description, isRemoved , Info FROM  `

var dbconn *sql.DB

func OpenConnection() (err error) {

	databaseServer := GetCDAConfigValue("dbhost", "localhost")
	databaseUser := GetCDAConfigValue("dbuser", "")

	password := GetCDAConfigValue("dbpass", "")
	connectionString := fmt.Sprintf("%v:%v@tcp(%s:3306)/",
		databaseUser, password, databaseServer)

	dbconn, err = sql.Open("mysql", connectionString)
	if err != nil {

		writeLog("Error in SQLConnection" + err.Error())

	} else {
		dbconn.SetMaxOpenConns(10)
		dbconn.SetConnMaxLifetime(time.Second * 120)
		dbconn.SetConnMaxIdleTime(time.Second * 30)

		err = dbconn.Ping()
		if err != nil {
			writeLog("Error in database connection " + err.Error())
		}
	}

	return
}

func CheckTable(databasename, tablename string) (dbOk, tableOk bool, err error) {

	row := dbconn.QueryRow("select id from " + databasename + "." + tablename + " limit 1")
	err = row.Err()
	if err == nil {
		row.Scan()

		dbOk = true
		tableOk = true
	} else {
		if strings.Contains(err.Error(), "Unknown database") {
			dbOk = false
			tableOk = false
		} else {
			dbOk = false
			tableOk = false
		}
	}

	return
}

func createDatabase(databasename string) (success bool, err error) {

	_, err = dbconn.Exec("create database `" + databasename +
		"` CHARACTER SET utf8 COLLATE utf8_general_ci")
	success = err == nil
	if !success {
		writeLog("Error in createDatabase: " + err.Error())
	}

	return
}

func createAttachmentTable(databasename string) (success bool, err error) {

	_, err = dbconn.Exec(`CREATE TABLE ` + databasename + `.documents (
 		 id int NOT NULL AUTO_INCREMENT,
  		 DocID int NOT NULL,
  		 RevisionID varchar(25) NOT NULL,
  		 filename varchar(245) DEFAULT NULL,
  		 insertionTime timestamp NULL DEFAULT NULL,
  		 Content longblob,
  		 FileType varchar(25) DEFAULT NULL,
		 MD5 varchar(40) default NULL,
  		 PRIMARY KEY (id),
  		 UNIQUE KEY RevisionID_UNIQUE (RevisionID)	);`)

	success = err == nil
	if !success {
		writeLog("Error in createAttachmentTable: " + err.Error())
	}

	return
}

func createMD5Field(databasename string) (success bool, err error) {

	_, err = dbconn.Exec(`alter table ` + databasename + `.documents 
			              ADD COLUMN MD5 VARCHAR(40) null`)

	success = err == nil
	if !success {
		writeLog("Error in createMD5Field: " + err.Error())
	}

	return
}

func GetNewRevisionID(databasename string) (revisionID string) {

	revisionID = codeutils.GetMD5(time.Now().String())[:20]
	record, err := GetHistoryRecord(databasename, revisionID)
	if err == nil && record.RevisionID != "" {
		revisionID = GetNewRevisionID(databasename)
	}
	return
}

func strToTime(timeStr string) (timeResult time.Time) {
	timeResult, _ = time.Parse(time.DateTime, timeStr)
	return
}

func timeToStr(atime time.Time) (result string) {
	result = atime.Format(time.DateTime)
	return
}

func InsertDocumentInfo(databasename string, doc *DocumentType, info DocumentInfoType) (id int64, err error) {

	id = 0
	infoJson, _ := json.Marshal(info)
	st := `INSERT INTO ` + databasename + `.documents (revisionID, filename, year,
		   insertionTime, documentDate, updatedTime, userID, sectionID,
	       isRemoved, Description, info) values (?, ?, ?, now(), ?, now(), ?, ?, 0, ?, ?)`
	doc.RevisionID = GetNewRevisionID(databasename)

	var result sql.Result

	result, err = dbconn.Exec(st, doc.RevisionID, doc.FileName, doc.Year,
		doc.DocumentDate, doc.UserID, doc.SectionID, doc.Description, infoJson)

	if err == nil {
		id, _ = result.LastInsertId()

	} else {
		writeLog("Error in InsertDocumentInfo: " + err.Error())
	}

	return
}

func ImportDocumentInfo(databasename string, doc *DocumentType, info DocumentInfoType) (id int64, err error) {

	id = 0
	infoJson, _ := json.Marshal(info)

	st := `INSERT INTO ` + databasename + `.documents (revisionID, filename, year,
		   insertionTime, documentDate, updatedTime, userID, sectionID,
	       isRemoved, Description, info) values (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`

	insertTime := timeToStr(doc.InsertionTime)
	updatedTime := timeToStr(doc.InsertionTime)
	documentDate := doc.DocumentDate.Format(time.DateOnly)
	var result sql.Result

	result, err = dbconn.Exec(st, doc.RevisionID, doc.FileName, doc.Year,
		insertTime, documentDate, updatedTime,
		doc.UserID, doc.SectionID, doc.Description, infoJson)

	if err == nil {
		id, _ = result.LastInsertId()

	} else {
		writeLog("Error in ImportsDocumentInfo: " + err.Error())
	}

	return
}

func InsertAttachement(databasename, year string, docID int64, revisionID, filename string,
	buf *bytes.Buffer) (success bool, err error) {

	attachmentdatabase := databasename + year
	success = false
	var fileType string
	if strings.Contains(filename, ".") {
		fileType = filename[strings.LastIndex(filename, "."):]
	} else {
		fileType = ""
	}
	var tableOk bool
	_, tableOk, err = CheckTable(attachmentdatabase, "documents")
	if !tableOk {
		createDatabase(attachmentdatabase)
		tableOk, err = createAttachmentTable(attachmentdatabase)
	}

	if err == nil {
		fileMD5 := GetFileMD5(buf.Bytes())
		st := `INSERT INTO ` + attachmentdatabase + `.documents (docID, revisionID, 
		      filename, FileType, md5, Content, insertionTime) values (?, ?, ?, ?, ?, ?, now())`
		_, err = dbconn.Exec(st, docID, revisionID, filename, fileType, fileMD5, buf.Bytes())

		success = err == nil
		if !success {
			writeLog("Error in InsertAttachement: " + err.Error())
		}
	}

	return
}

type DocumentType struct {
	ID             int64
	RevisionID     string
	ShowEdit       bool
	AttachmentSize int64
	AttachmentRecordInputType
	Info DocumentInfoType
}

type DocumentInfoType struct {
	IsPublic bool
}

func readOneDocument(rows *sql.Rows) (doc DocumentType, err error) {

	var updatedTimeSQL sql.NullString
	var insertionTimeSQL sql.NullString
	var infoSQL sql.NullString
	var documentDateSql string
	err = rows.Scan(&doc.ID, &doc.RevisionID, &doc.FileName, &insertionTimeSQL,
		&documentDateSql, &updatedTimeSQL, &doc.UserID,
		&doc.SectionID, &doc.Year, &doc.Description, &doc.Removed, &infoSQL)
	doc.DocumentDate, _ = time.Parse(time.DateOnly, documentDateSql)

	if insertionTimeSQL.Valid {
		doc.InsertionTime = strToTime(insertionTimeSQL.String)
	}
	if updatedTimeSQL.Valid {
		doc.UpdatedTime = strToTime(updatedTimeSQL.String)
	} else {
		doc.UpdatedTime = doc.InsertionTime
	}

	if infoSQL.Valid {
		json.Unmarshal([]byte(infoSQL.String), &doc.Info)

	}
	return
}

func readDocumentsList(rows *sql.Rows) (docs []DocumentType, err error) {

	docs = make([]DocumentType, 0)
	for rows.Next() {
		var doc DocumentType

		doc, err = readOneDocument(rows)

		if err == nil {
			docs = append(docs, doc)
		} else {
			writeLog("Error in readDocumentsList scan: " + err.Error())
		}
	}
	return
}

func GetLastDocuments(databasename string, page int) (docs []DocumentType, err error) {

	stmt := SELECT_FROM_DOCUMENTS + databasename + `.documents 
		    where not isRemoved
			order by updatedTime desc limit ?,10`

	rows, err := dbconn.Query(stmt, (page-1)*10)
	if err != nil {
		writeLog("Error in GetLastDocuments: " + err.Error())
	} else {
		defer rows.Close()

		docs, err = readDocumentsList(rows)
	}
	return
}

func GetAllDocuments(databasename string) (docs []DocumentType, err error) {

	stmt := SELECT_FROM_DOCUMENTS + databasename + `.documents 
		    where not isRemoved
			order by id`

	rows, err := dbconn.Query(stmt)
	if err != nil {
		writeLog("Error in GetAllDocuments: " + err.Error())
	} else {
		defer rows.Close()

		docs, err = readDocumentsList(rows)
	}
	return
}

func GetDocumentByRevisionID(databasename, revisionID string) (doc DocumentType, err error) {

	stmt := SELECT_FROM_DOCUMENTS + databasename + `.documents 
	        where revisionID  = ?`
	var rows *sql.Rows
	rows, err = dbconn.Query(stmt, revisionID)
	if err != nil {
		writeLog("Error in GetDocumentByRevisionID: " + err.Error())
	} else {
		defer rows.Close()
		if rows.Next() {
			doc, err = readOneDocument(rows)
		} else {
			err = errors.New("Document not found")
		}

		if err != nil {
			writeLog("Error in GetDocumentByRevisionID scan: " + err.Error())
		}
	}

	return
}

func GetFileMD5(buf []byte) (md5hash string) {

	md5hash = fmt.Sprintf("%x", md5.Sum(buf))

	return
}

func GetAttachment(databasename, year, revisionID string) (closer io.ReadCloser,
	doc DocumentType, err error) {

	stmt := `SELECT filename, md5, content FROM ` + databasename + year + `.documents
		     where revisionID  = ?`

	row := dbconn.QueryRow(stmt, revisionID)
	err = row.Err()
	if err != nil {
		if strings.Contains(err.Error(), "md5") {
			createMD5Field(databasename + year)
		}
		writeLog("Error in GetAttachment: " + err.Error())

	} else {

		var buf []byte
		var fileMD5 sql.NullString
		err = row.Scan(&doc.FileName, &fileMD5, &buf)
		if fileMD5.Valid {
			doc.FileMD5 = fileMD5.String
		}
		closer = ioutil.NopCloser(bytes.NewReader(buf))
		doc.AttachmentSize = int64(len(buf))
		if doc.FileMD5 == "" {
			doc.FileMD5 = GetFileMD5(buf)
			UpdateAttachmentFileMD5(databasename, revisionID, year, doc.FileMD5)
		}

		if err != nil {
			writeLog("Error in GetAttachment scan: " + err.Error())
		}
	}

	return
}

func queryAttachmentInfo(databasename, year, revisionID string) (filename, fileMD5 string, err error) {

	stmt := `SELECT filename, md5 FROM ` + databasename + year + `.documents
		     where revisionID  = ?`

	row := dbconn.QueryRow(stmt, revisionID)
	err = row.Err()
	if err != nil {

		writeLog("Error in GetAttachmentFileName: " + err.Error())
	} else {

		err = row.Scan(&filename, &fileMD5)

		if err != nil {
			writeLog("Error in GetAttachmentFileName scan: " + err.Error())
		}
	}

	return
}

func GetAttachmentInfo(databasename, year, revisionID string) (filename, fileMD5 string, err error) {

	filename, fileMD5, err = queryAttachmentInfo(databasename, year, revisionID)
	if err != nil && strings.Contains(err.Error(), "md5") {
		createMD5Field(databasename + year)
		filename, fileMD5, err = queryAttachmentInfo(databasename, year, revisionID)
	}
	return

}

func GetDocumentsCount(databasename string) (count int, err error) {

	stmt := `SELECT count(*) from ` + databasename + `.documents where not isRemoved`

	row := dbconn.QueryRow(stmt)
	if err != nil {
		writeLog("Error in GetDocumentsCount: " + err.Error())
	} else {

		err = row.Scan(&count)

		if err != nil {
			writeLog("Error in GetDocumentsCount scan: " + err.Error())
		}
	}

	return

}

func SearchDocuments(databasename, text string) (docs []DocumentType, err error) {

	text = strings.ToLower(text)

	stmt := SELECT_FROM_DOCUMENTS + databasename + `.documents 
		where lower(filename) like '%` +
		text + `%' or lower(description) like '%` +
		text + `%' 		order by id desc`
	rows, err := dbconn.Query(stmt)
	if err != nil {
		writeLog("Error in SearchDocuments: " + err.Error())

	} else {
		defer rows.Close()

		docs, err = readDocumentsList(rows)
	}

	return
}

func UpdateDocumentInfo(databasename string, doc DocumentType) (success bool, err error) {

	st := `update ` + databasename + `.documents set revisionID = ?, description =?, 
	       documentDate = ?,  sectionID =?, year = ?, filename = ?, userID =?,
		   updatedTime = now(), info = ?
	       where ID = ?`
	info, _ := json.Marshal(doc.Info)
	_, err = dbconn.Exec(st, doc.RevisionID, doc.Description, doc.DocumentDate, doc.SectionID,
		doc.Year, doc.FileName, doc.UserID, info, doc.ID)
	success = err == nil
	if !success {
		writeLog("Error in UpdateDocumentInfo: " + err.Error())
	}

	return
}

func UpdateDocumentIsRemoved(databasename string, docID int64, isRemoved bool) (success bool, err error) {

	st := `update ` + databasename + `.documents set isRemoved =?
	       where ID = ?`
	_, err = dbconn.Exec(st, isRemoved, docID)
	success = err == nil
	if !success {
		writeLog("Error in UpdateDocumentIsRemoved: " + err.Error())
	}

	return
}

func UpdateAttachment(databasename string, docID int64, revisionID, year, filename, fileMD5 string,
	buf *bytes.Buffer) (success bool, err error) {

	var fileType string
	if strings.Contains(filename, ".") {
		fileType = filename[strings.LastIndex(filename, "."):]
	} else {
		fileType = ""
	}
	_, err = dbconn.Exec(`update `+databasename+`.documents 
			 set filename = ?, updatedTime=now()
		     where ID = ?`, filename, docID)
	if err == nil {
		st := `update ` + databasename + year + `.documents set filename = ?, fileType = ?,
		 md5=?, Content=?, insertionTime = now() where revisionID = ?`
		_, err = dbconn.Exec(st, filename, fileType, fileMD5, buf.Bytes(), revisionID)
		success = err == nil
		if !success {
			writeLog("Error in UpdateAttachment: " + err.Error())
		}
	}

	return
}

func UpdateAttachmentFileMD5(databasename string, revisionID, year, fileMD5 string) (success bool, err error) {

	st := `update ` + databasename + year + `.documents set MD5 = ?
		 	   where revisionID = ?`

	_, err = dbconn.Exec(st, fileMD5, revisionID)
	success = err == nil
	if !success {
		writeLog("Error in UpdateAttachmentFileMD5: " + err.Error())
	}

	return
}

func GetFilteredDocuments(databasename string, sectionID int) (docs []DocumentType, err error) {

	stmt := SELECT_FROM_DOCUMENTS + databasename + `.documents 
					where not isRemoved and SectionID = ?
				 	order by id desc`

	rows, err := dbconn.Query(stmt, sectionID)
	if err != nil {
		writeLog("Error in GetFilteredDocuments: " + err.Error())
	} else {
		defer rows.Close()

		docs, err = readDocumentsList(rows)
	}

	return
}

func InsertHistory(databasename, year, revisionID, event string, docID int64, userID int) (success bool, err error) {

	st := `INSERT INTO ` + databasename + `.history
		   (DocID, RevisionID, EventTime, UserID, Year, Event)
		   VALUES (?, ?, now(), ?, ?, ?)`

	_, err = dbconn.Exec(st, docID, revisionID, userID, year, event)
	success = err == nil
	if !success {
		writeLog("Error in InsertHistory: " + err.Error())
	}

	return
}

func InsertHistoryWithTime(databasename, year, revisionID, event string, docID int64, userID int, eventTime time.Time) (success bool, err error) {

	st := `INSERT INTO ` + databasename + `.history
		   (DocID, RevisionID, EventTime, UserID, Year, Event)
		   VALUES (?, ?, ?, ?, ?, ?)`

	_, err = dbconn.Exec(st, docID, revisionID, timeToStr(eventTime), userID, year, event)
	success = err == nil
	if !success {
		writeLog("Error in InsertHistoryWithTime: " + err.Error())
	}

	return
}

type HistoryType struct {
	ID         int
	RevisionID string
	EventTime  time.Time
	UserID     int
	UserName   string
	Filename   string
	Year       string
	Event      string
	FileMD5    string
}

func readHistoryRecord(databasename string, rows *sql.Rows) (record HistoryType, err error) {

	var eventTimeStr string
	err = rows.Scan(&record.ID, &record.RevisionID, &eventTimeStr, &record.UserID,
		&record.Year, &record.Event)

	if err == nil {
		record.EventTime = strToTime(eventTimeStr)
		userInfo, _ := GetUserByid(databasename, record.UserID)
		record.UserName = userInfo.UserName

		record.Filename, record.FileMD5, _ = GetAttachmentInfo(databasename, record.Year, record.RevisionID)

	} else {
		writeLog("Error in readHistoryRecord scan: " + err.Error())
	}
	return
}

func GetHistoryRecord(databasename string, revisionID string) (record HistoryType, err error) {

	query := `SELECT id, RevisionID, EventTime, UserID, Year, Event
		FROM ` + databasename + `.history
		WHERE revisionID=? order by id desc`

	rows, err := dbconn.Query(query, revisionID)
	if err != nil {
		writeLog("Error in GetHistoryRecord: " + err.Error())

	} else {
		defer rows.Close()

		if rows.Next() {
			record, err = readHistoryRecord(databasename, rows)
		} else {
			err = errors.New("Not found")
		}
	}
	return
}

func GetHistory(databasename string, docID int64) (history []HistoryType, err error) {

	query := `SELECT id, RevisionID, EventTime, UserID, Year, Event
	    	  FROM ` + databasename + `.history
	    	  WHERE docID=? order by id desc`

	rows, err := dbconn.Query(query, docID)
	if err != nil {
		writeLog("Error in GetHistory: " + err.Error())

	} else {
		defer rows.Close()

		for rows.Next() {
			var record HistoryType
			record, err = readHistoryRecord(databasename, rows)
			if err == nil {

				history = append(history, record)
			}
		}
	}
	return
}

func TempRemoveDocument(databasename string, theDoc DocumentType) (success bool, err error) {

	success, err = UpdateDocumentIsRemoved(databasename, theDoc.ID, true)

	return
}

func RestoreDocument(databasename string, theDoc DocumentType) (success bool, err error) {

	success, err = UpdateDocumentIsRemoved(databasename, theDoc.ID, false)

	return
}
