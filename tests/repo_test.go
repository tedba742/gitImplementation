package tests

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/tedba742/gitImplementation/internal/repo"
)

func TestCheckoutWritesHEAD(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
        t.Fatal(err)
    }

    // create a branch ref file
    branch := "feature"
    refPath := filepath.Join(gitDir, "refs", "heads", branch)
    if err := os.WriteFile(refPath, []byte("abc123\n"), 0644); err != nil {
        t.Fatal(err)
    }

    if err := repo.Checkout(dir, branch); err != nil {
        t.Fatalf("Checkout failed: %v", err)
    }

    // read HEAD
    headPath := filepath.Join(gitDir, "HEAD")
    b, err := os.ReadFile(headPath)
    if err != nil {
        t.Fatalf("reading HEAD: %v", err)
    }
    got := string(b)
    want := "ref: refs/heads/feature\n"
    if got != want {
        t.Fatalf("HEAD = %q; want %q", got, want)
    }
}
