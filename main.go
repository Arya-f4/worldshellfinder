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

type FileModification struct {
	path           string
	originalSize   int64
	modifiedSize   int64
	stringsRemoved int
}

//go:embed wordlists/default.txt
var defaultWordlist embed.FS
var shellCount int
var verbose bool

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
 made with love by ` + Yellow + ` Worldsavior/Arya-f4 ` + Magenta + `^^	 ` + Green + `	v.1.2.3.0 Stable Build  ` + Reset + `
===========================================================================================
`

const menuText = `
Please choose an option:
1. Normal WebShell Detection
2. Remove String from Files
`

func loadingAnimation(done chan bool) {
	chars := []rune{'|', '/', '-', '\\'}
	for {
		select {
		case <-done:
			fmt.Print("\rOperation complete!                          \n")
			return
		default:
			for _, c := range chars {
				fmt.Printf("\rProcessing... %c", c)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func removeStringFromFile(filePath string, stringToRemove string) (*FileModification, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	originalSize := int64(len(content))
	originalContent := string(content)

	// Count how many instances were removed
	stringsRemoved := strings.Count(originalContent, stringToRemove)

	// Only proceed if we found matches
	if stringsRemoved == 0 {
		return nil, nil
	}

	// Convert to string and remove the specified string
	newContent := strings.ReplaceAll(originalContent, stringToRemove, "")

	// Write back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return nil, err
	}

	// Return modification info
	return &FileModification{
		path:           filePath,
		originalSize:   originalSize,
		modifiedSize:   int64(len(newContent)),
		stringsRemoved: stringsRemoved,
	}, nil
}
func removeStringFromDirectory(directory string, stringToRemove string) error {
	var modifications []*FileModification
	var totalFilesScanned int
	var totalFilesModified int
	var totalStringsRemoved int

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing file or directory: %v\n", err)
			return nil
		}

		if !info.IsDir() {
			totalFilesScanned++
			modification, err := removeStringFromFile(path, stringToRemove)
			if err != nil {
				log.Printf("Error processing file %s: %v\n", path, err)
				return nil
			}

			if modification != nil {
				modifications = append(modifications, modification)
				totalFilesModified++
				totalStringsRemoved += modification.stringsRemoved
			}

			if verbose {
				fmt.Printf("Processed file: %s\n", path)
			}
		}
		return nil
	})

	// Print summary after processing
	fmt.Printf("\n%sString Removal Summary:%s\n", Yellow, Reset)
	fmt.Printf("Total files scanned: %d\n", totalFilesScanned)
	fmt.Printf("Total files modified: %d\n", totalFilesModified)
	fmt.Printf("Total strings removed: %d\n", totalStringsRemoved)

	if len(modifications) > 0 {
		fmt.Printf("\n%sModified Files:%s\n", Green, Reset)
		for _, mod := range modifications {
			fmt.Printf("- %s\n", mod.path)
			fmt.Printf("  Strings removed: %d\n", mod.stringsRemoved)
			fmt.Printf("  Size change: %d bytes -> %d bytes\n", mod.originalSize, mod.modifiedSize)
		}
	} else {
		fmt.Printf("\n%sNo files were modified.%s\n", Blue, Reset)
	}

	return err
}

// [Previous functions remain the same: loadKeywords, scanFiles, containsKeywords, updateFromRepository]

func printHelp() {
	fmt.Print("Usage: worldfind <directory>\n")
	fmt.Print("Options:\n")
	fmt.Print("1. Normal WebShell Detection where it's only detect webshells in files\n")
	fmt.Print("   Output Example : File: C:/directory/file.php\n")
	fmt.Print("  			Line number: 4\n")
	fmt.Print("			Detection type: Keyword Match\n")
	fmt.Print("			Matched keyword: gzinflate(base64_decode(")
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

type ShellDetection struct {
	path         string
	keyword      string
	lineNumber   int
	isRegexMatch bool
	matchedLine  string
}

func containsKeywords(filename string, keywords []string, regexes []*regexp.Regexp) (*ShellDetection, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)    // 1MB buffer
	scanner.Buffer(buf, 20*1024*1024) // Set max token size to 20MB

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check keywords
		for _, keyword := range keywords {
			if strings.Contains(line, keyword) {
				return &ShellDetection{
					path:         filename,
					keyword:      keyword,
					lineNumber:   lineNumber,
					isRegexMatch: false,
					matchedLine:  line,
				}, nil
			}
		}

		// Check regex patterns
		for _, rx := range regexes {
			if rx.MatchString(line) {
				return &ShellDetection{
					path:         filename,
					keyword:      "regex_pattern",
					lineNumber:   lineNumber,
					isRegexMatch: true,
					matchedLine:  line,
				}, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func scanFiles(directory string, keywords []string, regexes []*regexp.Regexp) {
	var detections []*ShellDetection
	var totalFilesScanned int

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing file or directory: %v\n", err)
			return nil
		}

		if !info.IsDir() {
			totalFilesScanned++

			if verbose {
				fmt.Printf("\rScanning: %s", path)
			}

			detection, err := containsKeywords(path, keywords, regexes)
			if err != nil {
				log.Printf("\nError reading file: %v\n", err)
				return nil
			}

			if detection != nil {
				detections = append(detections, detection)
				shellCount++
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("\nError scanning files: %v\n", err)
	}

	// Clear the scanning line if in verbose mode
	if verbose {
		fmt.Print("\r")
	}

	// Print summary
	fmt.Printf("\n%sWebShell Detection Summary:%s\n", Yellow, Reset)
	fmt.Printf("Total files scanned: %d\n", totalFilesScanned)
	fmt.Printf("Total potential webshells found: %d\n", shellCount)

	if len(detections) > 0 {
		fmt.Printf("\n%sPotential WebShells Found:%s\n", Red, Reset)
		for _, detect := range detections {
			fmt.Printf("\n- File: %s\n", detect.path)
			fmt.Printf("  Line number: %d\n", detect.lineNumber)
			if detect.isRegexMatch {
				fmt.Printf("  Detection type: %sRegex Pattern Match%s\n", Magenta, Reset)
			} else {
				fmt.Printf("  Detection type: %sKeyword Match%s\n", Blue, Reset)
				fmt.Printf("  Matched keyword: %s\n", detect.keyword)
			}

			if verbose {
				// Show the matched line with some context
				fmt.Printf("  Matched line: %s\n", strings.TrimSpace(detect.matchedLine))
			}
		}
	} else {
		fmt.Printf("\n%sNo potential webshells were found.%s\n", Green, Reset)
	}
}
func writeDetectionsToFile(filepath string, detections []*ShellDetection, totalFilesScanned int) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	fmt.Fprintf(writer, "WebShell Detection Report\n")
	fmt.Fprintf(writer, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "Total files scanned: %d\n", totalFilesScanned)
	fmt.Fprintf(writer, "Total potential webshells found: %d\n\n", len(detections))

	// Write each detection
	for i, detect := range detections {
		fmt.Fprintf(writer, "Detection #%d:\n", i+1)
		fmt.Fprintf(writer, "- File: %s\n", detect.path)
		fmt.Fprintf(writer, "- Line number: %d\n", detect.lineNumber)
		if detect.isRegexMatch {
			fmt.Fprintf(writer, "- Detection type: Regex Pattern Match\n")
		} else {
			fmt.Fprintf(writer, "- Detection type: Keyword Match\n")
			fmt.Fprintf(writer, "- Matched keyword: %s\n", detect.keyword)
		}
		fmt.Fprintf(writer, "- Matched line: %s\n\n", strings.TrimSpace(detect.matchedLine))
	}

	return writer.Flush()
}

func writeModificationsToFile(filepath string, modifications []*FileModification, totalFilesScanned int, totalStringsRemoved int) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	fmt.Fprintf(writer, "String Removal Report\n")
	fmt.Fprintf(writer, "Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "Total files scanned: %d\n", totalFilesScanned)
	fmt.Fprintf(writer, "Total files modified: %d\n", len(modifications))
	fmt.Fprintf(writer, "Total strings removed: %d\n\n", totalStringsRemoved)

	// Write each modification
	for i, mod := range modifications {
		fmt.Fprintf(writer, "Modification #%d:\n", i+1)
		fmt.Fprintf(writer, "- File: %s\n", mod.path)
		fmt.Fprintf(writer, "- Strings removed: %d\n", mod.stringsRemoved)
		fmt.Fprintf(writer, "- Original size: %d bytes\n", mod.originalSize)
		fmt.Fprintf(writer, "- Modified size: %d bytes\n", mod.modifiedSize)
		fmt.Fprintf(writer, "- Size difference: %d bytes\n\n", mod.originalSize-mod.modifiedSize)
	}

	return writer.Flush()
}

func main() {
	// Define flags
	helpFlag := flag.Bool("h", false, "display help information")
	helpFlagLong := flag.Bool("help", false, "display help information")

	// Parse flags
	flag.Parse()

	// Print banner first
	fmt.Print(banner)

	// Check for help flags
	if *helpFlag || *helpFlagLong {
		printHelp()
		return
	}

	args := flag.Args()

	// Handle update command
	if len(args) > 0 && args[0] == "--update" {
		repoURL := "github.com/Arya-f4/worldshellfinder"
		err := updateFromRepository(repoURL)
		if err != nil {
			log.Fatalf("Error While Updating: %v\n", err)
		}
		fmt.Println("Update done.")
		return
	}

	// Show menu if no specific command-line options
	fmt.Print(menuText)
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your choice (1 or 2): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	// Get directory path
	fmt.Print("Enter the directory to scan: ")
	directory, _ := reader.ReadString('\n')
	directory = strings.TrimSpace(directory)

	// Get output file path
	fmt.Print("Enter the output file path (press Enter for no file output): ")
	outputFile, _ := reader.ReadString('\n')
	outputFile = strings.TrimSpace(outputFile)

	switch choice {
	case "1":
		// Normal WebShell Detection
		keywords, err := loadKeywords("")
		if err != nil {
			log.Fatalf("Fail to load Keyword: %v\n", err)
		}

		regexPatterns := []string{
			`(?i)(eval|assert|system|shell_exec|passthru)\s*\(\s*["']?[a-zA-Z0-9+/=]{20,}["']?\s*\)`,
			`(?i)(exec|system|popen|proc_open)\s*\(\s*\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[([^\]]+)\]\s*\)`,
			`(?i)move_uploaded_file\s*\(.*?,\s*['"]\.\./(.*?)\.php['"]\s*\)`,
			`(?i)(passthru|shell_exec|system|exec)\s*\(\s*\$_(?:GET|POST|REQUEST|COOKIE|SERVER)\[.*?\]\s*\)`,
			`eval\(\s*\$\w+\s*\(\s*\$\w+\s*\(\s*\$\w+\s*\(\s*\$\w+\s*\(\s*\$\w+\s*\)\s*\)\s*\)\s*\)\s*\)\s*;`,
			`(?i)is_dir\s*\(\s*\$\w+\s*\)\s*\?\s*rmdir\s*\(\s*\$\w+\s*\)\s*:\s*unlink\s*\(\s*\$\w+\s*\)\s*;`,
			`(?i)if\s*\(\s*is_dir\s*\(\s*\$\w+\s*\)\s*&&\s*is_readable\s*\(\s*\$\w+\s*\)\s*&&\s*!\s*is_link\s*\(\s*\$\w+\s*\)\s*\)\s*{`,
			`(?i)__halt_compiler\s*\(\s*\)\s*`,
			`(?:isset\s*\(\s*\$_SERVER\s*\[\s*['"]H(?:['"]\s*\.\s*['"])?T(?:['"]\s*\.\s*['"])?T(?:['"]\s*\.\s*['"])?P(?:['"]\s*\.\s*['"])?S['"]\]\s*\)|\$_SERVER\s*\[\s*['"]H(?:['"]\s*\.\s*['"])?T(?:['"]\s*\.\s*['"])?T(?:['"]\s*\.\s*['"])?P(?:['"]\s*\.\s*['"])?S['"]\]\s*===\s*(?:['"]o(?:['"]\s*\.\s*['"])?n['"]|['"]on['"]))\s*(?:\?\s*['"]ht(?:['"]\s*\.\s*['"])?tp(?:['"]\s*\.\s*['"])?s['"]\s*:\s*['"]ht(?:['"]\s*\.\s*['"])?tp['"]|\)\s*\.\s*['"]://['"]\s*\.\s*\$_SERVER\s*\[\s*['"]HT(?:['"]\s*\.\s*['"])?T(?:['"]\s*\.\s*['"])?P(?:['"]\s*\.\s*['"])?_H(?:['"]\s*\.\s*['"])?OS(?:['"]\s*\.\s*['"])?T['"])`,
		}

		var regexes []*regexp.Regexp
		for _, pattern := range regexPatterns {
			rx, err := regexp.Compile(pattern)
			if err != nil {
				log.Fatalf("Failed to compile regex: %v\n", err)
			}
			regexes = append(regexes, rx)
		}

		var detections []*ShellDetection
		var totalFilesScanned int

		done := make(chan bool)
		go loadingAnimation(done)

		// Walk through directory and collect detections
		err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Error accessing file or directory: %v\n", err)
				return nil
			}

			if !info.IsDir() {
				totalFilesScanned++

				if verbose {
					fmt.Printf("\rScanning: %s", path)
				}

				detection, err := containsKeywords(path, keywords, regexes)
				if err != nil {
					log.Printf("\nError reading file: %v\n", err)
					return nil
				}

				if detection != nil {
					detections = append(detections, detection)
					shellCount++
				}
			}
			return nil
		})

		done <- true

		// Print results to console
		fmt.Printf("\n%sWebShell Detection Summary:%s\n", Yellow, Reset)
		fmt.Printf("Total files scanned: %d\n", totalFilesScanned)
		fmt.Printf("Total potential webshells found: %d\n", len(detections))

		if len(detections) > 0 {
			fmt.Printf("\n%sPotential WebShells Found:%s\n", Red, Reset)
			for _, detect := range detections {
				fmt.Printf("\n- File: %s\n", detect.path)
				fmt.Printf("  Line number: %d\n", detect.lineNumber)
				if detect.isRegexMatch {
					fmt.Printf("  Detection type: %sRegex Pattern Match%s\n", Magenta, Reset)
				} else {
					fmt.Printf("  Detection type: %sKeyword Match%s\n", Blue, Reset)
					fmt.Printf("  Matched keyword: %s\n", detect.keyword)
				}
				if verbose {
					fmt.Printf("  Matched line: %s\n", strings.TrimSpace(detect.matchedLine))
				}
			}
		} else {
			fmt.Printf("\n%sNo potential webshells were found.%s\n", Green, Reset)
		}

		// Write to file if specified
		if outputFile != "" {
			err = writeDetectionsToFile(outputFile, detections, totalFilesScanned)
			if err != nil {
				log.Printf("Error writing to output file: %v\n", err)
			} else {
				fmt.Printf("\nResults have been saved to: %s\n", outputFile)
			}
		}

	case "2":
		fmt.Print("Enter string to remove (press Ctrl+D or Ctrl+Z when done): ")

		// Create a buffer with 10MB capacity
		largeBuffer := make([]byte, 10*1024*1024) // 10MB buffer
		var totalSize int64
		maxSize := int64(10 * 1024 * 1024) // 10MB limit

		// Create a buffer to store the complete input
		var builder strings.Builder
		builder.Grow(len(largeBuffer)) // Pre-allocate capacity

		// Read input in chunks
		for {
			n, err := reader.Read(largeBuffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error reading input: %v\n", err)
			}

			// Check total size
			totalSize += int64(n)
			if totalSize > maxSize {
				log.Fatalf("Input exceeds maximum size of 10MB\n")
			}

			// Write chunk to builder
			builder.Write(largeBuffer[:n])
		}

		stringToRemove := strings.TrimSpace(builder.String())

		if stringToRemove == "" {
			fmt.Println("Error: Empty string provided")
			return
		}

		// Print size of string being removed
		fmt.Printf("String size to remove: %.2f MB\n", float64(len(stringToRemove))/(1024*1024))

		var modifications []*FileModification
		var totalFilesScanned int
		var totalStringsRemoved int

		done := make(chan bool)
		go loadingAnimation(done)

		// Walk through directory
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Error accessing file or directory: %v\n", err)
				return nil
			}

			if !info.IsDir() {
				totalFilesScanned++
				modification, err := removeStringFromFile(path, stringToRemove)
				if err != nil {
					log.Printf("Error processing file %s: %v\n", path, err)
					return nil
				}

				if modification != nil {
					modifications = append(modifications, modification)
					totalStringsRemoved += modification.stringsRemoved
				}

				if verbose {
					fmt.Printf("Processed file: %s\n", path)
				}
			}
			return nil
		})

		done <- true

		if err != nil {
			log.Fatalf("Error removing string: %v\n", err)
		}

		// Print results to console
		fmt.Printf("\n%sString Removal Summary:%s\n", Yellow, Reset)
		fmt.Printf("Total files scanned: %d\n", totalFilesScanned)
		fmt.Printf("Total files modified: %d\n", len(modifications))
		fmt.Printf("Total strings removed: %d\n", totalStringsRemoved)

		// Write to file if specified
		if outputFile != "" {
			err = writeModificationsToFile(outputFile, modifications, totalFilesScanned, totalStringsRemoved)
			if err != nil {
				log.Printf("Error writing to output file: %v\n", err)
			} else {
				fmt.Printf("\nResults have been saved to: %s\n", outputFile)
			}
		}

	default:
		fmt.Println("Invalid choice!")
		return
	}
}
