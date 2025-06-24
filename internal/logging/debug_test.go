package logging

import (
	"os"
	"testing"
)

func TestDebugEnabled(t *testing.T) {
	// Test with TT_DEBUG not set
	os.Unsetenv("TT_DEBUG")
	if DebugEnabled() {
		t.Error("DebugEnabled() should return false when TT_DEBUG is not set")
	}

	// Test with TT_DEBUG set to empty string
	os.Setenv("TT_DEBUG", "")
	if DebugEnabled() {
		t.Error("DebugEnabled() should return false when TT_DEBUG is empty")
	}

	// Test with TT_DEBUG set to any value
	os.Setenv("TT_DEBUG", "1")
	if !DebugEnabled() {
		t.Error("DebugEnabled() should return true when TT_DEBUG is set")
	}

	// Test with TT_DEBUG set to "true"
	os.Setenv("TT_DEBUG", "true")
	if !DebugEnabled() {
		t.Error("DebugEnabled() should return true when TT_DEBUG is 'true'")
	}

	// Clean up
	os.Unsetenv("TT_DEBUG")
}

func TestDebugf(t *testing.T) {
	// This test verifies that Debugf doesn't panic
	// We can't easily capture stdout in tests, so we just ensure it doesn't crash

	// Test with debug disabled
	os.Unsetenv("TT_DEBUG")
	Debugf("This should not appear: %s", "test")

	// Test with debug enabled
	os.Setenv("TT_DEBUG", "1")
	Debugf("This should appear: %s", "test")

	// Clean up
	os.Unsetenv("TT_DEBUG")
}

func TestDebugln(t *testing.T) {
	// This test verifies that Debugln doesn't panic
	// We can't easily capture stdout in tests, so we just ensure it doesn't crash

	// Test with debug disabled
	os.Unsetenv("TT_DEBUG")
	Debugln("This should not appear")

	// Test with debug enabled
	os.Setenv("TT_DEBUG", "1")
	Debugln("This should appear")

	// Clean up
	os.Unsetenv("TT_DEBUG")
}
