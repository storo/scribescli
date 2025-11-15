package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/storo/scribescli/internal/storage"
	"github.com/storo/scribescli/internal/tui"
)

// Version can be overridden at build time with -ldflags
var Version = "1.0.0"

func main() {
	// Parse command-line flags
	versionFlag := flag.Bool("version", false, "Print version and exit")
	versionShortFlag := flag.Bool("v", false, "Print version and exit (short)")
	helpFlag := flag.Bool("help", false, "Print help message and exit")
	helpShortFlag := flag.Bool("h", false, "Print help message and exit (short)")
	debugFlag := flag.Bool("debug", false, "Enable debug logging to scribescli.log")

	flag.Parse()

	// Handle version flag
	if *versionFlag || *versionShortFlag {
		fmt.Printf("ScribesAI v%s\n", Version)
		os.Exit(0)
	}

	// Handle help flag
	if *helpFlag || *helpShortFlag {
		printHelp()
		os.Exit(0)
	}

	// Load .env file (ignore error if not found)
	if err := godotenv.Load(); err != nil {
		// Only log if the error is not "file not found"
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Check required environment variables
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY environment variable not set")
		fmt.Fprintln(os.Stderr, "Please create a .env file with your API key or set it in your environment")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Example .env file:")
		fmt.Fprintln(os.Stderr, "  ANTHROPIC_API_KEY=your-api-key-here")
		os.Exit(1)
	}

	// Setup debug logging if requested
	if *debugFlag {
		f, err := tea.LogToFile("scribescli.log", "debug")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up debug log: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		log.Println("Debug logging enabled")
	}

	// Create necessary directories
	dirs := []string{"data", "data/exports", "models"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/scribescli.db"
	}

	// Initialize database
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize TUI model
	m := tui.NewModel(db)

	// Create Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle signals in a goroutine
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down gracefully...")
		p.Quit()
	}()

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

// printHelp prints the help message
func printHelp() {
	fmt.Printf(`ScribesAI - Meeting Intelligence System

Usage:
  scribescli [flags]

Flags:
  -h, --help     Show this help message
  -v, --version  Show version information
  --debug        Enable debug logging to scribescli.log

Environment Variables:
  ANTHROPIC_API_KEY  Claude API key (required)
  VOSK_MODEL_PATH    Path to Vosk model (optional)
  DB_PATH            Database path (default: ./data/scribescli.db)

Examples:
  # Run with default settings
  scribescli

  # Run with debug logging
  scribescli --debug

  # Set API key via environment
  export ANTHROPIC_API_KEY=your-api-key-here
  scribescli

For more information, visit: https://github.com/storo/scribescli
`)
}
