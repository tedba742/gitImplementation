package repo

import (
    "bufio"
    "fmt"
    "io/fs"
    "os/exec"
    "os"
    "path/filepath"
    "strings"

    "github.com/tedba742/gitImplementation/internal/object"
)

// Checkout updates HEAD to point to the named branch. The repoPath should be the working
// directory containing the `.mygit` directory.
func Checkout(repoPath, branch string) error {
    gitDir := filepath.Join(repoPath, ".mygit")
    refPath := filepath.Join(gitDir, "refs", "heads", branch)
    if _, err := os.Stat(refPath); os.IsNotExist(err) {
        return fmt.Errorf("branch %s does not exist", branch)
    } else if err != nil {
        return err
    }

    headPath := filepath.Join(gitDir, "HEAD")
    content := []byte(fmt.Sprintf("ref: refs/heads/%s\n", branch))
    if err := os.WriteFile(headPath, content, 0644); err != nil {
        return err
    }
    return nil
}

// Add stages a file into the repository index. It computes the object hash of the file
// contents, stores the object under .mygit/objects/<hash>, and appends an entry to .mygit/index
// in the form "<hash> <path>\n".
func Add(repoPath, relPath string) error {
    gitDir := filepath.Join(repoPath, ".mygit")
    // read file content
    fp := filepath.Join(repoPath, relPath)
    b, err := os.ReadFile(fp)
    if err != nil {
        return err
    }
    objectsDir := filepath.Join(gitDir, "objects")
    if _, err := object.WriteObject(objectsDir, "blob", b); err != nil {
        return err
    }

    indexPath := filepath.Join(gitDir, "index")
    f, err := os.OpenFile(indexPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    // store the blob hash in index; compute hash using canonical header+content
    h := object.HashData("blob", b)
    if _, err := f.WriteString(fmt.Sprintf("%s %s\n", h, relPath)); err != nil {
        return err
    }
    return nil
}

// CreateCommit creates a new commit from the current index, writes it under .mygit/objects/commits/<hash>
// and updates the current branch ref. It returns the commit hash.
func CreateCommit(repoPath, message string) (string, error) {
    gitDir := filepath.Join(repoPath, ".mygit")
    // read HEAD to learn current branch
    headPath := filepath.Join(gitDir, "HEAD")
    hb, err := os.ReadFile(headPath)
    if err != nil {
        return "", err
    }
    head := strings.TrimSpace(string(hb))
    var branch string
    if strings.HasPrefix(head, "ref:") {
        branch = strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
        branch = filepath.Base(branch)
    } else {
        branch = "master"
    }

    // build tree from index (only root-level, no nested dirs in this simple impl)
    indexPath := filepath.Join(gitDir, "index")
    var entries []string
    if b, err := os.ReadFile(indexPath); err == nil {
        lines := strings.Split(strings.TrimSpace(string(b)), "\n")
        for _, ln := range lines {
            if ln == "" {
                continue
            }
            parts := strings.SplitN(ln, " ", 2)
            if len(parts) != 2 {
                continue
            }
            entries = append(entries, ln)
        }
    }

    objectsDir := filepath.Join(gitDir, "objects")

    // create a simple tree content: entries sorted by name, each line "<mode> <name>\x00<rawsha>" is canonical, but here
    // we'll use a simplified textual tree listing to generate a tree object that contains filenames and blob ids.
    var treeBuf strings.Builder
    for _, e := range entries {
        parts := strings.SplitN(e, " ", 2)
        treeBuf.WriteString(parts[0] + " " + parts[1] + "\n")
    }
    treeContent := []byte(treeBuf.String())
    treeHash, err := object.WriteObject(objectsDir, "tree", treeContent)
    if err != nil {
        return "", err
    }

    // determine parent
    refPath := filepath.Join(gitDir, "refs", "heads", branch)
    var parent string
    if b, err := os.ReadFile(refPath); err == nil {
        parent = strings.TrimSpace(string(b))
    }

    // build commit body like git: tree <sha>\nparent <sha>\n\n<message>\n
    var cb strings.Builder
    cb.WriteString("tree ")
    cb.WriteString(treeHash)
    cb.WriteString("\n")
    if parent != "" {
        cb.WriteString("parent ")
        cb.WriteString(parent)
        cb.WriteString("\n")
    }
    cb.WriteString("\n")
    cb.WriteString(message)
    cb.WriteString("\n")

    commitHash, err := object.WriteObject(objectsDir, "commit", []byte(cb.String()))
    if err != nil {
        return "", err
    }

    // update branch ref
    if err := os.MkdirAll(filepath.Dir(refPath), 0755); err != nil {
        return "", err
    }
    if err := os.WriteFile(refPath, []byte(commitHash+"\n"), 0644); err != nil {
        return "", err
    }

    // clear index
    _ = os.Remove(indexPath)

    return commitHash, nil
}

// Push copies the current branch ref from repoPath to remotePath's .mygit/refs/heads/<branch>.
// This is a lightweight simulation of push and does not copy objects.
func Push(repoPath, remotePath string) error {
    gitDir := filepath.Join(repoPath, ".mygit")
    headPath := filepath.Join(gitDir, "HEAD")
    hb, err := os.ReadFile(headPath)
    if err != nil {
        return err
    }
    head := strings.TrimSpace(string(hb))
    var branch string
    if strings.HasPrefix(head, "ref:") {
        branch = strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
        branch = filepath.Base(branch)
    } else {
        branch = "master"
    }

    // read our branch ref
    refPath := filepath.Join(gitDir, "refs", "heads", branch)
    rb, err := os.ReadFile(refPath)
    if err != nil {
        return err
    }

    // write into remote
    remoteRef := filepath.Join(remotePath, ".mygit", "refs", "heads", branch)
    if err := os.MkdirAll(filepath.Dir(remoteRef), 0755); err != nil {
        return err
    }
    if err := os.WriteFile(remoteRef, rb, 0644); err != nil {
        return err
    }
    // update remote HEAD to point to this branch
    remoteHead := filepath.Join(remotePath, ".mygit", "HEAD")
    if err := os.WriteFile(remoteHead, []byte(fmt.Sprintf("ref: refs/heads/%s\n", branch)), 0644); err != nil {
        return err
    }
    return nil
}

// Log builds a commit map from .mygit/objects/commits and returns a visualization string
// using Visualize. It reads HEAD to determine the branch head commit.
func Log(repoPath string) (string, error) {
    gitDir := filepath.Join(repoPath, ".mygit")
    // read HEAD
    headPath := filepath.Join(gitDir, "HEAD")
    hb, err := os.ReadFile(headPath)
    if err != nil {
        return "", err
    }
    head := strings.TrimSpace(string(hb))
    var branch string
    if strings.HasPrefix(head, "ref:") {
        branch = strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
        branch = filepath.Base(branch)
    } else {
        branch = "master"
    }

    refPath := filepath.Join(gitDir, "refs", "heads", branch)
    headHash := ""
    if b, err := os.ReadFile(refPath); err == nil {
        headHash = strings.TrimSpace(string(b))
    }

    objectsDir := filepath.Join(gitDir, "objects")
    commits := make(map[string]*Commit)

    // Walk object store and find commit objects
    _ = filepath.WalkDir(objectsDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil
        }
        if d.IsDir() {
            return nil
        }
        // path like objects/aa/bbbb...
        // reconstruct hash from dir+file
        dir := filepath.Base(filepath.Dir(path))
        file := filepath.Base(path)
        hex := dir + file
        typ, content, err := object.ReadObject(objectsDir, hex)
        if err != nil {
            return nil
        }
        if typ != "commit" {
            return nil
        }
        // parse commit content: lines until blank are headers like "tree <sha>", "parent <sha>"
        sc := bufio.NewScanner(strings.NewReader(string(content)))
        msg := ""
        parents := []string{}
        for sc.Scan() {
            line := sc.Text()
            if line == "" {
                // rest is message
                var rest strings.Builder
                for sc.Scan() {
                    rest.WriteString(sc.Text())
                    rest.WriteString("\n")
                }
                msg = strings.TrimSpace(rest.String())
                break
            }
            if strings.HasPrefix(line, "parent ") {
                parents = append(parents, strings.TrimSpace(strings.TrimPrefix(line, "parent ")))
            }
        }
        commits[hex] = &Commit{Hash: hex, Message: msg, Parents: parents}
        return nil
    })

    vis := Visualize(commits, headHash)
    return vis, nil
}

// InteractiveRebaseFromTodo processes a todo contents for an interactive rebase.
// todoContent expected lines: "pick <hash> <message>", or "reword <hash> <new message>", or "drop <hash>".
func InteractiveRebaseFromTodo(repoPath, base, todoContent string) error {
    gitDir := filepath.Join(repoPath, ".mygit")
    // find current head
    headPath := filepath.Join(gitDir, "HEAD")
    hb, err := os.ReadFile(headPath)
    if err != nil {
        return err
    }
    head := strings.TrimSpace(string(hb))
    var headHash string
    if strings.HasPrefix(head, "ref:") {
        // resolve ref
        ref := strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
        b, err := os.ReadFile(filepath.Join(gitDir, ref))
        if err != nil {
            return err
        }
        headHash = strings.TrimSpace(string(b))
    } else {
        headHash = head
    }

    // build list of commits from base (exclusive) to head (inclusive), oldest->newest
    var commits []string
    cur := headHash
    for cur != "" {
        if cur == base {
            // stop — base found; reverse later
            break
        }
        commits = append(commits, cur)
        // read commit to get parent
        typ, content, err := object.ReadObject(filepath.Join(gitDir, "objects"), cur)
        if err != nil {
            return err
        }
        if typ != "commit" {
            return fmt.Errorf("object %s not a commit", cur)
        }
        // parse parent if any
        sc := bufio.NewScanner(strings.NewReader(string(content)))
        parent := ""
        for sc.Scan() {
            line := sc.Text()
            if strings.HasPrefix(line, "parent ") {
                parent = strings.TrimSpace(strings.TrimPrefix(line, "parent "))
                break
            }
            if line == "" {
                break
            }
        }
        cur = parent
    }
    // reverse commits to oldest->newest
    for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
        commits[i], commits[j] = commits[j], commits[i]
    }

    // parse todoContent by lines
    var todoLines []string
    for _, ln := range strings.Split(strings.ReplaceAll(todoContent, "\r\n", "\n"), "\n") {
        ln = strings.TrimSpace(ln)
        if ln == "" || strings.HasPrefix(ln, "#") {
            continue
        }
        todoLines = append(todoLines, ln)
    }

    // map original commits to their parsed messages
    orig := map[string]struct{ tree, msg string }{}
    for _, c := range commits {
        typ, content, err := object.ReadObject(filepath.Join(gitDir, "objects"), c)
        if err != nil {
            return err
        }
        if typ != "commit" {
            return fmt.Errorf("object %s not commit", c)
        }
        sc := bufio.NewScanner(strings.NewReader(string(content)))
        tree := ""
        msg := ""
        for sc.Scan() {
            line := sc.Text()
            if line == "" {
                // rest is message
                var rest strings.Builder
                for sc.Scan() {
                    rest.WriteString(sc.Text())
                    rest.WriteString("\n")
                }
                msg = strings.TrimSpace(rest.String())
                break
            }
            if strings.HasPrefix(line, "tree ") {
                tree = strings.TrimSpace(strings.TrimPrefix(line, "tree "))
            }
        }
        orig[c] = struct{ tree, msg string }{tree: tree, msg: msg}
    }

    // now process todo lines and replay
    var lastNew string
    objectsDir := filepath.Join(gitDir, "objects")
    for _, ln := range todoLines {
        parts := strings.Fields(ln)
        if len(parts) < 2 {
            return fmt.Errorf("invalid todo line: %s", ln)
        }
        cmd := parts[0]
        oldHash := parts[1]
        var newMsg string
        if cmd == "reword" {
            if len(parts) >= 3 {
                newMsg = strings.Join(parts[2:], " ")
            } else {
                // default to original message
                newMsg = orig[oldHash].msg
            }
        } else {
            newMsg = orig[oldHash].msg
        }

        switch cmd {
        case "pick", "reword":
            // create new commit with same tree and parent = lastNew
            tree := orig[oldHash].tree
            var cb strings.Builder
            cb.WriteString("tree ")
            cb.WriteString(tree)
            cb.WriteString("\n")
            if lastNew != "" {
                cb.WriteString("parent ")
                cb.WriteString(lastNew)
                cb.WriteString("\n")
            }
            cb.WriteString("\n")
            cb.WriteString(newMsg)
            cb.WriteString("\n")
            newHash, err := object.WriteObject(objectsDir, "commit", []byte(cb.String()))
            if err != nil {
                return err
            }
            lastNew = newHash
        case "drop":
            // do nothing
            continue
        default:
            return fmt.Errorf("unsupported todo command: %s", cmd)
        }
    }

    // update current branch ref to lastNew
    if lastNew != "" {
        // determine current branch ref from HEAD
        head := strings.TrimSpace(string(hb))
        if strings.HasPrefix(head, "ref:") {
            ref := strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
            if err := os.WriteFile(filepath.Join(gitDir, ref), []byte(lastNew+"\n"), 0644); err != nil {
                return err
            }
        } else {
            // detached head — write HEAD directly
            if err := os.WriteFile(headPath, []byte(lastNew+"\n"), 0644); err != nil {
                return err
            }
        }
    }
    return nil
}

// InteractiveRebase opens an editor to allow interactive rebase. editorCmd is the
// executable to run; if empty the environment variable GIT_EDITOR is used; if that
// is empty the editor is treated as a no-op.
func InteractiveRebase(repoPath, base, editorCmd string) error {
    gitDir := filepath.Join(repoPath, ".mygit")
    // gather commits
    headPath := filepath.Join(gitDir, "HEAD")
    hb, err := os.ReadFile(headPath)
    if err != nil {
        return err
    }
    head := strings.TrimSpace(string(hb))
    var headHash string
    if strings.HasPrefix(head, "ref:") {
        ref := strings.TrimSpace(strings.TrimPrefix(head, "ref:"))
        b, err := os.ReadFile(filepath.Join(gitDir, ref))
        if err != nil {
            return err
        }
        headHash = strings.TrimSpace(string(b))
    } else {
        headHash = head
    }

    // collect commits until base
    var commits []string
    cur := headHash
    for cur != "" {
        if cur == base {
            break
        }
        commits = append(commits, cur)
        typ, content, err := object.ReadObject(filepath.Join(gitDir, "objects"), cur)
        if err != nil {
            return err
        }
        if typ != "commit" {
            return fmt.Errorf("object %s not commit", cur)
        }
        sc := bufio.NewScanner(strings.NewReader(string(content)))
        parent := ""
        for sc.Scan() {
            line := sc.Text()
            if strings.HasPrefix(line, "parent ") {
                parent = strings.TrimSpace(strings.TrimPrefix(line, "parent "))
                break
            }
            if line == "" {
                break
            }
        }
        cur = parent
    }
    // reverse to oldest->newest
    for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
        commits[i], commits[j] = commits[j], commits[i]
    }

    // build todo
    var todo strings.Builder
    for _, c := range commits {
        // read commit message
        _, content, err := object.ReadObject(filepath.Join(gitDir, "objects"), c)
        if err != nil {
            return err
        }
        sc := bufio.NewScanner(strings.NewReader(string(content)))
        // skip headers
        for sc.Scan() {
            if sc.Text() == "" {
                break
            }
        }
        // rest is message
        var rest strings.Builder
        for sc.Scan() {
            rest.WriteString(sc.Text())
            rest.WriteString(" ")
        }
        msg := strings.TrimSpace(rest.String())
        todo.WriteString(fmt.Sprintf("pick %s %s\n", c, msg))
    }

    // write todo to temp file
    tf, err := os.CreateTemp("", "gitimpl-rebase-*.todo")
    if err != nil {
        return err
    }
    tfName := tf.Name()
    if _, err := tf.WriteString(todo.String()); err != nil {
        return err
    }
    tf.Close()

    // choose editor
    ed := editorCmd
    if ed == "" {
        ed = os.Getenv("GIT_EDITOR")
    }
    if ed == "" {
        // no-op editor
        // read back file and process
        b, err := os.ReadFile(tfName)
        if err != nil {
            return err
        }
        return InteractiveRebaseFromTodo(repoPath, base, string(b))
    }

    // run editor with filename appended
    parts := strings.Fields(ed)
    parts = append(parts, tfName)
    cmd := exec.Command(parts[0], parts[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return err
    }

    b, err := os.ReadFile(tfName)
    if err != nil {
        return err
    }
    return InteractiveRebaseFromTodo(repoPath, base, string(b))
}
