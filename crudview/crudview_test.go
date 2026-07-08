//go:build !wasm

package crudview

import (
	"testing"

	. "github.com/tinywasm/html"
	"github.com/tinywasm/model"
)

func TestCrudView_Render_Basic(t *testing.T) {
	v := &CrudView{
		Title: "Test CRUD",
		Form:  Div().Text("The Form"),
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !contains(html, "Test CRUD") {
		t.Error("expected title")
	}
	if !contains(html, "The Form") {
		t.Error("expected form")
	}
	// Full page variant (no source)
	if !contains(html, "cv-article-contend-full-page") {
		t.Error("expected full page class")
	}
	if contains(html, "cv-aside-contend") {
		t.Error("did not expect aside without source")
	}
}

func TestCrudView_Render_WithSource(t *testing.T) {
	v := &CrudView{
		Title: "CRUD with List",
		Source: Source{
			Caller: &mockCaller{},
		},
		OnNew: func() {},
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !contains(html, "cv-article-contend") {
		t.Error("expected standard article class")
	}
	if !contains(html, "cv-aside-contend") {
		t.Error("expected aside with source")
	}
	if !contains(html, "name='btn_crudnew'") {
		t.Error("expected new button")
	}
	// OnDelete is nil, so no delete button
	if contains(html, "name='btn_cruddel'") {
		t.Error("did not expect delete button")
	}
}

type mockCaller struct{}
func (c *mockCaller) Call(op string, args model.Encodable) ([]byte, error) { return nil, nil }

type mockCtx struct{}
func (m *mockCtx) OnCleanup(fn func()) {}
