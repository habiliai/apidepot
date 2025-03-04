package services_test

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type ServicesTestSuite struct {
	suite.Suite
}

func TestServices(t *testing.T) {
	suite.Run(t, new(ServicesTestSuite))
}
