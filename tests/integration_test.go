package tests

import (
    "path/filepath"
    "testing"
    "os"

    "github.com/tedba742/gitImplementation/internal/repo"
)

// TestFlow simulates a small workflow: create repo dirs, create a file, add, commit,
// checkout a new branch, make another commit, push to a remote, and print logs.
func TestFlow(t *testing.T) {
    // setup local repo
    local := t.TempDir()
    gitDir := filepath.Join(local, ".mygit")
    if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
        t.Fatal(err)
    }

    // initialize HEAD to master
    if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644); err != nil {
        t.Fatal(err)
    }

    // create file and add
    file := filepath.Join(local, "file.txt")
    if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := repo.Add(local, "file.txt"); err != nil {
        t.Fatalf("add failed: %v", err)
    }

    // commit
    ch, err := repo.CreateCommit(local, "first commit")
    if err != nil {
        t.Fatalf("commit failed: %v", err)
    }
    t.Logf("created commit %s", ch)

    // create and checkout feature branch
    branchRef := filepath.Join(gitDir, "refs", "heads", "feature")
    if err := os.WriteFile(branchRef, []byte(ch+"\n"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := repo.Checkout(local, "feature"); err != nil {
        t.Fatalf("checkout failed: %v", err)
    }

    // change file and commit on feature
    if err := os.WriteFile(file, []byte("hello feature"), 0644); err != nil {
        t.Fatal(err)
    }
    if err := repo.Add(local, "file.txt"); err != nil {
        t.Fatalf("add failed: %v", err)
    }
    ch2, err := repo.CreateCommit(local, "feature commit")
    if err != nil {
        t.Fatalf("commit 2 failed: %v", err)
    }
    t.Logf("created commit %s", ch2)

    // setup remote and push
    remote := t.TempDir()
    if err := os.MkdirAll(filepath.Join(remote, ".mygit", "refs", "heads"), 0755); err != nil {
        t.Fatal(err)
    }
    if err := repo.Push(local, remote); err != nil {
        t.Fatalf("push failed: %v", err)
    }

    // print logs from local and remote
    llog, err := repo.Log(local)
    if err != nil {
        t.Fatalf("log local failed: %v", err)
    }
    rlog, err := repo.Log(remote)
    if err != nil {
        t.Fatalf("log remote failed: %v", err)
    }
    t.Log("local log:\n" + llog)
    t.Log("remote log:\n" + rlog)
}
