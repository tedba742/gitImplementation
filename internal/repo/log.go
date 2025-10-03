package repo

import (
    "fmt"
    "sort"
    "strings"
)

// Commit is a minimal commit node used for visualization in tests.
type Commit struct {
    Hash    string
    Message string
    Parents []string
}

// Visualize returns a simple ASCII visualization of the commit graph starting from head.
// It walks the first-parent (mainline) chain and for each mainline commit prints any
// additional children (branches) as indented lines under their parent. This is intentionally
// small and deterministic for tests.
func Visualize(commits map[string]*Commit, head string) string {
    if head == "" {
        return ""
    }

    // build parent -> children map
    children := make(map[string][]string)
    for h, c := range commits {
        for _, p := range c.Parents {
            children[p] = append(children[p], h)
        }
    }

    // build mainline: follow first parent repeatedly
    mainline := []string{}
    cur := head
    for cur != "" {
        mainline = append(mainline, cur)
        commit, ok := commits[cur]
        if !ok || len(commit.Parents) == 0 {
            break
        }
        cur = commit.Parents[0]
    }

    var lines []string
    // helper to get commit by hash safely
    get := func(h string) *Commit {
        if c, ok := commits[h]; ok {
            return c
        }
        return &Commit{Hash: h, Message: "(missing)", Parents: nil}
    }

    // print mainline from head down to root
    for i, h := range mainline {
        c := get(h)
        lines = append(lines, fmt.Sprintf("* %s %s", c.Hash, c.Message))

        // list children of this commit that are not the next mainline commit
        extra := []string{}
        for _, ch := range children[h] {
            // next mainline commit (child) is at i-1 position in mainline
            var nextMain string
            if i > 0 {
                nextMain = mainline[i-1]
            }
            if ch == nextMain {
                continue
            }
            extra = append(extra, ch)
        }
        if len(extra) > 0 {
            sort.Strings(extra)
            for _, ch := range extra {
                cc := get(ch)
                lines = append(lines, fmt.Sprintf("  * %s %s", cc.Hash, cc.Message))
            }
        }
    }

    return strings.Join(lines, "\n") + "\n"
}
