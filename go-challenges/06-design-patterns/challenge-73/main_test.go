package main

import "testing"

func TestCommand(t *testing.T) {
	light := &Light{}
	remote := &RemoteControl{}
	
	onCmd := &LightOnCommand{light: light}
	remote.SetCommand(onCmd)
	remote.PressButton()
	
	if !light.IsOn {
		t.Error("Light should be on")
	}
	
	offCmd := &LightOffCommand{light: light}
	remote.SetCommand(offCmd)
	remote.PressButton()
	
	if light.IsOn {
		t.Error("Light should be off")
	}
	
	t.Log("✓ Command pattern works!")
}
