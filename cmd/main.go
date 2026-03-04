package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/quynh-vo/go-beats/internal/audio"
	"github.com/quynh-vo/go-beats/internal/radio"
	"github.com/quynh-vo/go-beats/internal/ui"
)

func main() {
	fmt.Println("🎵 Starting go-beats...")

	// Determine music directory
	musicDir := "./music"
	if len(os.Args) > 1 {
		musicDir = os.Args[1]
	}

	absDir, err := filepath.Abs(musicDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Initialize audio engine
	engine := audio.NewEngine()

	if err := engine.InitSpeaker(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing audio: %v\n", err)
		os.Exit(1)
	}

	// Initialize radio player
	radioPlayer := radio.NewPlayer()

	// Scan music directory
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Printf("Music directory not found: %s\n", absDir)
		fmt.Println("Note: You can use --radio flag to listen to internet radio!")
	} else {
		if err := engine.ScanDirectory(absDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning music: %v\n", err)
			fmt.Printf("Add .mp3 files to: %s\n", absDir)
		} else {
			fmt.Printf("📂 Found %d tracks in %s\n", engine.TrackCount(), absDir)
		}
	}

	// Run TUI
	model := ui.NewModel(engine, radioPlayer)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	engine.Stop()
	radioPlayer.Stop()
	fmt.Println("\n👋 Thanks for chilling with go-beats!")
}
