package util_test

import (
	"github.com/habiliai/apidepot/pkg/internal/util"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AESTestSuite struct {
	suite.Suite
}

func TestAES(t *testing.T) {
	suite.Run(t, new(AESTestSuite))
}

func (s *AESTestSuite) TestEncryptAndDecrypt() {
	expected := "simple_t"

	key := []byte("simple_key")
	cipherBytes, err := util.EncryptAES(key, []byte(expected))
	s.Require().NoError(err)

	s.T().Logf("cipher: %x\n", cipherBytes)

	plainBytes, err := util.DecryptAES(key, cipherBytes)
	s.Require().NoError(err)

	s.Equal(expected, string(plainBytes))
}
