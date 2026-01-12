package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cc-filter/internal/filter"
	"cc-filter/internal/logger"
)

// version is set at build time via -ldflags
var version = "dev"

func main() {
	// check for help or version flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			showHelp()
			return
		case "-v", "--version", "version":
			showVersion()
			return
		}
	}

	logger.Setup()

	start := time.Now()

	f, err := filter.New()
	if err != nil {
		log.Printf("Failed to initialize filter: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to initialize filter: %v\n", err)
		os.Exit(1)
	}

	input := readStdin()
	inputLength := len(input)

	result := f.Process(input)

	// Check for blocking error from hooks - EXIT CODE 2 BLOCKS THE PROMPT
	if result.Error != nil {
		fmt.Fprintln(os.Stderr, result.Error.Error())
		os.Exit(2) // Exit code 2 = blocks UserPromptSubmit, erases prompt
	}

	outputLength := len(result.Output)

	fmt.Print(result.Output)

	if result.Filtered {
		duration := time.Since(start)
		log.Printf("cc-filter applied filtering at %s - Input: %d bytes, Output: %d bytes, Duration: %v",
			start.Format(time.RFC3339), inputLength, outputLength, duration)
	}
}

func readStdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	var input strings.Builder
	
	for scanner.Scan() {
		input.WriteString(scanner.Text())
		input.WriteString("\n")
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	
	return strings.TrimSuffix(input.String(), "\n")
}

func showHelp() {
	fmt.Print(`cc-filter - Claude Code Sensitive Information Filter

USAGE:
    cc-filter [OPTIONS]

OPTIONS:
    -h, --help, help       Show this help message
    -v, --version, version Show version information

DESCRIPTION:
    cc-filter is a security tool that filters sensitive information from text input.
    It reads from stdin and outputs filtered text to stdout.

    The tool intercepts and filters:
    • API keys, secret keys, access tokens
    • Database URLs and connection strings
    • Private keys and certificates
    • Environment variables (KEY=value format)
    • OpenAI keys (sk-...), Slack tokens (xoxb-...)

EXAMPLES:
    # Filter sensitive data from text
    echo "API_KEY=secret123" | cc-filter

    # Filter a file
    cat config.txt | cc-filter

    # Use with Claude Code hooks (see README for setup)

CONFIGURATION:
    • Default rules: configs/default-rules.yaml
    • User config: ~/.cc-filter/config.yaml
    • Project config: config.yaml

    See README.md for configuration examples.

LOG FILE:
    ~/.cc-filter/filter.log

MORE INFO:
    https://github.com/wissem/cc-filter
`)
}

func showVersion() {
	fmt.Printf("cc-filter version %s\n", version)
}