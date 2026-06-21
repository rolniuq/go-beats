package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rolniuq/go-beats/internal/audio"
	"github.com/rolniuq/go-beats/internal/radio"
	"github.com/rolniuq/go-beats/internal/ui"
	"github.com/rolniuq/go-beats/internal/util"
)

// Set by GoReleaser ldflags at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// CLI flags
	radioModeFlag := flag.Bool("radio", false, "Start directly in radio mode")
	stationIdxFlag := flag.Int("station", -1, "Auto-play station index (implies --radio)")
	listStationsFlag := flag.Bool("list-stations", false, "List available radio stations and exit")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("go-beats %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	radioPlayer := radio.NewPlayer()

	if *listStationsFlag {
		for i, station := range radioPlayer.Stations() {
			fmt.Printf("%d: %s [%s]\n   %s\n   %s\n", i, station.Name, station.Genre, station.Description, station.URL)
		}
		return
	}

	startInRadio := *radioModeFlag
	autoStation := *stationIdxFlag
	if autoStation >= 0 {
		if autoStation >= radioPlayer.StationCount() {
			fmt.Fprintf(os.Stderr, "Error: station index %d out of range (0-%d)\n", autoStation, radioPlayer.StationCount()-1)
			os.Exit(1)
		}
		startInRadio = true
	}

	// Determine music directory
	musicDir := "./music"
	if flag.NArg() > 0 {
		musicDir = flag.Arg(0)
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

	if info, statErr := os.Stat(absDir); statErr != nil {
		if os.IsNotExist(statErr) {
			if !startInRadio {
				fmt.Fprintf(os.Stderr, "Music directory not found: %s\n", absDir)
				fmt.Println("No local library found, defaulting to radio mode. Use --list-stations to browse stations.")
				startInRadio = true
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error reading music directory: %v\n", statErr)
			os.Exit(1)
		}
	} else if !info.IsDir() {
		if !startInRadio {
			fmt.Fprintf(os.Stderr, "Music path is not a directory: %s\n", absDir)
			os.Exit(1)
		}
	} else {
		if err := engine.ScanDirectory(absDir); err != nil {
			if !startInRadio {
				if !util.HasMP3Files(absDir) {
					fmt.Printf("No local tracks found in %s\n", absDir)
					fmt.Println("Starting in radio mode. Use --list-stations to browse stations.")
					startInRadio = true
				} else {
					fmt.Fprintf(os.Stderr, "Error scanning music: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			fmt.Printf("📂 Found %d tracks in %s\n", engine.TrackCount(), absDir)
		}
	}

	// Create and run TUI
	model := ui.NewModel(engine, radioPlayer)
	if startInRadio {
		model.SetMode(ui.ModeRadio)
	}

	if autoStation >= 0 {
		if err := radioPlayer.Play(autoStation); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting station %d: %v\n", autoStation, err)
			os.Exit(1)
		}
	}
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	engine.Stop()
	radioPlayer.Stop()
}
