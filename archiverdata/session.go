package archiverdata

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/motaz/redisaccess"
)

func InitRedis() (err error) {
	_, err = redisaccess.InitRedisLocalhost()
	return
}

type SessionInfoType struct {
	UserID      int
	UserName    string
	IsAdmin     bool
	SessionInfo string
	SessionID   string
	KeepLogin   bool
	SessionTime time.Time
}

func InsertSession(session SessionInfoType) (err error) {

	session.SessionTime = time.Now()
	if session.UserID > 0 {
		err = redisaccess.SetValue("cda-session::"+session.SessionID, session,
			time.Hour*time.Duration(GetSessionExpiary(session.KeepLogin)))
		if err != nil {
			writeLog("Error in InsertSession: " + err.Error())
		}
	} else {
		errors.New("UserID is 0")
		writeLog("Error in InsertSesson: UserID is 0 ")
	}

	return
}

func GetSessionExpiary(keepLogin bool) (expiaryHours int) {

	if keepLogin {
		expiaryHours = 24 * 30
	} else {
		expiaryHours = 8
	}
	return
}

func GetSession(sessionID string) (session SessionInfoType, err error) {

	value, _, err := redisaccess.GetValue("cda-session::" + sessionID)
	if err != nil {

		writeLog("Error in GetSession: " + err.Error())
	} else {
		err = json.Unmarshal([]byte(value), &session)
	}

	return

}
