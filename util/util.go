package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
)

var ErrExistingAccount = errors.New("existing account detected, refusing to overwrite")
var ErrNoAccount = errors.New("no account detected, register one first")
var ErrTOSNotAccepted = errors.New("TOS not accepted")

func RandomHexString(count int) string {
	b := make([]byte, count)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}

func GetTimestamp() string {
	return getTimestamp(time.Now())
}

func getTimestamp(t time.Time) string {
	timestamp := t.Format(time.RFC3339Nano)
	return timestamp
}

func IsHttp500Error(err error) bool {
	return err.Error() == "500 Internal Server Error"
}

func FormatMessage(shortMessage string, longMessage string) string {
	if longMessage != "" {
		longMessage = strings.TrimPrefix(longMessage, "\n")
		longMessage = strings.ReplaceAll(longMessage, "\n", " ")
	}
	if shortMessage != "" && longMessage != "" {
		return shortMessage + ". " + longMessage
	} else if shortMessage != "" {
		return shortMessage
	} else {
		return longMessage
	}
}

func RunCommandFatal(cmd func() error) {
	if err := cmd(); err != nil {
		expectedErrs := []error{ErrNoAccount, ErrExistingAccount, ErrTOSNotAccepted}
		if slices.ContainsFunc(expectedErrs, func(e error) bool { return errors.Is(err, e) }) {
			log.Fatalln(err)
		} else {
			log.Fatalf("%+v\n", err)
		}
	}
}
