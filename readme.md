## Worldfind: A Simple Webshell Detection Tool

Worldfind is a basic command-line tool written in Go that helps you identify potential webshells hidden within your web server directories. It works by scanning files for suspicious keywords and regular expressions commonly found in malicious scripts.

**Disclaimer:** This tool is intended for educational and informational purposes only. It is not a substitute for comprehensive security measures. Use at your own risk. False positives are possible.

### Features:

- Scans files for specified keywords.
- Uses regular expressions to detect common webshell patterns.
- Customizable wordlist (optional).
- Simple and easy to use.

### Installation:

1. **Prerequisites:** Make sure you have Go installed on your system.
   - You can download and install it from [https://go.dev/dl/](https://go.dev/dl/).
2. **Download Worldfind:**
   - Clone the repository: `git clone https://github.com/yourusername/worldfind.git`
   - Or download the source code as a ZIP file and extract it.
3. **Build the Executable:**
   - Open a terminal and navigate to the worldfind directory.
   - Run the command: `go build`
   - This will create an executable file named `worldfind` in the same directory.

### Usage:

1. **Basic Scan:**
   ```bash
   ./worldfind <directory> 
   ```
   - Replace `<directory>` with the path to the directory you want to scan.

2. **Custom Wordlist:**
   ```bash
   ./worldfind <directory> <wordlist_path (optional)>
   ```
   - Replace `<wordlist_path>` with the path to your custom wordlist file.

**Wordlist Format:**

The wordlist should be a plain text file with one keyword per line. You can use the provided `wordlists/default.txt` file as a starting point.

**Example:**

```bash
./worldfind /var/www/html wordlists/my_wordlist.txt
```

This command will scan the `/var/www/html` directory using keywords from the `wordlists/my_wordlist.txt` file.

### Contributing:

Contributions are welcome! Please feel free to submit pull requests for new features, improvements, or bug fixes.

**Please note:** This tool is under development and may be updated in the future.

## Compatibility :
- Windows
- Linux
- Mac (Not Tested Yet)
