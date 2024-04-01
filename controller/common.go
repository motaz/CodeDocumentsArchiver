package controller

import (
	"CodeDocumentsArchiver/archiverdata"
	"bytes"
	"crypto/md5"
	"encoding/hex"

	"strconv"
	"time"
)

func InitDB() (err error) {
	archiverdata.InitRedis()
	err = archiverdata.OpenConnection()
	return
}

func GetMD5Hash(text string) string {

	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetCurrentYear() string {
	return strconv.Itoa(time.Now().Year())
}

func GetFileMD5(buf *bytes.Buffer) (md5hash string) {

	md5hash = archiverdata.GetFileMD5(buf.Bytes())
	return
}
