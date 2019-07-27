package fftest

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// CreateTempFile creates a temporary file in os.TempDir
// with the given content. Call the cleanup func to
// remove the file.
func CreateTempFile(t *testing.T, content string) (filename string, cleanup func()) {
	t.Helper()

	filename = filepath.Join(os.TempDir(), "fftest_"+fmt.Sprint(10000*rand.Intn(10000)))
	f, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}

	f.Write([]byte(content))
	f.Close()
	return f.Name(), func() { os.Remove(f.Name()) }
}
