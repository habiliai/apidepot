package stringbuilder_test

import (
	"github.com/habiliai/apidepot/pkg/internal/util/stringbuilder"
	"github.com/stretchr/testify/suite"
	"testing"
)

type StringBuilderTestSuite struct {
	suite.Suite
}

func TestStringBuilder(t *testing.T) {
	suite.Run(t, new(StringBuilderTestSuite))
}

func (s *StringBuilderTestSuite) TestStringBuilder_toString() {
	output := stringbuilder.New("test", stringbuilder.Indent(2)).
		AddField("field1", "value1").
		AddField("field2", 2).
		AddField("field3", 3.0).
		AddField("field4", true).
		AddField("field5", []string{"a", "b", "c"}).
		AddField("field6", map[string]string{"a": "b", "c": "d"}).
		AddField("field7", nil).
		String()

	expected := `test{
  field1: "value1",
  field2: 2,
  field3: 3,
  field4: true,
  field5: [a b c],
  field6: map[a:b c:d],
  field7: null,
}
`
	s.Equal(expected, output)
}
