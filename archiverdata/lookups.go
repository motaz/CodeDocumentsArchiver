package archiverdata

type DocumentSection struct {
	SectionID   int
	SectionName string
}

type Documents struct {
	Docs []DocumentSection
}

type sectionsIndex struct {
	FieldName string `json:"SectionID"`
}

type UsersDoc struct {
	Docs []UserType
}

func CheckUsers(databasename string) (ThereIsUser bool, err error) {

	users, err := GetUsers(databasename)
	ThereIsUser = users != nil && len(users) > 0
	return
}

func GetSections(databasename string) (sections []DocumentSection, err error) {

	stmt := `SELECT  sectionID, SectionName from ` + databasename + `.sections`

	rows, err := dbconn.Query(stmt)
	if err != nil {
		writeLog("Error in GetSections: " + err.Error())
	} else {
		defer rows.Close()

		for rows.Next() {
			var section DocumentSection

			err = rows.Scan(&section.SectionID, &section.SectionName)

			if err == nil {
				sections = append(sections, section)
			} else {
				writeLog("Error in GetSections scan: " + err.Error())
			}
		}
	}

	return
}

func GetSectionName(databasename string, sectionID int) (sectionName string, err error) {

	stmt := `SELECT  SectionName from ` + databasename + `.sections`

	row := dbconn.QueryRow(stmt)
	err = row.Scan(&sectionName)
	if err != nil {
		writeLog("Error in GetSectionName: " + err.Error())
	}

	return
}

func InsertSection(databasename string, sectionID int, sectionName string) (success bool, err error) {

	st := `insert into ` + databasename + `.sections (sectionID, sectionName) values (?, ?)`

	_, err = dbconn.Exec(st, sectionID, sectionName)
	success = err == nil
	if !success {
		writeLog("Error in InsertSection: " + err.Error())
	}

	return
}
