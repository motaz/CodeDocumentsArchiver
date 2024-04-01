package archiverdata

import (
	"fmt"

	"testing"
)

func TestAnyName(t *testing.T) {

	db, err := OpenConnection("localhost")

	if err != nil {
		fmt.Printf(err.Error())
	} else {
		dbname := "newdb"
		dbOk, tableOk, err := CheckTable(db, dbname+".users")

		fmt.Printf("dbOk: %v, table Ok: %v, Error: %v\n", dbOk, tableOk, err)
		if !dbOk {
			_, err := createDatabase(db, dbname)
			if err == nil {
				fmt.Println("Database created: ", dbname)
			} else {
				fmt.Println(err.Error())
			}
		}
	}

}
