package internal

import (
    "fmt"
    "os"
    "path/filepath"
)

func ResolveDir(dirFlag string) (string, error) {
    if dirFlag == "" {
        return "", fmt.Errorf("no directory specified")
    }
    abs, err := filepath.Abs(dirFlag)
    if err != nil {
        return "", err
    }
    fi, err := os.Stat(abs)
    if err != nil {
        return "", fmt.Errorf("directory %s does not exist", abs)
    }
    if !fi.IsDir() {
        return "", fmt.Errorf("path %s is not a directory", abs)
    }
    return abs, nil
}
