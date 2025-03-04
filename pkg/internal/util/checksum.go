package util

import (
	"crypto/md5"
	"fmt"
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
)

var logger = tclog.GetLogger()

func GetChecksum(data []byte) (string, error) {
	hasher := md5.New()

	_, err := hasher.Write(data)
	if err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	return checksum, nil
}
