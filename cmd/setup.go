/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"lazysync/application"
	"os"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up your application",
	Long:  `Run application set up process`,
	Run: func(cmd *cobra.Command, args []string) {
		app := SetupApplication()
		app.Setup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

type configModel struct {
	choices     []string
	cursor      int
	selected    map[int]application.App
	quit        bool
	application application.App
}

func SetupApplication() application.App {
	var app application.App
	setup := tea.NewProgram(initialModel())
	teaModel, err := setup.Run()
	if err != nil {
		fmt.Println("\n" + err.Error())
		os.Exit(1)
	}
	updatedModel := teaModel.(configModel)
	if updatedModel.quit {
		return app
	}
	app = updatedModel.application
	return app
}

func initialModel() *configModel {

	return &configModel{
		// Our to-do list is a grocery list
		choices: []string{"Server", "Client"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		//selected: make(map[int]struct{}),
		selected: map[int]application.App{
			0: &application.Server{},
			1: &application.Client{},
		},
		quit: false,
	}
}

func (m configModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key toggles
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			instance := m.selected[m.cursor]
			m.application = instance

			return m, tea.Quit
			/*_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}*/
		}
	}

	// Return the updated configModel to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m configModel) View() string {
	// The header
	s := "Select system role:\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
