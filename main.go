package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"

	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

// Embed the default wordlist file
//
//go:embed wordlists/default.txt
var defaultWordlist embed.FS
var shellCount int
var verbose bool // Global flag for verbose mode

const banner = `
` + Red + `
===========================================================================================
` + Cyan + `
 _    _            _     _ _____ _          _ _  ______ _           _           
| |  | |          | |   | /  ___| |        | | | |  ___(_)         | |          
| |  | | ___  _ __| | __| \ ` + "`" + `--.| |__   ___| | | | |_   _ _ __   __| | ___ _ __ 
| |/\| |/ _ \| '__| |/ _` + "`" + ` |` + "`" + `--. \ '_ \ / _ \ | | |  _| | | '_ \ / _` + "`" + ` |/ _ \ '__|
\  /\  / (_) | |  | | (_| /\__/ / | | |  __/ | | | |   | | | | | (_| |  __/ |   
 \/  \/ \___/|_|  |_|\__,_\____/|_| |_|\___|_|_| \_|   |_|_| |_|\__,_|\___|_|  
 ` + Reset + `	
 made with love by ` + Yellow + ` Worldsavior/Arya-f4 ` + Magenta + `^^	 ` + Green + `	v.1.1.1.3 Stable Build  ` + Reset + `
===========================================================================================
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

	// Load default embedded wordlist first
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

	// If an external wordlist is provided, append its keywords to the default list
	if wordlistPath != "" {
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
			match, keyword, err := containsKeywords(path, keywords, regexes)
			if err != nil {
				log.Printf("Error reading file: %v\n", err)
				return nil // Continue scanning even if one file fails
			}
			if match {
				fmt.Printf("\nPotential webshell found in: %s\n", path)
				if verbose && keyword != "" {
					fmt.Printf("Keyword detected: %s\n", keyword)
				}
				shellCount++
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error scanning files: %v\n", err)
	}
}

func containsKeywords(filename string, keywords []string, regexes []*regexp.Regexp) (bool, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)    // 1MB buffer
	scanner.Buffer(buf, 20*1024*1024) // Set max token size to 20MB

	for scanner.Scan() {
		line := scanner.Text()
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				return true, keyword, nil
			}
		}
		for _, rx := range regexes {
			if rx.MatchString(line) {
				return true, "regex_match", nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, "", err
	}
	return false, "", nil
}

func updateFromRepository(repoURL string) error {
	// Detect OS and architecture
	osType := runtime.GOOS
	archType := runtime.GOARCH

	// Construct the download URL based on the OS and architecture
	downloadURL := fmt.Sprintf("%s/releases/latest/download/%s_%s", repoURL, osType, archType)
	fmt.Printf("Downloading update from: %s\n", downloadURL)

	// Create a temp file to download the new binary
	tmpFile, err := os.CreateTemp("", "worldshellfinder_*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Download the new binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: HTTP %d", resp.StatusCode)
	}

	// Write the downloaded binary to the temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write update to file: %w", err)
	}

	// Get the path of the current running binary
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Replace the current binary with the new one
	err = os.Rename(tmpFile.Name(), executablePath)
	if err != nil {
		return fmt.Errorf("failed to replace current binary: %w", err)
	}

	// Make the new binary executable
	err = os.Chmod(executablePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to make new binary executable: %w", err)
	}

	fmt.Println("Update complete! Restarting the application...")

	// Restart the application
	cmd := exec.Command(executablePath)
	cmd.Start() // Start the new process
	os.Exit(0)  // Exit the current process

	return nil
}
func printHelp() {

	fmt.Println("Usage: worldfind [option] <directory> [wordlist]")
	fmt.Println("Option:")
	fmt.Println("  --update     Update latest version from repository.")
	fmt.Println("  -v           Enable verbose mode.")
	fmt.Println("  -h, --help   Display this help.")
}

func main() {
	// Define flags
	fmt.Print(banner)
	flagVerbose := flag.Bool("v", false, "enable verbose mode")
	flag.Parse()

	// Set verbose flag
	verbose = *flagVerbose

	// Get non-flag arguments
	args := flag.Args()
	if len(args) < 1 {
		printHelp()
		return
	}

	// Handle options
	for _, arg := range args {
		switch arg {
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

	if len(args) > 0 {
		directory = args[0]
	}
	if len(args) > 1 {
		wordlistPath = args[1]
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
		`(?i)(eval|assert|system|shell_exec|passthru)\s*\(\s*["']?[a-zA-Z0-9+/=]{20,}["']?\s*\)`,                    // Obfuscated eval with base64-like strings
		`(?i)(exec|system|popen|proc_open)\s*\(\s*\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[([^\]]+)\]\s*\)`,           // Remote command execution via superglobals
		`(?i)move_uploaded_file\s*\(.*?,\s*['"]\.\./(.*?)\.php['"]\s*\)`,                                            // File upload and renaming to PHP
		`(?i)(passthru|shell_exec|system|exec)\s*\(\s*\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[.*?\]\s*\)`,            // Command execution via superglobals
		`(?i)\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\s*\[\s*["']REMOTE_ADDR["']\s*\]`,                                 // Accessing superglobal arrays with user input
		`(?i)\$_FILES\s*\[\s*["'][^"']+["']\s*\]\s*\[\s*["']tmp_name["']\s*\]`,                                      // File upload with temp file
		`(?i)\$_FILES\s*\[\s*["'][^"']+["']\s*\]\s*\[\s*["']name["']\s*\]\s*\.\s*["']\.php["']`,                     // File upload with PHP extension
		`eval\(\s*\$\w+\s*\(\s*\$\w+\s*\(\s*\$\w+\s*\(\s*\$\w+\s*\(\s*\$\w+\s*\)\s*\)\s*\)\s*\)\s*\)\s*;`,           // Nested eval
		`(?i)\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[[^\]]+\]\s*=.*?\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[[^\]]+\]`, // Variable variable assignments
	}

	// Compile regexes
	var regexes []*regexp.Regexp
	for _, pattern := range regexPatterns {
		rx, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatalf("Failed to compile regex: %v\n", err)
		}
		regexes = append(regexes, rx)
	}

	// Start scanning with loading animation
	done := make(chan bool)
	go loadingAnimation(done)

	scanFiles(directory, keywords, regexes)

	// Stop loading animation after scanning

	done <- true
	// Print summary
	fmt.Printf("Number of potential webshells found: %d\n", shellCount)

}
