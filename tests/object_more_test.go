package tests

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/tedba742/gitImplementation/internal/object"
)

func TestWriteReadObjectRoundtrip(t *testing.T) {
    dir := t.TempDir()
    // write blob
    content := []byte("hello-object")
    hash, err := object.WriteObject(dir, "blob", content)
    if err != nil {
        t.Fatalf("WriteObject failed: %v", err)
    }

    typ, got, err := object.ReadObject(dir, hash)
    if err != nil {
        t.Fatalf("ReadObject failed: %v", err)
    }
    if typ != "blob" {
        t.Fatalf("ReadObject type = %s; want blob", typ)
    }
    if !bytes.Equal(got, content) {
        t.Fatalf("ReadObject content mismatch")
    }
}

func TestHexToBytes(t *testing.T) {
    b, err := object.HexToBytes("a1")
    if err != nil {
        t.Fatalf("HexToBytes valid failed: %v", err)
    }
    if len(b) != 1 {
        t.Fatalf("HexToBytes length = %d; want 1", len(b))
    }

    if _, err := object.HexToBytes("zzz"); err == nil {
        t.Fatalf("HexToBytes should error on invalid hex")
    }
}

func TestWriteObjectIdempotent(t *testing.T) {
    dir := t.TempDir()
    content := []byte("same")
    h1, err := object.WriteObject(dir, "blob", content)
    if err != nil {
        t.Fatal(err)
    }
    // second write should succeed and return same hash
    h2, err := object.WriteObject(dir, "blob", content)
    if err != nil {
        t.Fatal(err)
    }
    if h1 != h2 {
        t.Fatalf("hash mismatch %s != %s", h1, h2)
    }
    // ensure file exists at two-level path
    p := filepath.Join(dir, h1[:2], h1[2:])
    if _, err := os.Stat(p); err != nil {
        t.Fatalf("object file missing: %v", err)
    }
}

func TestWriteObjectBadDir(t *testing.T) {
    td := t.TempDir()
    // create a file where objects dir should be
    objectsDir := filepath.Join(td, "objects")
    if err := os.WriteFile(objectsDir, []byte("notadir"), 0644); err != nil {
        t.Fatal(err)
    }
    // attempt to write object should fail because it cannot create subdir
    if _, err := object.WriteObject(objectsDir, "blob", []byte("x")); err == nil {
        t.Fatalf("WriteObject should fail when objectsDir is a file")
    }
}
