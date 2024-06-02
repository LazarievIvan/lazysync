/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"lazysync/application"
	"lazysync/application/client"
	"lazysync/application/server"
	"lazysync/application/service"
	"lazysync/modules"
	"os"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up your application",
	Long:  `Run application set up process`,
	Run: func(cmd *cobra.Command, args []string) {
		app := SetupApplication()
		fmt.Println("Selected mode: " + app.GetType())
		module := setupModule()
		if app.GetType() == server.Type {
			module.SetupModule()
			fmt.Println("Selected module: " + module.GetId())
		}
		app.SetMode(module)
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

type modulesModel struct {
	choices        []string
	cursor         int
	selected       map[int]string
	quit           bool
	moduleHandler  *modules.ModuleHandler
	selectedModule string
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
			0: &server.Server{Configuration: &service.AppConfiguration{}},
			1: &client.Client{Configuration: &service.AppConfiguration{}},
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

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func setupModule() modules.Module {
	setup := tea.NewProgram(moduleSelectModel())
	teaModel, err := setup.Run()
	if err != nil {
		fmt.Println("\n" + err.Error())
		os.Exit(1)
	}
	updatedModel := teaModel.(modulesModel)
	selected := updatedModel.selectedModule
	module, err := updatedModel.moduleHandler.GetModuleByName(selected)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return module
}

func moduleSelectModel() *modulesModel {
	modulesHandler := modules.InitModuleHandler()
	return &modulesModel{
		// Our to-do list is a grocery list
		choices: modulesHandler.GetModuleNamesList(),

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected:       make(map[int]string),
		quit:           false,
		moduleHandler:  modulesHandler,
		selectedModule: "",
	}
}

func (m modulesModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m modulesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.selectedModule = m.choices[m.cursor]
			fmt.Println("Selected module: " + m.selectedModule)

			return m, tea.Quit
		}
	}

	// Return the updated configModel to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m modulesModel) View() string {
	// The header
	s := "Select module:\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
