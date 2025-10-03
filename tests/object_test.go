package tests

import (
    "testing"

    "github.com/tedba742/gitImplementation/internal/object"
)

func TestHashData(t *testing.T) {
    data := []byte("hello world")
    want := "95d09f2b10159347eece71399a7e2e907ea3df4f" // git blob SHA1 for "hello world"
    got := object.HashData("blob", data)
    if got != want {
        t.Fatalf("HashData() = %s; want %s", got, want)
    }
}
