package main

import (
	"crypto/md5"
	"encoding/hex"

	"math/rand"
	"net/http"
	"time"

	"github.com/motaz/codeutils"
)

func getConfigurationParameter(domain, paramname string) (value string) {

	value = codeutils.GetConfigValue(domain+".ini", paramname)
	return
}

func getRandom(r int) int {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(r)
}

func getCookie(w http.ResponseWriter, req *http.Request, name string) (value string) {

	cookie, err := req.Cookie(name)

	if err == nil {
		value = cookie.Value
	}
	return
}

func setCookies(w http.ResponseWriter, r *http.Request, name, value string, remember bool) {

	var expiration time.Time
	if remember {
		expiration = time.Now().Add(time.Hour * 24 * 90)
	} else {
		expiration = time.Now().Add(time.Hour * 8)

	}
	cookie := http.Cookie{Name: name, Value: value, Expires: expiration}

	http.SetCookie(w, &cookie)
	return
}

func writeLog(event string) {

	codeutils.WriteToLog(event, "cda")
}

func getMD5Hash(text string) string {

	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
