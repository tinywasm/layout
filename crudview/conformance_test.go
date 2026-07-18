package crudview

import (
	"testing"

	"github.com/tinywasm/fmt"
	"github.com/tinywasm/view"
	"github.com/tinywasm/view/conformance"
)

func TestViewConformance(t *testing.T) {
	conformance.Run(t, conformance.Factory{
		New: func(t *testing.T, p view.Presenter) conformance.Driver {
			v, err := New(Config{ParentID: "conformance", Presenter: p})
			if err != nil {
				t.Fatalf("crudview.New: %v", err)
			}
			v.Init(&fakeCtx{})

			// Skip validation on form inputs during conformance tests because MockRecord
			// fields (like Name "Bob" with ID "2" or Name "X") violate standard form validation
			// defaults (like Text's default 2-character minimum). Doing this here keeps our
			// production code completely free of test-aware validation-bypass logic.
			if v.form != nil {
				for _, inp := range v.form.Inputs {
					if skipper, ok := inp.(interface{ SetSkipValidation(bool) }); ok {
						skipper.SetSkipValidation(true)
					}
				}
			}

			return conformance.Driver{
				Mount:  func() { _ = v.Reload() },
				Labels: func() []string { return cardLabels(v) },
				Select: func(id string) { v.selectAction(view.Item{ID: id}) },
				SetField: func(name, value string) { v.form.SetValues(name, value) },
				Save: func() {
					if s, ok := p.(view.Saver); ok { v.saveAction(s) }
				},
				Delete: func() {
					if d, ok := p.(view.Deleter); ok { v.deleteAction(d, v.selected.Get()) }
				},
			}
		},
	})
}

// cardLabels extracts the visible label of each rendered card. A card renders as
// <li class='…'>LABEL<span class='…'>DESC</span></li>; the label is the text
// between the first '>' and the following '<' of el.String().
func cardLabels(v *CrudView) []string {
	els := v.items.Get()
	labels := make([]string, 0, len(els))
	for _, el := range els {
		s := el.String()
		start := fmt.Index(s, ">")
		end := -1
		for i := start + 1; i < len(s); i++ {
			if s[i] == '<' {
				end = i
				break
			}
		}
		if start >= 0 && end > start {
			labels = append(labels, s[start+1:end])
		}
	}
	return labels
}
