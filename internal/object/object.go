package object

import (
    "bytes"
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// hashHeaderAndContent computes SHA1(header+content) where header is
// "<type> <size>\x00" like Git's object format.
func hashHeaderAndContent(objType string, content []byte) string {
    header := fmt.Sprintf("%s %d\x00", objType, len(content))
    h := sha1.New()
    h.Write([]byte(header))
    h.Write(content)
    return hex.EncodeToString(h.Sum(nil))
}

// HashData returns the hex SHA1 of the canonical object (type+size+null+content).
func HashData(objType string, content []byte) string {
    return hashHeaderAndContent(objType, content)
}

// WriteObject writes an object into objectsDir using the git two-level layout
// (objects/aa/bbbbb...) and returns the hex hash.
func WriteObject(objectsDir, objType string, content []byte) (string, error) {
    hash := hashHeaderAndContent(objType, content)
    dir := filepath.Join(objectsDir, hash[:2])
    if err := os.MkdirAll(dir, 0755); err != nil {
        return "", err
    }
    path := filepath.Join(dir, hash[2:])
    // if already exists, skip writing
    if _, err := os.Stat(path); err == nil {
        return hash, nil
    }
    data := append([]byte(fmt.Sprintf("%s %d\x00", objType, len(content))), content...)
    if err := os.WriteFile(path, data, 0644); err != nil {
        return "", err
    }
    return hash, nil
}

// ReadObject reads an object by hex hash from objectsDir and returns its type and content.
func ReadObject(objectsDir, hexHash string) (string, []byte, error) {
    path := filepath.Join(objectsDir, hexHash[:2], hexHash[2:])
    b, err := os.ReadFile(path)
    if err != nil {
        return "", nil, err
    }
    idx := bytes.IndexByte(b, 0)
    if idx < 0 {
        return "", nil, fmt.Errorf("bad object format")
    }
    header := string(b[:idx])
    parts := strings.SplitN(header, " ", 2)
    objType := parts[0]
    content := b[idx+1:]
    return objType, content, nil
}

// HexToBytes converts a hex hash string to raw bytes.
func HexToBytes(s string) ([]byte, error) {
    return hex.DecodeString(s)
}
