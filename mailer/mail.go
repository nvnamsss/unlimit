package mailer

type SendMailRequest struct {
	From     string
	To       []string
	Subject  string
	Body     string
	Template Template
}

type Mailer interface {
	SendMail(req *SendMailRequest) error
	Name() string
}
