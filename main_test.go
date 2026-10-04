package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// fakeKeyboard stands for the system's keyboard in tests.
type fakeKeyboard struct {
	disabled bool
	err      error
}

func (k *fakeKeyboard) Disable() error {
	if k.err != nil {
		return k.err
	}
	k.disabled = true
	return nil
}

func (k *fakeKeyboard) Enable() { k.disabled = false }

// The view runs without a window in tests, which click and type as a user
// would.
func TestView(t *testing.T) {
	kb := &fakeKeyboard{}
	a := &app{kb: kb}
	tt := ui.NewTester(a.view, 280, 300)
	if !tt.HasText("Keyboard Unlocked") {
		t.Fatalf("texts %q at the start", tt.Texts())
	}

	if err := tt.Click("Lock Keyboard"); err != nil {
		t.Fatal(err)
	}
	if !kb.disabled || !tt.HasText("Keyboard Locked") {
		t.Errorf("disabled %v after a click, texts %q", kb.disabled, tt.Texts())
	}

	if err := tt.Click("Lock Keyboard"); err != nil {
		t.Fatal(err)
	}
	if kb.disabled || !tt.HasText("Keyboard Unlocked") {
		t.Errorf("disabled %v after a second click, texts %q", kb.disabled, tt.Texts())
	}
}

// Without the Accessibility permission the keyboard stays on, and the view
// says how to give it.
func TestViewWithoutAccess(t *testing.T) {
	a := &app{kb: &fakeKeyboard{err: errNoAccess}}
	tt := ui.NewTester(a.view, 280, 300)
	if err := tt.Click("Lock Keyboard"); err != nil {
		t.Fatal(err)
	}
	if a.disabled || !tt.HasText("Keyboard Unlocked") || !tt.HasText("Open Accessibility Settings") {
		t.Errorf("disabled %v without access, texts %q", a.disabled, tt.Texts())
	}
}
