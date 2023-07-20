package fftest

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TempFile returns the filename of a temporary file that has been created with
// the provided content. The file is created in t.TempDir(), which is
// automatically removed when the test finishes.
func TempFile(t *testing.T, content string) string {
	t.Helper()

	filename := filepath.Join(t.TempDir(), strconv.Itoa(rand.Int()))

	if err := os.WriteFile(filename, []byte(content), 0o0600); err != nil {
		t.Fatal(err)
	}

	t.Logf("created %s", filename)

	return filename
}
