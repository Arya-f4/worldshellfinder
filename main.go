package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
)

const defaultWordlist = `eval
base64_encode
base64_decode
exec
shell_exec
system
passthru
popen`

const banner = `
 _    _            _     _ _____ _          _ _  ______ _           _           
| |  | |          | |   | /  ___| |        | | | |  ___(_)         | |          
| |  | | ___  _ __| | __| \ ` + "`" + `--.| |__   ___| | | | |_   _ _ __   __| | ___ _ __ 
| |/\| |/ _ \| '__| |/ _` + "`" + ` |` + "`" + `--. \ '_ \ / _ \ | | |  _| | | '_ \ / _` + "`" + ` |/ _ \ '__|
\  /\  / (_) | |  | | (_| /\__/ / | | |  __/ | | | |   | | | | | (_| |  __/ |   
 \/  \/ \___/|_|  |_|\__,_\____/|_| |_|\___|_|_| \_|   |_|_| |_|\__,_|\___|_|   
                                                                                
`

func loadKeywords(wordlistPath string) ([]string, error) {
    var keywords []string

    if wordlistPath == "" {
        keywords = strings.Split(defaultWordlist, "\n")
    } else {
        file, err := os.Open(wordlistPath)
        if err != nil {
            return nil, err
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            keyword := scanner.Text()
            if keyword != "" {
                keywords = append(keywords, keyword)
            }
        }

        if err := scanner.Err(); err != nil {
            return nil, err
        }
    }

    return keywords, nil
}

func scanFiles(directory string, keywords []string) {
    err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Printf("Error accessing file: %v\n", err)
            return err
        }
        if !info.IsDir() {
            match, err := containsKeywords(path, keywords)
            if err != nil {
                log.Printf("Error reading file: %v\n", err)
                return err
            }
            if match {
                fmt.Println("Potential webshell found in:", path)
            }
        }
        return nil
    })
    if err != nil {
        log.Fatalf("Error scanning files: %v\n", err)
    }
}

func containsKeywords(filename string, keywords []string) (bool, error) {
    file, err := os.Open(filename)
    if err != nil {
        return false, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        for _, keyword := range keywords {
            if strings.Contains(line, keyword) {
                return true, nil
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return false, err
    }
    return false, nil
}

func main() {
    fmt.Println(banner)

    if len(os.Args) < 2 {
        fmt.Println("Usage: worldfind <directory> [optional wordlist]")
        return
    }

    directory := os.Args[1]
    var wordlistPath string
    if len(os.Args) > 2 {
        wordlistPath = os.Args[2]
    }

    keywords, err := loadKeywords(wordlistPath)
    if err != nil {
        log.Fatalf("Failed to load keywords: %v\n", err)
    }

    scanFiles(directory, keywords)
}
