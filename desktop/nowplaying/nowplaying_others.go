//go:build !desktop

package nowplaying

import "time"

type noopController struct{}

func New() Controller {
	return &noopController{}
}

func (n *noopController) SetPlaying(title string, duration time.Duration)  {}
func (n *noopController) SetStation(name string)                           {}
func (n *noopController) SetPaused()                                       {}
func (n *noopController) SetStopped()                                      {}
func (n *noopController) SetProgress(position time.Duration)               {}
func (n *noopController) SetCommandHandler(handler func(Command, float64)) {}
