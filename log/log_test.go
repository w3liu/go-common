package log

import "testing"

func TestLog(t *testing.T) {
	Logger.Info("hello")
	Logger.Error("error")
	Logger.Debug("debug")
	Logger.Warn("warn")
}
