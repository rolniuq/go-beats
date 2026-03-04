package radio

// Station represents an internet radio station
type Station struct {
	Name        string
	URL         string
	Genre       string
	Description string
}

// DefaultStations returns curated lofi/chill radio stations
// These are public internet radio streams (Icecast/Shoutcast/MP3 streams)
func DefaultStations() []Station {
	return []Station{
		{
			Name:        "Lofi Girl",
			URL:         "https://play.streamafrica.net/lofiradio",
			Genre:       "lofi hip-hop",
			Description: "lofi hip hop radio - beats to relax/study to",
		},
		{
			Name:        "Chillhop",
			URL:         "http://stream.zeno.fm/fyn8eh3h5f8uv",
			Genre:       "chillhop",
			Description: "Chillhop Music - jazzy & lofi hip hop beats",
		},
		{
			Name:        "Box Lofi",
			URL:         "http://stream.zeno.fm/f3wvbbqmdg8uv",
			Genre:       "lofi",
			Description: "Box Lofi - 24/7 lofi beats",
		},
		{
			Name:        "Nightride FM",
			URL:         "https://stream.nightride.fm/nightride.m4a",
			Genre:       "synthwave",
			Description: "Synthwave & retro electronic for late night coding",
		},
		{
			Name:        "Plaza One",
			URL:         "https://radio.plaza.one/mp3",
			Genre:       "vaporwave",
			Description: "Vaporwave & future funk radio",
		},
		{
			Name:        "SomaFM Groove Salad",
			URL:         "https://ice2.somafm.com/groovesalad-128-mp3",
			Genre:       "ambient/downtempo",
			Description: "A nicely chilled plate of ambient/downtempo beats",
		},
		{
			Name:        "SomaFM DEF CON",
			URL:         "https://ice2.somafm.com/defcon-128-mp3",
			Genre:       "electronic",
			Description: "Music for hacking - DEF CON radio",
		},
		{
			Name:        "SomaFM Drone Zone",
			URL:         "https://ice2.somafm.com/dronezone-128-mp3",
			Genre:       "ambient/drone",
			Description: "Served best chilled, safe with most medications",
		},
		{
			Name:        "SomaFM Deep Space One",
			URL:         "https://ice2.somafm.com/deepspaceone-128-mp3",
			Genre:       "space ambient",
			Description: "Deep ambient electronic, space music",
		},
		{
			Name:        "SomaFM Lush",
			URL:         "https://ice2.somafm.com/lush-128-mp3",
			Genre:       "electronic/female vocal",
			Description: "Sensuous and mellow vocals, mostly female",
		},
	}
}
