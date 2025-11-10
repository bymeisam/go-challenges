package main

type Command interface {
	Execute() string
}

type Light struct {
	IsOn bool
}

type LightOnCommand struct {
	light *Light
}

func (c *LightOnCommand) Execute() string {
	c.light.IsOn = true
	return "Light is ON"
}

type LightOffCommand struct {
	light *Light
}

func (c *LightOffCommand) Execute() string {
	c.light.IsOn = false
	return "Light is OFF"
}

type RemoteControl struct {
	command Command
}

func (r *RemoteControl) SetCommand(command Command) {
	r.command = command
}

func (r *RemoteControl) PressButton() string {
	return r.command.Execute()
}

func main() {}
