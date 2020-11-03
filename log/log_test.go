package log

import "testing"

func TestLog(t *testing.T) {
	L().Info("hello")
	L().Error("error")
	L().Debug("debug")
	L().Warn("warn")
}
