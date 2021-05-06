package storage

import (
	"io"
	"testing"
)

//TODO FILE TESTS
func TestFile(t *testing.T) {
	var f io.ReadWriteSeeker = new(File)
	if f == nil {
		t.Error(f)
	}
}
