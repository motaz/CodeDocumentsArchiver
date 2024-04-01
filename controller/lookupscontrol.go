package controller

import (
	"CodeDocumentsArchiver/archiverdata"
)

func GetDocumentTypes(domain string) (sections []archiverdata.DocumentSection, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	sections, err = archiverdata.GetSections(databasename)

	return
}

func GetSectionName(domain string, sectionID int) (sectionName string) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)
	sectionName, _ = archiverdata.GetSectionName(databasename, sectionID)
	return
}

func GetUserByID(domain string, userID int) (userInfo archiverdata.UserType) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)
	userInfo, _ = archiverdata.GetUserByid(databasename, userID)
	return
}

func ThereIsUser(domain string) (yes bool, err error) {

	databasename := archiverdata.GetDatabaseNameFromDomain(domain)

	yes, err = archiverdata.CheckUsers(databasename)

	return
}
