package archiverdata

import (
	"os"

	"strings"
	"time"

	"github.com/motaz/codeutils"
)

func writeLog(event string) {
	codeutils.WriteToLog(event, "cda")
}

func GetConfigValue(domain, paramName string, defaultValue string) string {

	value := codeutils.GetConfigValue(domain+".ini", paramName)
	if value == "" {
		value = defaultValue
	}
	return value
}

func GetCDAConfigValue(paramName string, defaultValue string) (avalue string) {

	avalue = codeutils.GetConfigValue("config.ini", paramName)
	if avalue == "" {
		avalue = defaultValue
	}
	return
}

func GetDatabaseNameFromDomain(domain string) (databasename string) {

	databasename = GetConfigValue(domain, "dbname", "")
	return
}

type AttachmentRecordInputType struct {
	FileName      string `json:"filename"`
	Type          string
	Description   string    `json:"description"`
	InsertionTime time.Time `json:"insertiontime"`
	DocumentDate  time.Time `json:"documentdate"`
	UpdatedTime   time.Time `json:"-"`
	SectionID     int
	SectionName   string `json:"-"`
	Year          string
	FileMD5       string

	UserID       int
	UserName     string `json:"-"`
	Removed      bool   `json:"removed"`
	IsNew        bool
	IsNewUpdated bool
}

type AttachmentRecordOutputType struct {
	ID       string `json:"_id"`
	Rev      string `json:"_rev"`
	ShowEdit bool   `json:"-"`

	AttachmentRecordInputType
}

func GetOnlyFile(filename string) string {

	sep := string(os.PathSeparator)
	if strings.Contains(filename, sep) {
		filename = filename[strings.LastIndex(filename, sep)+1:]
	}
	return filename
}

type UserResult struct {
	Docs []UserType
}
