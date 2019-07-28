package fftest

import (
	"io/ioutil"
	"os"
	"testing"
)

// TempFile uses ioutil.TempFile to create a temporary file with the given
// content. Use the cleanup func to remove the file.
func TempFile(t *testing.T, content string) (filename string, cleanup func()) {
	t.Helper()

	f, err := ioutil.TempFile("", "fftest_")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	cleanup = func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("os.Remove(%q): %v", f.Name(), err)
		}
	}

	return f.Name(), cleanup
}
