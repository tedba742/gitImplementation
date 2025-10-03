package tests

import (
    "os/exec"
    "strings"
    "testing"
)

func TestCLIPrints(t *testing.T) {
    cmd := exec.Command("go", "run", "./cmd/gitimpl")
    // ensure the command runs from the repository root (parent of tests/)
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
