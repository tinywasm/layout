//go:build !wasm

package crudview

import (
	"testing"

	. "github.com/tinywasm/html"
	. "github.com/tinywasm/fmt"
)

func TestCrudView_Render_Basic(t *testing.T) {
	v := &CrudView{
		Title: "Test CRUD",
		Form:  Div().Text("The Form"),
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !Contains(html, "Test CRUD") {
		t.Error("expected title")
	}
	if !Contains(html, "The Form") {
		t.Error("expected form")
	}
	// Full page variant (no presenter)
	if !Contains(html, "cv-article-contend-full-page") {
		t.Error("expected full page class")
	}
	if Contains(html, "cv-aside-contend") {
		t.Error("did not expect aside without source")
	}
}

func TestCrudView_Render_WithSource(t *testing.T) {
	p := &fakePresenter{title: "CRUD with List", record: &Device{}, canDelete: false}
	v := &CrudView{
		Title: "CRUD with List",
		Presenter: p,
		OnNew: func() {},
	}
	v.Init(&mockCtx{})

	html := v.Render().String()

	if !Contains(html, "cv-article-contend") {
		t.Error("expected standard article class")
	}
	if !Contains(html, "cv-aside-contend") {
		t.Error("expected aside with source")
	}
	if !Contains(html, "name='btn_crudnew'") {
		t.Error("expected new button")
	}
	// OnDelete is nil, so no delete button
	if Contains(html, "name='btn_cruddel'") {
		t.Error("did not expect delete button")
	}
}

type mockCtx struct{}
func (m *mockCtx) OnCleanup(fn func()) {}
