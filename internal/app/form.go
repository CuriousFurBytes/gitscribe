package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func newForm(titlePlaceholder string, bodyPlaceholder string, charLimit int) formState {
	title := textinput.New()
	title.Prompt = "> "
	title.Placeholder = titlePlaceholder
	title.CharLimit = charLimit
	title.Focus()

	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	body := textarea.New()
	body.Placeholder = bodyPlaceholder
	body.Prompt = "│"
	body.FocusedStyle.Prompt = muted
	body.BlurredStyle.Prompt = muted
	body.FocusedStyle.LineNumber = muted
	body.BlurredStyle.LineNumber = muted
	body.ShowLineNumbers = true
	body.SetHeight(8)
	body.Blur()

	feedback := textinput.New()
	feedback.Prompt = "> "
	feedback.Placeholder = "Describe what to change..."

	return formState{
		Title:         title,
		Body:          body,
		Focus:         focusTitle,
		FeedbackInput: feedback,
	}
}

func (f *formState) focusNext() {
	f.Focus = nextFocus([]formFocus{focusTitle, focusBody}, f.Focus)
	f.applyFocus()
}

func (f *formState) focusPrev() {
	f.Focus = prevFocus([]formFocus{focusTitle, focusBody}, f.Focus)
	f.applyFocus()
}

func (f *formState) applyFocus() {
	switch f.Focus {
	case focusTitle:
		f.Title.Focus()
		f.Body.Blur()
	case focusBody:
		f.Title.Blur()
		f.Body.Focus()
	default:
		f.Title.Blur()
		f.Body.Blur()
	}
}

func validateFormTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	return nil
}

func nextFocus(order []formFocus, current formFocus) formFocus {
	for i, focus := range order {
		if focus == current {
			return order[(i+1)%len(order)]
		}
	}
	return order[0]
}

func prevFocus(order []formFocus, current formFocus) formFocus {
	for i, focus := range order {
		if focus == current {
			return order[(i-1+len(order))%len(order)]
		}
	}
	return order[0]
}
