package runtime

import (
	"encoding/json"
	"fmt"

	api "github.com/onpremless/go-client"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	RCInitStep    = 0
	RCNameStep    = iota
	RCLoadingStep = iota
)

type RuntimeCreateResponse struct {
	Runtime *api.Runtime
	Err     error
}

type RuntimeCreateResponseMsg struct {
	Resp *RuntimeCreateResponse
}

type runtimeCreateStartMsg struct{}

func runtimeCreateStart() tea.Msg {
	return runtimeCreateStartMsg{}
}

type RuntimeCreator interface {
	Create(name string, path string) tea.Cmd
}

type RuntimeCreateModel struct {
	Name    string
	Path    string
	Creator RuntimeCreator

	static string

	resp *RuntimeCreateResponse

	step           int
	nameInput      textinput.Model
	loadingSpinner spinner.Model
}

func InitRuntimeCreateModel(m *RuntimeCreateModel) *RuntimeCreateModel {
	m.nameInput = textinput.New()
	m.nameInput.CharLimit = 156
	m.nameInput.Placeholder = "Runtime name"

	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m RuntimeCreateModel) Init() tea.Cmd {
	return runtimeCreateStart
}

func (m RuntimeCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case runtimeCreateStartMsg:
		return m.incStep()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	switch m.step {
	case RCNameStep:
		return m.handleRCNameStep(msg)
	case RCLoadingStep:
		return m.handleRCLoadingStep(msg)
	}
	return m, nil
}

func (m RuntimeCreateModel) View() string {
	static := m.static
	if static != "" {
		static += "\n\n"
	}

	active := ""
	if m.step == RCNameStep {
		active = m.nameInput.View()
	} else if m.step == RCLoadingStep {
		active = fmt.Sprintf("%s Creating runtime...", m.loadingSpinner.View())
	}

	return fmt.Sprintf("%s%s", static, active)
}

func (m RuntimeCreateModel) handleRCNameStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Name = m.nameInput.Value()
			return m.incStep()
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m RuntimeCreateModel) handleRCLoadingStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RuntimeCreateResponseMsg:
		m.resp = msg.Resp
		return m.incStep()
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m *RuntimeCreateModel) incStep(cmds ...tea.Cmd) (*RuntimeCreateModel, tea.Cmd) {
	if m.step == RCInitStep {
		m.step++
		m.static = fmt.Sprintf("Path: %s", m.Path)
		return m.incStep(m.nameInput.Cursor.SetMode(cursor.CursorBlink), m.nameInput.Focus())
	}

	if m.step == RCNameStep && m.Name != "" {
		m.step++
		m.nameInput.Blur()
		m.static = fmt.Sprintf("%s\nName: %s", m.static, m.Name)

		return m.incStep(m.Creator.Create(m.Name, m.Path), m.loadingSpinner.Tick)
	}

	if m.step == RCLoadingStep && m.resp != nil {
		m.step++
		if m.resp.Err != nil {
			m.static = fmt.Sprintf("%s\n\nFailed to create runtime: %s", m.static, m.resp.Err)
		} else {
			j, _ := json.MarshalIndent(m.resp.Runtime, "", "  ")
			m.static = fmt.Sprintf("%s\n\n%s", m.static, j)
		}

		return m.incStep(tea.Quit)
	}

	return m, tea.Batch(cmds...)
}
