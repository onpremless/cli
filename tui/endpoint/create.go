package endpoint

import (
	"encoding/json"
	"fmt"

	"github.com/onpremless/cli/tui/lambda"
	api "github.com/onpremless/go-client"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	ECInitStep            = 0
	ECNameStep            = iota
	ECLambdasLoadingStep  = iota
	ECLambdaSelectionStep = iota
	ECEndpointStep        = iota
	ECLoadingStep         = iota
)

type EndpointCreateResponse struct {
	Endpoint *api.Endpoint
	Err      error
}

type EndpointCreateResponseMsg struct {
	Resp *EndpointCreateResponse
}

type lambdaItem struct {
	Lambda api.Lambda
}

func (i lambdaItem) Title() string       { return i.Lambda.Name }
func (i lambdaItem) Description() string { return i.Lambda.LambdaType }
func (i lambdaItem) FilterValue() string { return i.Lambda.Name }

type endpointCreateStartMsg struct{}

func endpointCreateStart() tea.Msg {
	return endpointCreateStartMsg{}
}

type EndpointCreator interface {
	Create(name string, path string, lambda string) tea.Cmd
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type EndpointCreateModel struct {
	Name            string
	Endpoint        string
	Lambda          *api.Lambda
	EndpointCreator EndpointCreator
	LambdaLister    lambda.LambdaLister

	static string

	lambdas       []api.Lambda
	lambdasLoaded bool
	resp          *EndpointCreateResponse

	step           int
	nameInput      textinput.Model
	pathInput      textinput.Model
	endpointList   list.Model
	loadingSpinner spinner.Model
}

func InitEndpointCreateModel(m *EndpointCreateModel) *EndpointCreateModel {
	m.nameInput = textinput.New()
	m.nameInput.CharLimit = 156
	m.nameInput.Placeholder = "Endpoint name"

	m.pathInput = textinput.New()
	m.pathInput.CharLimit = 156
	m.pathInput.Placeholder = "Endpoint path"

	m.endpointList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.endpointList.Title = "Lambda endpoints"
	m.endpointList.SetFilteringEnabled(false)

	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m EndpointCreateModel) Init() tea.Cmd {
	return endpointCreateStart
}

func (m EndpointCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case endpointCreateStartMsg:
		return m.incStep()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.endpointList.SetSize(msg.Width-h, msg.Height-v)
	}

	switch m.step {
	case ECNameStep:
		return m.handleECNameStep(msg)
	case ECLambdasLoadingStep:
		return m.handleECLambdasLoadingStep(msg)
	case ECLambdaSelectionStep:
		return m.handleECLambdaSelectionStep(msg)
	case ECEndpointStep:
		return m.handleECEndpointStep(msg)
	case ECLoadingStep:
		return m.handleECLoadingStep(msg)
	}
	return m, nil
}

func (m EndpointCreateModel) View() string {
	static := m.static
	if static != "" {
		static += "\n"
	}

	active := ""
	if m.step == ECNameStep {
		active = docStyle.Render(m.nameInput.View())
	} else if m.step == ECLambdasLoadingStep {
		active = docStyle.Render(fmt.Sprintf("%s Loading lambdas...", m.loadingSpinner.View()))
	} else if m.step == ECLambdaSelectionStep {
		return m.endpointList.View()
	} else if m.step == ECEndpointStep {
		active = docStyle.Render(m.pathInput.View())
	} else if m.step == ECLoadingStep {
		active = docStyle.Render(fmt.Sprintf("%s Creating endpoint...", m.loadingSpinner.View()))
	}

	return fmt.Sprintf("%s%s", static, active)
}

func (m EndpointCreateModel) handleECLambdasLoadingStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case lambda.LambdaListResponseMsg:
		if msg.Resp.Err != nil {
			m.static = fmt.Sprintf("%s\n\nFailed to load lambdas: %s", m.static, msg.Resp.Err)
			return m, tea.Quit
		}

		for _, lambda := range msg.Resp.Lambdas {
			if lambda.LambdaType == "ENDPOINT" {
				m.lambdas = append(m.lambdas, lambda)
			}
		}

		if len(m.lambdas) == 0 {
			m.static = fmt.Sprintf("%s\n\nNo suitable lambdas was found", m.static)
			return m, tea.Quit
		}

		m.lambdasLoaded = true
		return m.incStep()
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m EndpointCreateModel) handleECLambdaSelectionStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Lambda = &m.lambdas[m.endpointList.Cursor()]
			return m.incStep()
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.endpointList, cmd = m.endpointList.Update(msg)
	return m, cmd
}

func (m EndpointCreateModel) handleECNameStep(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m EndpointCreateModel) handleECEndpointStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Endpoint = m.pathInput.Value()
			return m.incStep()
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func (m EndpointCreateModel) handleECLoadingStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EndpointCreateResponseMsg:
		m.resp = msg.Resp
		return m.incStep()
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m *EndpointCreateModel) incStep(cmds ...tea.Cmd) (*EndpointCreateModel, tea.Cmd) {
	if m.step == ECInitStep {
		m.step++
		return m.incStep(m.nameInput.Cursor.SetMode(cursor.CursorBlink), m.nameInput.Focus())
	}

	if m.step == ECNameStep && m.Name != "" {
		m.step++
		m.nameInput.Blur()
		m.static = fmt.Sprintf("%s\nName: %s", m.static, m.Name)

		return m.incStep(m.LambdaLister.List(), m.loadingSpinner.Tick)
	}

	if m.step == ECLambdasLoadingStep && (m.Lambda != nil || m.lambdasLoaded) {
		m.step++

		target := []list.Item{}
		for _, lambda := range m.lambdas {
			target = append(target, lambdaItem{lambda})
		}

		return m.incStep(m.endpointList.SetItems(target))
	}

	if m.step == ECLambdaSelectionStep && m.Lambda != nil {
		m.step++
		m.static = fmt.Sprintf("%s\nLambda endpoint: %s", m.static, m.Lambda.Name)

		return m.incStep(m.pathInput.SetCursorMode(textinput.CursorBlink), m.pathInput.Focus())
	}

	if m.step == ECEndpointStep && m.Endpoint != "" {
		m.step++
		m.pathInput.Blur()
		m.static = fmt.Sprintf("%s\nPath: %s", m.static, m.Lambda.Name)

		return m.incStep(m.EndpointCreator.Create(m.Name, m.Endpoint, m.Lambda.Id), m.loadingSpinner.Tick)
	}

	if m.step == ECLoadingStep && m.resp != nil {
		m.step++
		if m.resp.Err != nil {
			m.static = fmt.Sprintf("%s\n\nFailed to create endpoint: %s", m.static, m.resp.Err)
		} else {
			j, _ := json.MarshalIndent(m.resp.Endpoint, "", "  ")
			m.static = fmt.Sprintf("%s\n\n%s", m.static, j)
		}

		return m.incStep(tea.Quit)
	}

	return m, tea.Batch(cmds...)
}
