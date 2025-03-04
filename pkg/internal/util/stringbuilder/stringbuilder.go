package stringbuilder

import (
	"fmt"
	"github.com/modern-go/reflect2"
	"reflect"
	"strings"
)

type StringBuilder struct {
	writer  *strings.Builder
	options Options
}

type Options struct {
	indent uint
}
type OptionsFunc func(*Options)

func New(objectName string, optionsFuncs ...OptionsFunc) *StringBuilder {
	sb := &StringBuilder{
		writer: &strings.Builder{},
		options: Options{
			indent: 2,
		},
	}
	sb.writer.WriteString(objectName)
	sb.writer.WriteString("{\n")

	if len(optionsFuncs) > 0 {
		for _, opt := range optionsFuncs {
			opt(&sb.options)
		}
	}

	return sb
}

func (sb *StringBuilder) AddField(fieldName string, fieldValue any) *StringBuilder {
	for i := uint(0); i < sb.options.indent; i++ {
		sb.writer.WriteRune(' ')
	}
	sb.writer.WriteString(fieldName)
	sb.writer.WriteString(": ")

	if fieldValue == nil {
		sb.writer.WriteString("null")
		sb.writer.WriteString(",\n")
		return sb
	}

	type_ := reflect2.TypeOf(fieldValue)
	if type_.Kind() == reflect.String {
		sb.writer.WriteString(fmt.Sprintf("\"%v\"", fieldValue))
	} else {
		sb.writer.WriteString(fmt.Sprintf("%v", fieldValue))
	}
	sb.writer.WriteString(",\n")

	return sb
}

func (sb *StringBuilder) String() string {
	sb.writer.WriteString("}\n")
	return sb.writer.String()
}

func Indent(indent uint) OptionsFunc {
	return func(o *Options) {
		o.indent = indent
	}
}
