package tclog_test

import (
	tclog "github.com/habiliai/apidepot/pkg/internal/log"
	"github.com/pkg/errors"
	"testing"
)

func TestLogger(t *testing.T) {
	var logger = tclog.GetLogger()
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error", tclog.Err(errors.New("test 123")))
}
