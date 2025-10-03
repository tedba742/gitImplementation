package tests

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/tedba742/gitImplementation/internal/repo"
)

func TestAddMissingFile(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    _ = os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755)
    _ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644)

    if err := repo.Add(dir, "nofile.txt"); err == nil {
        t.Fatalf("Add should fail for missing file")
    }
}

func TestCreateCommitEmptyIndex(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    _ = os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755)
    _ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644)

    ch, err := repo.CreateCommit(dir, "msg")
    if err != nil {
        t.Fatalf("CreateCommit failed: %v", err)
    }
    if ch == "" {
        t.Fatalf("CreateCommit returned empty hash")
    }
}

func TestPushCreatesRemoteHead(t *testing.T) {
    local := t.TempDir()
    gitDir := filepath.Join(local, ".mygit")
    _ = os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755)
    _ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644)

    // make a commit
    f := filepath.Join(local, "f.txt")
    _ = os.WriteFile(f, []byte("x"), 0644)
    if err := repo.Add(local, "f.txt"); err != nil {
        t.Fatal(err)
    }
    ch, err := repo.CreateCommit(local, "m")
    if err != nil {
        t.Fatal(err)
    }

    remote := t.TempDir()
    _ = os.MkdirAll(filepath.Join(remote, ".mygit", "refs", "heads"), 0755)
    if err := repo.Push(local, remote); err != nil {
        t.Fatalf("push failed: %v", err)
    }
    // remote HEAD should point to master
    b, err := os.ReadFile(filepath.Join(remote, ".mygit", "HEAD"))
    if err != nil {
        t.Fatal(err)
    }
    if string(b) != "ref: refs/heads/master\n" {
        t.Fatalf("remote HEAD unexpected: %s", string(b))
    }
    // remote ref should equal commit
    rb, err := os.ReadFile(filepath.Join(remote, ".mygit", "refs", "heads", "master"))
    if err != nil {
        t.Fatal(err)
    }
    if string(rb) != ch+"\n" {
        t.Fatalf("remote ref mismatch %s != %s", string(rb), ch)
    }
}

func TestCheckoutMissingBranch(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    _ = os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755)
    _ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644)

    if err := repo.Checkout(dir, "nope"); err == nil {
        t.Fatalf("Checkout should fail for missing branch")
    }
}

func TestLogMissingHEAD(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    _ = os.MkdirAll(filepath.Join(gitDir, "objects"), 0755)
    // no HEAD file
    if _, err := repo.Log(dir); err == nil {
        t.Fatalf("Log should fail when HEAD missing")
    }
}

func TestCreateCommitMalformedIndex(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".mygit")
    _ = os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755)
    _ = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644)
    // write malformed index
    _ = os.WriteFile(filepath.Join(gitDir, "index"), []byte("malformedline\n"), 0644)

    if _, err := repo.CreateCommit(dir, "m"); err != nil {
        t.Fatalf("CreateCommit should tolerate malformed index: %v", err)
    }
}
