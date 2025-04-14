package archiverdata

import (
	"fmt"

	"testing"
)

func TestAnyName(t *testing.T) {

	err := OpenConnection()

	if err != nil {
		fmt.Printf(err.Error())
	} else {
		dbname := "newdb"
		dbOk, tableOk, err := CheckTable(dbname, "users")

		fmt.Printf("dbOk: %v, table Ok: %v, Error: %v\n", dbOk, tableOk, err)
		if !dbOk {
			_, err := createDatabase(dbname)
			if err == nil {
				fmt.Println("Database created: ", dbname)
			} else {
				fmt.Println(err.Error())
			}
		}
	}

}
