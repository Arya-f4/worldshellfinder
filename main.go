
package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "strings"
)

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

    // Use default wordlist if no path is provided
    if wordlistPath == "" {
        wordlistPath = "wordlists/default.txt"
    }

    file, err := os.Open(wordlistPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        keyword := strings.TrimSpace(scanner.Text())
        if keyword != "" {
            keywords = append(keywords, keyword)
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return keywords, nil
}

func scanFiles(directory string, keywords []string, regexes []*regexp.Regexp) {
    err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Printf("Error accessing file: %v\n", err)
            return err
        }
        if !info.IsDir() {
            match, err := containsKeywords(path, keywords, regexes)
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

func containsKeywords(filename string, keywords []string, regexes []*regexp.Regexp) (bool, error) {
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
        for _, rx := range regexes {
            if rx.MatchString(line) {
                return true, nil
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return false, err
    }
    return false, nil
}

func updateFromRepository(repoURL string) error {
    // Mendapatkan path direktori saat ini 
    currentDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("Fail to get directory: %w", err)
    }

    // Menjalankan perintah "go get -u" untuk mengupdate dari repository
    cmd := exec.Command("go", "get", "-u", repoURL)
    cmd.Dir = currentDir // Menjalankan perintah di direktori saat ini

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("Failed to update from repository: %w\nOutput: %s", err, output)
    }

    fmt.Println("Update Success!")
    return nil
}

func printHelp() {
    fmt.Println("Usage: worldfind [option] <directory> [wordlist]")
    fmt.Println("Option:")
    fmt.Println("  --update     Update latest version from repository.")
    fmt.Println("  -h, --help    Display this help. ")
}


func main() {
    if len(os.Args) < 2 {
        printHelp()
        return
    }

    // Menangani opsi
    for i := 1; i < len(os.Args); i++ {
        switch os.Args[i] {
        case "--update":
            repoURL := "github.com/Arya-f4/worldshellfinder" // Ganti dengan URL repository Anda!
            err := updateFromRepository(repoURL)
            if err != nil {
                log.Fatalf("Error While Updating: %v\n", err)
            }
            fmt.Println("Update done.")
            return
        case "-h", "--help":
            printHelp()
            return
        default:
            // Jika bukan opsi, lanjutkan ke logika scan direktori
        }
    }

    // Mendapatkan direktori dan path wordlist dari argumen
    var directory string
    var wordlistPath string

    if len(os.Args) > 1 {
        directory = os.Args[1]
    }
    if len(os.Args) > 2 {
        wordlistPath = os.Args[2]
    }

    // Jika direktori tidak diberikan, tampilkan bantuan
    if directory == "" {
        printHelp()
        return
    }

    // Memuat keyword
    keywords, err := loadKeywords(wordlistPath)
    if err != nil {
        log.Fatalf("Fail to load Keyword: %v\n", err)
    }

    // Regex untuk mendeteksi potensi webshell
	regexPatterns := []string{
		`(?i)\b(eval|assert|preg_replace|create_function)\s*\(\s*(base64_decode|str_rot13|gzinflate|gzuncompress|shell_exec|exec|system|passthru)\s*\(.*?\)\s*\)`, // decode + eksekusi
		`(?i)\b(system|exec|shell_exec|passthru|popen|proc_open)\s*\(\s*$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[([^\]]+)\]\s*\)`, // command injection langsung
		`(?i)function\s+([a-z0-9_]+)\s*\(\s*\)\s*\{\s*if\s*\(\s*isset\s*\(\s*\$_(GET|POST|REQUEST)\['?(.*?)'?]\s*\)\s*\)\s*\{\s*(eval|assert|preg_replace|create_function)\s*\((.*?)\);`, // backdoor dengan fungsi dan isset
		`(?i)(?s)move_uploaded_file\s*\(.*?,\s*['"]\.\./(.*?)\.php['"]\s*\)`, // upload file dengan perpindahan direktori 
		`(?i)\b(file_put_contents|fwrite|fputs)\s*\(\s*['"](.*?\.\.\/.*?\.php)['"],\s*(.*?)\)`, // tulis file ke path yang mencurigakan
	}
	
    var regexes []*regexp.Regexp
    for _, pattern := range regexPatterns {
        rx, err := regexp.Compile(pattern)
        if err != nil {
            log.Fatalf("Regex Pattern Not Valid: %s\n", err)
        }
        regexes = append(regexes, rx)
    }

    fmt.Println(banner)
    scanFiles(directory, keywords, regexes)
}