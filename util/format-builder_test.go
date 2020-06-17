package sqlutil

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatBuilder(t *testing.T) {
	var b FormatBuilder
	b.Add("Hey %s, ", "Bob")
	b.Add("how are you this %s?", "thursday")
	s := fmt.Sprintf(b.Format(), b.Args()...)
	assert.Equal(t, "Hey Bob, how are you this thursday?", s)
}
