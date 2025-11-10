package main

type Observer interface {
	Update(message string)
}

type Subject struct {
	observers []Observer
}

func (s *Subject) Attach(observer Observer) {
	s.observers = append(s.observers, observer)
}

func (s *Subject) Notify(message string) {
	for _, observer := range s.observers {
		observer.Update(message)
	}
}

type EmailNotifier struct {
	Messages []string
}

func (e *EmailNotifier) Update(message string) {
	e.Messages = append(e.Messages, "Email: "+message)
}

type SMSNotifier struct {
	Messages []string
}

func (s *SMSNotifier) Update(message string) {
	s.Messages = append(s.Messages, "SMS: "+message)
}

func main() {}
