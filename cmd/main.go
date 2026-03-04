package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quynh-vo/go-beats/internal/audio"
	"github.com/quynh-vo/go-beats/internal/ui"
)

func main() {
	fmt.Println("🎵 Starting go-beats...")

	// Determine music directory
	musicDir := "./music"
	if len(os.Args) > 1 {
		musicDir = os.Args[1]
	}

	// Resolve to absolute path
	absDir, err := filepath.Abs(musicDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Ensure music directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Music directory not found: %s\n", absDir)
		fmt.Println("Create a 'music' directory and add some .mp3 files!")
		fmt.Println("Usage: go-beats [music-directory]")
		os.Exit(1)
	}

	// Initialize audio engine
	engine := audio.NewEngine()

	if err := engine.InitSpeaker(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing audio: %v\n", err)
		os.Exit(1)
	}

	// Scan music directory
	if err := engine.ScanDirectory(absDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning music: %v\n", err)
		fmt.Printf("Add .mp3 files to: %s\n", absDir)
		os.Exit(1)
	}

	fmt.Printf("📂 Found %d tracks in %s\n", engine.TrackCount(), absDir)

	// Create and run TUI
	model := ui.NewModel(engine)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	engine.Stop()
	fmt.Println("\n👋 Thanks for chilling with go-beats!")
}
