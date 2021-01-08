package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	Info("hello")
}

func TestReplace(t *testing.T) {
	if err := ReplaceGlobal(New(EnvProduct)); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceGlobal(New(EnvProduct)); err != nil {
		t.Fatal(err)
	}
}
