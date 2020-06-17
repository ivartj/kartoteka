package sqlutil

import (
	"strings"
)

type FormatBuilder struct {
	sb     strings.Builder
	params []interface{}
}

func (b *FormatBuilder) Add(format string, params ...interface{}) *FormatBuilder {
	b.sb.WriteString(format)
	if params == nil {
		b.params = params
	} else {
		b.params = append(b.params, params...)
	}
	return b
}

func (b *FormatBuilder) Format() string {
	return b.sb.String()
}

func (b *FormatBuilder) Args() []interface{} {
	if b.params == nil {
		return []interface{}{}
	} else {
		return b.params
	}
}
