package main

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/tedba742/gitImplementation/internal/repo"
)

func usage() {
    fmt.Println("interactive commands:")
    fmt.Println("  init <path>             - create a minimal repo at path (creates .mygit)")
    fmt.Println("  add <file>              - stage a file")
    fmt.Println("  commit <message>        - commit staged files")
    fmt.Println("  checkout <branch>       - update HEAD to branch")
    fmt.Println("  push <remote-path>      - push refs to remote path")
    fmt.Println("  log                     - show commit visualization")
    fmt.Println("  rebase <base>           - interactive rebase (no editor; processes picks)")
    fmt.Println("  exit                    - quit")
}

func main() {
    cwd, _ := os.Getwd()
    fmt.Printf("working dir: %s\n", cwd)
    r := bufio.NewReader(os.Stdin)
    usage()
    for {
        fmt.Print("gitimpl> ")
        line, err := r.ReadString('\n')
        if err != nil {
            fmt.Println("error reading input:", err)
            return
        }
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        parts := strings.Fields(line)
        cmd := parts[0]
        args := parts[1:]
        switch cmd {
        case "help":
            usage()
        case "exit", "quit":
            return
        case "init":
            path := "."
            if len(args) >= 1 {
                path = args[0]
            }
            gitDir := filepath.Join(path, ".mygit")
            if err := os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755); err != nil {
                fmt.Println("init failed:", err)
                continue
            }
            if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/master\n"), 0644); err != nil {
                fmt.Println("init failed:", err)
                continue
            }
            fmt.Println("initialized", path)
        case "add":
            if len(args) < 1 {
                fmt.Println("usage: add <file>")
                continue
            }
            if err := repo.Add(cwd, args[0]); err != nil {
                fmt.Println("add failed:", err)
            } else {
                fmt.Println("added", args[0])
            }
        case "commit":
            if len(args) < 1 {
                fmt.Println("usage: commit <message>")
                continue
            }
            msg := strings.Join(args, " ")
            h, err := repo.CreateCommit(cwd, msg)
            if err != nil {
                fmt.Println("commit failed:", err)
            } else {
                fmt.Println("commit:", h)
            }
        case "checkout":
            if len(args) < 1 {
                fmt.Println("usage: checkout <branch>")
                continue
            }
            if err := repo.Checkout(cwd, args[0]); err != nil {
                fmt.Println("checkout failed:", err)
            } else {
                fmt.Println("switched to branch", args[0])
            }
        case "push":
            if len(args) < 1 {
                fmt.Println("usage: push <remote-path>")
                continue
            }
            if err := repo.Push(cwd, args[0]); err != nil {
                fmt.Println("push failed:", err)
            } else {
                fmt.Println("pushed to", args[0])
            }
        case "log":
            l, err := repo.Log(cwd)
            if err != nil {
                fmt.Println("log failed:", err)
            } else {
                fmt.Println(l)
            }
        case "rebase":
            if len(args) < 1 {
                fmt.Println("usage: rebase <base>")
                continue
            }
            if err := repo.InteractiveRebase(cwd, args[0], ""); err != nil {
                fmt.Println("rebase failed:", err)
            } else {
                fmt.Println("rebase completed")
            }
        default:
            fmt.Println("unknown cmd; type 'help'")
        }
    }
}
