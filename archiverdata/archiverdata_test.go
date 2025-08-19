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
		for i := 0; i < 10; i++ {
			revisionID := GetNewRevisionID("localhost", "test1.png")
			fmt.Println(revisionID)
		}
	}

}
