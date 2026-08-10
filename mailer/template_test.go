package mailer

import (
	"testing"
)

func TestCustomizeTemplate_New(t *testing.T) {
	subject := "Hello, {{.Name}}"
	body := "Welcome to {{.Platform}}!"
	subjectData := map[string]interface{}{"Name": "Alice"}
	bodyData := map[string]interface{}{"Platform": "Unlimit"}

	tmpl := NewCustomizeTemplate("tmpl1", subject, body, subjectData, bodyData)
	ct, ok := tmpl.(*CustomizeTemplate)
	if !ok {
		t.Fatalf("Expected *CustomizeTemplate, got %T", tmpl)
	}
	if ct.GetID() != "tmpl1" {
		t.Errorf("Expected ID 'tmpl1', got %s", ct.GetID())
	}
}

func TestCustomizeTemplate_RenderSubject(t *testing.T) {
	subject := "Hello, {{.Name}}"
	subjectData := map[string]interface{}{"Name": "Alice"}
	tmpl := NewCustomizeTemplate("id", subject, "", subjectData, nil)
	result, err := tmpl.RenderSubject()
	if err != nil {
		t.Fatalf("RenderSubject() error: %v", err)
	}
	if result != "Hello, Alice" {
		t.Errorf("Expected 'Hello, Alice', got '%s'", result)
	}
}

func TestCustomizeTemplate_RenderBody(t *testing.T) {
	body := "Welcome to {{.Platform}}!"
	bodyData := map[string]interface{}{"Platform": "Unlimit"}
	tmpl := NewCustomizeTemplate("id", "", body, nil, bodyData)
	result, err := tmpl.RenderBody()
	if err != nil {
		t.Fatalf("RenderBody() error: %v", err)
	}
	if result != "Welcome to Unlimit!" {
		t.Errorf("Expected 'Welcome to Unlimit!', got '%s'", result)
	}
}

func TestCustomizeTemplate_GetVariables(t *testing.T) {
	subjectData := map[string]interface{}{"A": 1}
	bodyData := map[string]interface{}{"B": 2}
	tmpl := NewCustomizeTemplate("id", "", "", subjectData, bodyData)
	vars := tmpl.GetVariables()
	if vars["A"] != 1 || vars["B"] != 2 {
		t.Errorf("Expected merged variables, got %v", vars)
	}
}

func TestCustomizeTemplate_RenderSubject_Error(t *testing.T) {
	subject := "Hello, {{.Name}"
	subjectData := map[string]interface{}{"Name": "Alice"}
	tmpl := NewCustomizeTemplate("id", subject, "", subjectData, nil)
	_, err := tmpl.RenderSubject()
	if err == nil {
		t.Error("Expected error for invalid subject template, got nil")
	}
}

func TestCustomizeTemplate_RenderBody_Error(t *testing.T) {
	body := "Welcome to {{.Platform}"
	bodyData := map[string]interface{}{"Platform": "Unlimit"}
	tmpl := NewCustomizeTemplate("id", "", body, nil, bodyData)
	_, err := tmpl.RenderBody()
	if err == nil {
		t.Error("Expected error for invalid body template, got nil")
	}
}

func TestCustomizeTemplate_Render_ExecutionError(t *testing.T) {
	body := "Welcome, {{.Name}}!"
	bodyData := map[string]interface{}{} // Missing "Name"
	tmpl := NewCustomizeTemplate("id", "", body, nil, bodyData)
	_, err := tmpl.RenderBody()
	if err == nil {
		t.Error("Expected error for missing variable in body template, got nil")
	}
}
