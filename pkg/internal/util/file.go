package util

import (
	"github.com/pkg/errors"
	"io"
	"os"
)

func CopyFile(src, tgt string, force bool) error {
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open source file=%s", src)
	}
	defer srcFile.Close()

	if _, err := os.Stat(tgt); !force && err == nil {
		return errors.Errorf("target file already exists. filepath=%s", tgt)
	}

	tgtFile, err := os.OpenFile(tgt, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open target file=%s", tgt)
	}
	defer tgtFile.Close()

	written, err := io.Copy(tgtFile, srcFile)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file: %s -> %s", src, tgt)
	} else if written == 0 {
		return errors.Errorf("no data copied: %s -> %s", src, tgt)
	}

	return nil
}
