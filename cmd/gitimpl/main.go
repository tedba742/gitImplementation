package main

import (
    "fmt"
    "os"
    "github.com/tedba742/gitImplementation/internal/repo"
)

func main() {
    if len(os.Args) >= 3 && os.Args[1] == "checkout" {
        // assume current working dir
        cwd, _ := os.Getwd()
        branch := os.Args[2]
        if err := repo.Checkout(cwd, branch); err != nil {
            fmt.Fprintf(os.Stderr, "checkout failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("switched to branch %s\n", branch)
        return
    }
    if len(os.Args) >= 4 && os.Args[1] == "rebase" && os.Args[2] == "-i" {
        cwd, _ := os.Getwd()
        base := os.Args[3]
        // run interactive rebase without external editor (process picks only)
        if err := repo.InteractiveRebase(cwd, base, ""); err != nil {
            fmt.Fprintf(os.Stderr, "rebase failed: %v\n", err)
            os.Exit(1)
        }
        fmt.Println("rebase completed")
        return
    }
    fmt.Println("gitImplementation - experimental")
}
