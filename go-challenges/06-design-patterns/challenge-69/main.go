package main

// Target interface our code expects
type MediaPlayer interface {
	Play(filename string) string
}

// Existing incompatible interface
type AdvancedPlayer interface {
	PlayAdvanced(filename, format string) string
}

// Concrete implementation
type MP4Player struct{}

func (m MP4Player) PlayAdvanced(filename, format string) string {
	return "Playing " + format + ": " + filename
}

// Adapter
type MediaAdapter struct {
	advancedPlayer AdvancedPlayer
}

func NewMediaAdapter(player AdvancedPlayer) *MediaAdapter {
	return &MediaAdapter{advancedPlayer: player}
}

func (a *MediaAdapter) Play(filename string) string {
	return a.advancedPlayer.PlayAdvanced(filename, "mp4")
}

func main() {}
