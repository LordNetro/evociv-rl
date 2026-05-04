package main

import (
	"os"
	"testing"
	"time"
)

func TestRunNoPanic(t *testing.T) {
	done := make(chan struct{})
	var runErr error
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("run() panicked: %v", r)
			}
		}()
		_ = os.Chdir("../..")
		runErr = run()
	}()
	select {
	case <-done:
		_ = runErr // acceptable to error when no TTY
	case <-time.After(2 * time.Second):
		t.Error("run() timed out")
	}
}
