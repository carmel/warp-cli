package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
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

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", src, err)
	}
	defer sourceFile.Close()

	// 创建目标文件（如果已存在会覆盖）
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dst, err)
	}
	defer destFile.Close()

	// 高效复制内容
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	// 同步到磁盘（可选，但推荐）
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}
