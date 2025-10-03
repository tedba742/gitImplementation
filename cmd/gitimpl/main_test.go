package main

import (
    "bytes"
    "io"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestMainPrints(t *testing.T) {
    // capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    // run main in-process
    main()

    // restore and read
    _ = w.Close()
    var buf bytes.Buffer
    _, _ = io.Copy(&buf, r)
    os.Stdout = old
    s := buf.String()
    if !strings.Contains(s, "gitImplementation - experimental") {
        t.Fatalf("unexpected output: %s", s)
    }
}

func TestMainCheckoutCmd(t *testing.T) {
    // create temp repo and branch
    td := t.TempDir()
    gitDir := filepath.Join(td, ".mygit")
    if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
        t.Fatal(err)
    }
    if err := os.WriteFile(filepath.Join(gitDir, "refs", "heads", "feature"), []byte("abc\n"), 0644); err != nil {
        t.Fatal(err)
    }
    // change cwd and set args
    oldwd, _ := os.Getwd()
    defer os.Chdir(oldwd)
    if err := os.Chdir(td); err != nil {
        t.Fatal(err)
    }

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"gitimpl", "checkout", "feature"}

    // capture stdout
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    main()
    _ = w.Close()
    var buf bytes.Buffer
    _, _ = io.Copy(&buf, r)
    os.Stdout = old
    s := buf.String()
    if !strings.Contains(s, "switched to branch feature") {
        t.Fatalf("unexpected output: %s", s)
    }
}
