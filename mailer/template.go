package mailer

import (
	"bytes"
	"text/template"
)

type Template interface {
	GetID() string
	GetVariables() map[string]interface{}
	// Render the template and return subject
	RenderSubject() (string, error)
	// Render the template and return body
	RenderBody() (string, error)
}

type SendgridTemplate struct {
	id        string
	variables map[string]any
}

func NewSendgridTemplate(id string, variables map[string]any) Template {
	return &SendgridTemplate{id: id}
}

func (t *SendgridTemplate) GetID() string {
	return t.id
}

func (t *SendgridTemplate) GetVariables() map[string]interface{} {
	return nil
}

func (t *SendgridTemplate) RenderSubject() (string, error) {
	// intentionally left blank
	return "", nil
}

func (t *SendgridTemplate) RenderBody() (string, error) {
	// intentionally left blank
	return "", nil
}

// CustomizeTemplate represents a template implementation that allows for custom subject and body templates
// using Go's text/template package. It supports separate data contexts for subject and body rendering,
// enabling flexible email content generation with variable substitution.
// This is ideal for dynamic email content where templates are defined at runtime rather than pre-configured.
type CustomizeTemplate struct {
	id              string
	subjectTemplate string
	bodyTemplate    string // The raw template content
	subjectData     map[string]interface{}
	bodyData        map[string]interface{}
}

// NewCustomizeTemplate creates a new CustomizeTemplate instance with separate templates and data for subject and body.
//
// Parameters:
//
// - id: unique identifier for this template
//
// - subject: Go template string for the email subject line
//
// - body: Go template string for the email body content
//
// - subjectData: variables to be substituted in the subject template
//
// - bodyData: variables to be substituted in the body template
//
// Use this when you need runtime-defined templates with variable substitution, such as:
//
// - User-generated email templates
//
// - Dynamic content based on user preferences
//
// - A/B testing with different template variations
func NewCustomizeTemplate(id, subject, body string, subjectData, bodyData map[string]interface{}) Template {
	return &CustomizeTemplate{
		id:              id,
		subjectTemplate: subject,
		bodyTemplate:    body,
		subjectData:     subjectData,
		bodyData:        bodyData,
	}
}

func (t *CustomizeTemplate) GetID() string {
	return t.id
}

func (t *CustomizeTemplate) GetVariables() map[string]interface{} {
	// Merge subjectData and bodyData into a single map
	variables := make(map[string]interface{})
	for k, v := range t.subjectData {
		variables[k] = v
	}
	for k, v := range t.bodyData {
		variables[k] = v
	}
	return variables
}

func (t *CustomizeTemplate) RenderSubject() (string, error) {
	return t.Render(t.subjectTemplate, t.subjectData)
}

func (t *CustomizeTemplate) RenderBody() (string, error) {
	return t.Render(t.bodyTemplate, t.bodyData)
}

func (t *CustomizeTemplate) Render(templateString string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateString)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// NoTemplate represents a null object pattern implementation for the Template interface.
//
// It provides empty/nil implementations of all Template methods, useful for cases where
// no template processing is needed or as a default fallback when template creation fails.
//
// This avoids nil pointer checks and provides a safe way to handle missing templates.
type NoTemplate struct{}

// NewNoTemplate creates a new NoTemplate instance.
//
// Use this when you need a Template that does nothing, such as:
//
// - Default fallback when template loading fails
//
// - Avoid nil pointer checks
//
// - Testing scenarios where template content is not relevant
//
// - Placeholder in systems that require a Template but don't need actual rendering
func NewNoTemplate() Template {
	return &NoTemplate{}
}

func (t *NoTemplate) GetID() string {
	return ""
}

func (t *NoTemplate) GetVariables() map[string]interface{} {
	return nil
}

func (t *NoTemplate) RenderSubject() (string, error) {
	return "", nil
}

func (t *NoTemplate) RenderBody() (string, error) {
	return "", nil
}
