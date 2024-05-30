package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
	"time"
)

type model struct {
	filepicker    filepicker.Model
	selectedFiles []string
	quitting      bool
	err           error
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "d":

		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		if didSelect, _ = m.filepicker.DidSelectDisabledFile(msg); didSelect {
			m.err = errors.New(path + " is disabled")
			return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
		}
		// Get the path of the selected file.
		m.selectedFiles = append(m.selectedFiles, path)
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if len(m.selectedFiles) == 0 {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected files:")
		for _, file := range m.selectedFiles {
			s.WriteString("\n- " + m.filepicker.Styles.Selected.Render(file))
		}
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}

func Setup() []string {
	fp := filepicker.New()
	//fp.AllowedTypes = []string{}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	m := model{
		filepicker: fp,
	}
	tm, _ := tea.NewProgram(&m).Run()
	mm := tm.(model)
	fmt.Println("\n  You selected:")
	var selectedFiles []string
	for _, file := range mm.selectedFiles {
		selectedFile := m.filepicker.Styles.Selected.Render(file)
		selectedFiles = append(selectedFiles, file)
		fmt.Println("- " + selectedFile)
	}
	return selectedFiles
}
