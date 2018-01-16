package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	l := New()
	log := l.Log
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	log.Critical("critical")

	err := l.SetLevel("critical")
	assert.Nil(t, err)
	log = l.Log
	// should only show critical
	log.Debug("debug")
	log.Info("info")
	log.Warn("warn")
	log.Error("error")
	log.Critical("critical")
}
