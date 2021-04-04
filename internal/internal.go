package internal

import (
	"github.com/JuulLabs-OSS/cbgo"
)

// this exists only so we can explicitly pin cbgo to our fork
func noop() {
	var _ cbgo.UUID
}
