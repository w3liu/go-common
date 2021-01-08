package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	Info("hello")
}

func TestReplace(t *testing.T) {
	ReplaceGlobal(New(EnvProduct))
	ReplaceGlobal(New(EnvProduct))
}
