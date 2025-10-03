package tests

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func TestMainPrints_InTests(t *testing.T) {
    cmd := exec.Command("go", "run", "./cmd/gitimpl")
    cmd.Dir = ".."
    out, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("go run failed: %v, output: %s", err, string(out))
    }
    s := string(out)
    if !strings.Contains(s, "gitImplementation - experimental") {
        t.Fatalf("unexpected output: %s", s)
    }
}

func TestMainCheckoutCmd_InTests(t *testing.T) {
    td := t.TempDir()
    gitDir := filepath.Join(td, ".mygit")
    if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
        t.Fatal(err)
    }
    if err := os.WriteFile(filepath.Join(gitDir, "refs", "heads", "feature"), []byte("abc\n"), 0644); err != nil {
        t.Fatal(err)
    }
    // build the binary first from module root
    binDir := t.TempDir()
    bin := filepath.Join(binDir, "gitimpl")
    if os.PathSeparator == '\\' {
        bin = bin + ".exe"
    }
    bcmd := exec.Command("go", "build", "-o", bin, "./cmd/gitimpl")
    bcmd.Dir = ".."
    if bout, err := bcmd.CombinedOutput(); err != nil {
        t.Fatalf("go build failed: %v, out=%s", err, string(bout))
    }

    cmd := exec.Command(bin, "checkout", "feature")
    // run the binary with working dir set to the temp repo
    cmd.Dir = td
    out, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("go run failed: %v, output: %s", err, string(out))
    }
    s := string(out)
    if !strings.Contains(s, "switched to branch feature") {
        t.Fatalf("unexpected output: %s", s)
    }
}
