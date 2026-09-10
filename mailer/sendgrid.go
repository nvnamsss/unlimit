package mailer

import (
	"encoding/json"
	"fmt"

	"github.com/nvnamsss/unlimit/utility"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	contentTypeTextPlain = "text/plain"
	contentTypeTextHTML  = "text/html"
)

// SendgridMailer implements the Mailer interface using SendGrid.
type SendgridMailer struct {
	APIKey string
	// add other fields as needed, e.g., sender email
}

func NewSendgridMailer(apiKey string) Mailer {
	return &SendgridMailer{
		APIKey: apiKey,
	}
}

func (s *SendgridMailer) SendMail(req *SendMailRequest) error {
	m := mail.NewV3Mail()
	// Set the sender
	m.From = mail.NewEmail(req.From, req.From)

	// Set up receiver
	tos := make([]*mail.Email, 0, len(req.To))
	for _, v := range req.To {
		tos = append(tos, mail.NewEmail(v, v))
	}

	m.Personalizations = append(m.Personalizations, &mail.Personalization{
		To:                  tos,
		Subject:             req.Subject,
		DynamicTemplateData: req.Template.GetVariables(),
	})

	// set the content
	if req.Template.GetID() != "" {
		m.TemplateID = req.Template.GetID()
	} else {
		m.Content = []*mail.Content{
			mail.NewContent(contentTypeTextHTML, req.Body), // or "text/html" if the body is HTML
		}
	}

	// setup tracking settings
	tracking := &mail.TrackingSettings{}
	tracking.SetClickTracking(&mail.ClickTrackingSetting{
		Enable: utility.GetPointer(true),
		// EnableText: utils.GetPointer(true),
	}).SetOpenTracking(&mail.OpenTrackingSetting{
		Enable: utility.GetPointer(true),
	})
	m.TrackingSettings = tracking

	// Marshal the mail object to JSON
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// send the email
	request := sendgrid.GetRequest(s.APIKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = body
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	// handle response
	if response.StatusCode >= 400 {
		// log the response body for debugging
		return fmt.Errorf("SendgridMailer.SendMail: failed with status code %d", response.StatusCode)
	}

	return nil
}

func (s *SendgridMailer) Name() string {
	return "SendgridMailer"
}
