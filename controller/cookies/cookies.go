package cookies

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	forumDB "forum/model/functions"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")

	db *sql.DB
)

func Write(w http.ResponseWriter, cookie http.Cookie) error {
	// check if total lenght exceeds 4096 bytes
	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	// write cookie
	http.SetCookie(w, &cookie)

	return nil
}

func Read(r *http.Request, name string) (string, error) {
	// read cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	// decode the base64-encoded cookie value
	// if the cookie didn't contain a valid encoded value, return error
	/* 	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	   	if err != nil {
	   		return "", ErrInvalidValue
	} */

	// return decoded cookie value
	return string(cookie.Value), nil
}

func WriteSessionCookie(w http.ResponseWriter, r *http.Request, userID int64) error {

	// define session ID
	sessionID := strconv.Itoa(rand.Intn(100)) + time.Now().String()

	// encode session ID
	sessionID = base64.URLEncoding.EncodeToString([]byte(sessionID))

	//  check if sessionID exists, by checking that fetchSession returns an empty object
	retrievedSession, _ := forumDB.FetchSession(db, sessionID)
	if retrievedSession != (forumDB.Session{}) {
		err := fmt.Errorf("session already exists")
		return err
	}

	// write cookie to db
	// userID is 1/placeholder just for testing
	err := forumDB.InsertSession(db, sessionID, userID)
	if err != nil {
		return err
	}

	// init new cookie with 1 hour expiration
	cookie := http.Cookie{
		Name:     "sessionCookie",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	err = Write(w, cookie)
	if err != nil {
		return err
	}

	return nil
}

// remove session cookie and relational table in DB
func EndSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("sessionCookie")
	if err != nil {
		return err
	}
	_, err = forumDB.DeleteSession(db, cookie.Value)
	if err != nil {
		return err
	}
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	return nil
}

// temporary - set db var so other functions in the package can use it
func SetDB(database *sql.DB) {
	db = database
}
