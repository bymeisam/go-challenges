package main

import "testing"

func TestAdapter(t *testing.T) {
	mp4 := MP4Player{}
	adapter := NewMediaAdapter(mp4)
	
	result := adapter.Play("video.mp4")
	if result != "Playing mp4: video.mp4" {
		t.Errorf("Adapter failed: %s", result)
	}
	
	t.Log("✓ Adapter pattern works!")
}
