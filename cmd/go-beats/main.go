package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

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

func runMCPServer(player *radio.Player) {
	s := server.NewMCPServer(
		"go-beats",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	listStationsTool := mcp.NewTool("list_stations",
		mcp.WithDescription("List all available lofi/instrumental radio stations with their index, name, and genre"),
	)
	s.AddTool(listStationsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		stations := player.Stations()
		var sb strings.Builder
		for i, st := range stations {
			sb.WriteString(fmt.Sprintf("%d. %s [%s]\n   %s\n", i, st.Name, st.Genre, st.Description))
		}
		return mcp.NewToolResultText(sb.String()), nil
	})

	playStationTool := mcp.NewTool("play_station",
		mcp.WithDescription("Play a radio station by its index number (0-10)"),
		mcp.WithNumber("index",
			mcp.Required(),
			mcp.Description("Index of the station to play (0-10)"),
		),
	)
	s.AddTool(playStationTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		indexFloat, err := req.RequireFloat("index")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid index argument"), nil
		}
		index := int(indexFloat)
		if index < 0 || index >= player.StationCount() {
			return mcp.NewToolResultError(fmt.Sprintf("station index %d out of range (0-%d)", index, player.StationCount()-1)), nil
		}
		if err := player.Play(index); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to play station: %v", err)), nil
		}
		st := player.CurrentStation()
		return mcp.NewToolResultText(fmt.Sprintf("Now playing: %s [%s]", st.Name, st.Genre)), nil
	})

	stopTool := mcp.NewTool("stop",
		mcp.WithDescription("Stop playback of the current station"),
	)
	s.AddTool(stopTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		player.Stop()
		return mcp.NewToolResultText("Playback stopped"), nil
	})

	pauseTool := mcp.NewTool("pause",
		mcp.WithDescription("Pause or resume the current station"),
	)
	s.AddTool(pauseTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		player.Pause()
		if player.IsPaused() {
			return mcp.NewToolResultText("Paused"), nil
		}
		return mcp.NewToolResultText("Resumed"), nil
	})

	setVolumeTool := mcp.NewTool("set_volume",
		mcp.WithDescription("Set the volume level (0-100)"),
		mcp.WithNumber("level",
			mcp.Required(),
			mcp.Description("Volume level from 0 (mute) to 100 (max)"),
		),
	)
	s.AddTool(setVolumeTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		levelFloat, err := req.RequireFloat("level")
		if err != nil {
			return mcp.NewToolResultError("missing or invalid level argument"), nil
		}
		level := int(levelFloat)
		if level < 0 || level > 100 {
			return mcp.NewToolResultError("volume must be between 0 and 100"), nil
		}
		volumeRaised := level - player.GetVolumePercent()
		if volumeRaised > 0 {
			for i := 0; i < volumeRaised/5+1; i++ {
				if player.GetVolumePercent() >= level {
					break
				}
				player.VolumeUp()
			}
		} else {
			for i := 0; i < (-volumeRaised)/5+1; i++ {
				if player.GetVolumePercent() <= level {
					break
				}
				player.VolumeDown()
			}
		}
		return mcp.NewToolResultText(fmt.Sprintf("Volume set to %d%%", player.GetVolumePercent())), nil
	})

	getStatusTool := mcp.NewTool("get_status",
		mcp.WithDescription("Get current playback status: which station, volume, playing/paused"),
	)
	s.AddTool(getStatusTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var status string
		if player.IsPlaying() {
			st := player.CurrentStation()
			status = fmt.Sprintf("Playing: %s [%s] at volume %d%%", st.Name, st.Genre, player.GetVolumePercent())
		} else if player.IsPaused() {
			st := player.CurrentStation()
			status = fmt.Sprintf("Paused: %s [%s] at volume %d%%", st.Name, st.Genre, player.GetVolumePercent())
		} else {
			status = fmt.Sprintf("Stopped. Volume at %d%%", player.GetVolumePercent())
		}
		return mcp.NewToolResultText(status), nil
	})

	nextStationTool := mcp.NewTool("next_station",
		mcp.WithDescription("Skip to the next radio station"),
	)
	s.AddTool(nextStationTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := player.NextStation(); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to switch station: %v", err)), nil
		}
		st := player.CurrentStation()
		return mcp.NewToolResultText(fmt.Sprintf("Now playing: %s [%s]", st.Name, st.Genre)), nil
	})

	prevStationTool := mcp.NewTool("prev_station",
		mcp.WithDescription("Go back to the previous radio station"),
	)
	s.AddTool(prevStationTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := player.PrevStation(); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to switch station: %v", err)), nil
		}
		st := player.CurrentStation()
		return mcp.NewToolResultText(fmt.Sprintf("Now playing: %s [%s]", st.Name, st.Genre)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// CLI flags
	radioModeFlag := flag.Bool("radio", false, "Start directly in radio mode")
	stationIdxFlag := flag.Int("station", -1, "Auto-play station index (implies --radio)")
	listStationsFlag := flag.Bool("list-stations", false, "List available radio stations and exit")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	mcpFlag := flag.Bool("mcp", false, "Run as MCP server over stdio")
	flag.Parse()

	if !*mcpFlag {
		fmt.Println("🎵 Starting go-beats...")
	}

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

	if *mcpFlag {
		runMCPServer(radioPlayer)
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
