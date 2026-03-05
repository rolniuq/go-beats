package radio

import (
	"net/url"
	"strings"
	"testing"
)

func TestDefaultStations(t *testing.T) {
	stations := DefaultStations()

	if len(stations) == 0 {
		t.Fatal("DefaultStations() returned empty slice")
	}

	for _, station := range stations {
		t.Run(station.Name, func(t *testing.T) {
			if station.Name == "" {
				t.Error("Station name is empty")
			}

			if station.URL == "" {
				t.Error("Station URL is empty")
			}

			_, err := url.ParseRequestURI(station.URL)
			if err != nil {
				t.Errorf("Station URL is invalid: %v", err)
			}

			if !strings.HasPrefix(station.URL, "http://") && !strings.HasPrefix(station.URL, "https://") {
				t.Error("Station URL must start with http:// or https://")
			}

			if station.Genre == "" {
				t.Error("Station genre is empty")
			}

			if station.Description == "" {
				t.Error("Station description is empty")
			}
		})
	}
}

func TestStationCount(t *testing.T) {
	stations := DefaultStations()
	player := &Player{
		stations: stations,
	}

	if player.StationCount() != len(stations) {
		t.Errorf("StationCount() = %v, want %v", player.StationCount(), len(stations))
	}
}

func TestStationsReturnsCopy(t *testing.T) {
	stations := DefaultStations()
	player := &Player{
		stations: stations,
	}

	copy := player.Stations()

	if len(copy) != len(stations) {
		t.Errorf("Stations() length = %v, want %v", len(copy), len(stations))
	}

	copy[0].Name = "Modified"
	if stations[0].Name == "Modified" {
		t.Error("Stations() should return a copy, not the original")
	}
}

func TestStationStruct(t *testing.T) {
	station := Station{
		Name:        "Test Station",
		URL:         "https://example.com/stream",
		Genre:       "test",
		Description: "Test description",
	}

	if station.Name != "Test Station" {
		t.Errorf("Name = %v, want %v", station.Name, "Test Station")
	}
	if station.URL != "https://example.com/stream" {
		t.Errorf("URL = %v, want %v", station.URL, "https://example.com/stream")
	}
	if station.Genre != "test" {
		t.Errorf("Genre = %v, want %v", station.Genre, "test")
	}
	if station.Description != "Test description" {
		t.Errorf("Description = %v, want %v", station.Description, "Test description")
	}
}

func TestDefaultStationsIncludesLofiRadio24(t *testing.T) {
	stations := DefaultStations()

	for _, station := range stations {
		if station.Name == "LofiRadio24" {
			if station.URL != "https://stream.zeno.fm/0r0xa792kwzuv" {
				t.Errorf("LofiRadio24 URL = %q, want %q", station.URL, "https://stream.zeno.fm/0r0xa792kwzuv")
			}
			if station.Genre != "lofi hip-hop" {
				t.Errorf("LofiRadio24 genre = %q, want %q", station.Genre, "lofi hip-hop")
			}
			if station.Description != "24/7 lofi hip hop radio - relaxing beats to study and chill" {
				t.Errorf("LofiRadio24 description = %q, want %q", station.Description, "24/7 lofi hip hop radio - relaxing beats to study and chill")
			}
			return
		}
	}

	t.Fatal("LofiRadio24 station not found")
}
