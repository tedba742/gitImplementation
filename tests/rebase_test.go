package tests

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/tedba742/gitImplementation/internal/object"
    "github.com/tedba742/gitImplementation/internal/repo"
)

// build a sequence A->B, then rebase interactively to reword B
func TestInteractiveRebaseReword(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
        t.Fatal(err)
    }
    if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644); err != nil {
        t.Fatal(err)
    }

    // create file and commit A
    f := filepath.Join(dir, "f.txt")
    if err := os.WriteFile(f, []byte("a"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := repo.Add(dir, "f.txt"); err != nil {
        t.Fatal(err)
    }
    a, err := repo.CreateCommit(dir, "A")
    if err != nil {
        t.Fatal(err)
    }

    // modify and commit B
    if err := os.WriteFile(f, []byte("b"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := repo.Add(dir, "f.txt"); err != nil {
        t.Fatal(err)
    }
    b, err := repo.CreateCommit(dir, "B")
    if err != nil {
        t.Fatal(err)
    }

    // interactive rebase from base=a, reword b
    todo := strings.Join([]string{
        "pick " + a + " A",
        "reword " + b + " NewB",
    }, "\n")
    if err := repo.InteractiveRebaseFromTodo(dir, a, todo); err != nil {
        t.Fatalf("rebase failed: %v", err)
    }

    // ensure master ref was updated to new commit (last commit hash)
    refb, err := os.ReadFile(filepath.Join(gitDir, "refs", "heads", "master"))
    if err != nil {
        t.Fatal(err)
    }
    newHash := strings.TrimSpace(string(refb))
    if newHash == a || newHash == b {
        t.Fatalf("expected new hash to differ after rebase")
    }

    // check that the last commit message is NewB
    typ, content, err := object.ReadObject(filepath.Join(gitDir, "objects"), newHash)
    if err != nil {
        t.Fatal(err)
    }
    if typ != "commit" {
        t.Fatalf("expected commit object")
    }
    if !strings.Contains(string(content), "NewB") {
        t.Fatalf("expected commit message to be NewB, got: %s", string(content))
    }
}
