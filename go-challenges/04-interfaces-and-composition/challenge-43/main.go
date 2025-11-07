package main

type EmailSender interface {
	Send(to, subject, body string) error
}

type NotificationService struct {
	sender EmailSender
}

func NewNotificationService(sender EmailSender) *NotificationService {
	return &NotificationService{sender: sender}
}

func (n *NotificationService) NotifyUser(email, message string) error {
	return n.sender.Send(email, "Notification", message)
}

// Mock for testing
type MockEmailSender struct {
	Calls []EmailCall
}

type EmailCall struct {
	To      string
	Subject string
	Body    string
}

func (m *MockEmailSender) Send(to, subject, body string) error {
	m.Calls = append(m.Calls, EmailCall{To: to, Subject: subject, Body: body})
	return nil
}

func main() {}
