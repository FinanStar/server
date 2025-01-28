package mailer

type Mailer interface {
	Send(to string, title string, message string) error
}
