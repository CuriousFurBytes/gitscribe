package app

import "testing"

func TestNewFormDefaults(t *testing.T) {
	form := newForm("title placeholder", "body placeholder", 72)
	if form.Focus != focusTitle {
		t.Fatalf("expected focus on title")
	}
	if form.Title.CharLimit != 72 {
		t.Fatalf("CharLimit = %d, want 72", form.Title.CharLimit)
	}
}

func TestFocusNextWrapsAroundCycle(t *testing.T) {
	form := newForm("a", "b", 72)
	form.focusNext()
	if form.Focus != focusBody {
		t.Fatalf("after one next, focus = %d, want body", form.Focus)
	}
	form.focusNext()
	if form.Focus != focusTitle {
		t.Fatalf("wrap-around should reach title, got %d", form.Focus)
	}
}

func TestFocusPrevWrapsAround(t *testing.T) {
	form := newForm("a", "b", 72)
	form.focusPrev()
	if form.Focus != focusBody {
		t.Fatalf("prev from title should land on body, got %d", form.Focus)
	}
}

func TestApplyFocusBlursOthers(t *testing.T) {
	form := newForm("a", "b", 72)
	form.Focus = focusBody
	form.applyFocus()
	if form.Title.Focused() {
		t.Fatalf("expected title blurred")
	}
	if !form.Body.Focused() {
		t.Fatalf("expected body focused")
	}
}

func TestValidateFormTitleRejectsBlank(t *testing.T) {
	if err := validateFormTitle("   "); err == nil {
		t.Fatalf("expected error on blank title")
	}
	if err := validateFormTitle("ok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNextFocusFallback(t *testing.T) {
	if got := nextFocus([]formFocus{focusTitle, focusBody}, formFocus(99)); got != focusTitle {
		t.Fatalf("expected fallback to first, got %d", got)
	}
}

func TestPrevFocusFallback(t *testing.T) {
	if got := prevFocus([]formFocus{focusTitle, focusBody}, formFocus(99)); got != focusTitle {
		t.Fatalf("expected fallback to first, got %d", got)
	}
}
