package main

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Embed the default wordlist file
//
//go:embed wordlists/default.txt
var defaultWordlist embed.FS

const banner = `
 _    _            _     _ _____ _          _ _  ______ _           _           
| |  | |          | |   | /  ___| |        | | | |  ___(_)         | |          
| |  | | ___  _ __| | __| \ ` + "`" + `--.| |__   ___| | | | |_   _ _ __   __| | ___ _ __ 
| |/\| |/ _ \| '__| |/ _` + "`" + ` |` + "`" + `--. \ '_ \ / _ \ | | |  _| | | '_ \ / _` + "`" + ` |/ _ \ '__|
\  /\  / (_) | |  | | (_| /\__/ / | | |  __/ | | | |   | | | | | (_| |  __/ |   
 \/  \/ \___/|_|  |_|\__,_\____/|_| |_|\___|_|_| \_|   |_|_| |_|\__,_|\___|_|   
`

// Loading animation while scanning
func loadingAnimation(done chan bool) {
	chars := []rune{'|', '/', '-', '\\'}
	for {
		select {
		case <-done:
			fmt.Print("\rScan complete!                          \n")
			return
		default:
			for _, c := range chars {
				fmt.Printf("\rScanning files... %c", c)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func loadKeywords(wordlistPath string) ([]string, error) {
	var keywords []string

	if wordlistPath == "" {
		// Load default embedded wordlist if no path is provided
		file, err := defaultWordlist.Open("wordlists/default.txt")
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

	// Load from external wordlist if a path is provided
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
			// Skip any directories or files that cause an error (e.g., symbolic links)
			log.Printf("Error accessing file or directory: %v\n", err)
			return nil // Continue scanning the next file
		}

		if !info.IsDir() {
			match, err := containsKeywords(path, keywords, regexes)
			if err != nil {
				log.Printf("Error reading file: %v\n", err)
				return nil // Continue scanning even if one file fails
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
	buf := make([]byte, 1024*1024)    // 1MB buffer
	scanner.Buffer(buf, 20*1024*1024) // Set max token size to 20MB

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
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("fail to get directory: %w", err)
	}

	// Run "go get -u" to update from repository
	cmd := exec.Command("go", "get", "-u", repoURL)
	cmd.Dir = currentDir // Run command in the current directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update from repository: %w\noutput: %s", err, output)
	}

	fmt.Println("Update Success!")
	return nil
}

func printHelp() {
	fmt.Println("Usage: worldfind [option] <directory> [wordlist]")
	fmt.Println("Option:")
	fmt.Println("  --update     Update latest version from repository.")
	fmt.Println("  -h, --help   Display this help.")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// Handle options
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--update":
			repoURL := "github.com/Arya-f4/worldshellfinder" // Change with your repository URL!
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
			// Proceed with directory scanning
		}
	}

	// Get directory and wordlist path from arguments
	var directory string
	var wordlistPath string

	if len(os.Args) > 1 {
		directory = os.Args[1]
	}
	if len(os.Args) > 2 {
		wordlistPath = os.Args[2]
	}

	// If directory is not provided, show help
	if directory == "" {
		printHelp()
		return
	}

	// Load keywords
	keywords, err := loadKeywords(wordlistPath)
	if err != nil {
		log.Fatalf("Fail to load Keyword: %v\n", err)
	}

	// Refined regex patterns for more specific webshell detection
	regexPatterns := []string{
		`(?i)(eval|assert|system|shell_exec|passthru)\s*\(\s*["']?[a-zA-Z0-9+/=]{20,}["']?\s*\)`,                     // Obfuscated eval with base64-like strings
		`(?i)(exec|system|popen|proc_open)\s*\(\s*\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[([^\]]+)\]\s*\)`,            // Remote command execution via superglobals
		`(?i)move_uploaded_file\s*\(.*?,\s*['"]\.\./(.*?)\.php['"]\s*\)`,                                             // File upload with .php
		`(?i)(file_put_contents|fwrite|fputs)\s*\(\s*['"](.*?\.php)['"],\s*(base64_decode|gzinflate|gzuncompress)\(`, // Write obfuscated PHP shell
		`(?i)(?s)\$_(?:GET|POST|REQUEST|COOKIE)\[.*?\]\s*\((eval|system|exec|shell_exec|passthru)\)`,                 // Superglobal with execution function
	}

	var regexes []*regexp.Regexp
	for _, pattern := range regexPatterns {
		rx, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatalf("Invalid regex pattern: %s\n", err)
		}
		regexes = append(regexes, rx)
	}

	// Start loading animation
	done := make(chan bool)
	go loadingAnimation(done)

	fmt.Print(banner)
	scanFiles(directory, keywords, regexes)

	// Stop the loading animation
	done <- true
}
