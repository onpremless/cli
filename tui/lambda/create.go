package lambda

import (
	"encoding/json"
	"fmt"

	api "github.com/onpremless/go-client"
	"github.com/onpremless/opcli/tui/runtime"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	LCInitStep             = 0
	LCCNameStep            = iota
	LCRuntimesLoadingStep  = iota
	LCRuntimeSelectionStep = iota
	LCTypeStep             = iota
	LCLoadingStep          = iota
)

type LambdaCreateResponse struct {
	Lambda *api.Lambda
	Err    error
}

type LambdaCreateResponseMsg struct {
	Resp *LambdaCreateResponse
}

type runtimeItem struct {
	Runtime api.Runtime
}

func (i runtimeItem) Title() string       { return i.Runtime.Name }
func (i runtimeItem) Description() string { return i.Runtime.GetId() }
func (i runtimeItem) FilterValue() string { return i.Runtime.Name }

type typeItem struct {
	LambdaType string
	Name       string
	Desc       string
}

func (i typeItem) Title() string       { return i.Name }
func (i typeItem) Description() string { return i.Desc }
func (i typeItem) FilterValue() string { return i.Name }

type lambdaCreateStartMsg struct{}

func lambdaCreateStart() tea.Msg {
	return lambdaCreateStartMsg{}
}

type LambdaCreator interface {
	Create(name string, runtime string, lambdaType string, path string) tea.Cmd
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type LambdaCreateModel struct {
	Name          string
	LambdaType    string
	Path          string
	Runtime       *api.Runtime
	LambdaCreator LambdaCreator
	RuntimeLister runtime.RuntimeLister

	static string

	lambdaTypes    []typeItem
	runtimes       []api.Runtime
	runtimesLoaded bool
	resp           *LambdaCreateResponse

	step           int
	nameInput      textinput.Model
	runtimeList    list.Model
	typeList       list.Model
	loadingSpinner spinner.Model
}

func InitLambdaCreateModel(m *LambdaCreateModel) *LambdaCreateModel {
	m.lambdaTypes = []typeItem{
		{LambdaType: "ENDPOINT", Name: "Endpoint", Desc: "Could be called outside"},
		{LambdaType: "INTERNAL", Name: "Intenal", Desc: "Being used for internal communication only"},
	}

	m.nameInput = textinput.New()
	m.nameInput.CharLimit = 156
	m.nameInput.Placeholder = "Lambda name"

	m.runtimeList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.runtimeList.Title = "Lambda endpoints"
	m.runtimeList.SetFilteringEnabled(false)

	genericLambdaTypes := make([]list.Item, len(m.lambdaTypes))
	for i, ltype := range m.lambdaTypes {
		genericLambdaTypes[i] = ltype
	}
	m.typeList = list.New(genericLambdaTypes, list.NewDefaultDelegate(), 0, 0)
	m.typeList.Title = "Lambda type"
	m.typeList.SetFilteringEnabled(false)

	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m LambdaCreateModel) Init() tea.Cmd {
	return lambdaCreateStart
}

func (m LambdaCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case lambdaCreateStartMsg:
		return m.incStep()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		x, y := docStyle.GetFrameSize()
		m.runtimeList.SetSize(msg.Width-x, msg.Height-y)
		m.typeList.SetSize(msg.Width-x, msg.Height-y)
	}

	switch m.step {
	case LCCNameStep:
		return m.handleLCNameStep(msg)
	case LCRuntimesLoadingStep:
		return m.handleLCRuntimesLoadingStep(msg)
	case LCRuntimeSelectionStep:
		return m.handleLCLambdaSelectionStep(msg)
	case LCTypeStep:
		return m.handleLCTypeStep(msg)
	case LCLoadingStep:
		return m.handleLCLoadingStep(msg)
	}
	return m, nil
}

func (m LambdaCreateModel) View() string {
	static := m.static
	if static != "" {
		static += "\n"
	}

	active := ""
	if m.step == LCCNameStep {
		active = docStyle.Render(m.nameInput.View())
	} else if m.step == LCRuntimesLoadingStep {
		active = docStyle.Render(fmt.Sprintf("%s Loading runtimes...", m.loadingSpinner.View()))
	} else if m.step == LCRuntimeSelectionStep {
		return docStyle.Render(m.runtimeList.View())
	} else if m.step == LCTypeStep {
		return docStyle.Render(m.typeList.View())
	} else if m.step == LCLoadingStep {
		active = docStyle.Render(fmt.Sprintf("%s Creating lambda...", m.loadingSpinner.View()))
	}

	return fmt.Sprintf("%s%s", static, active)
}

func (m LambdaCreateModel) handleLCRuntimesLoadingStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case runtime.RuntimeListResponseMsg:
		if msg.Resp.Err != nil {
			m.static = fmt.Sprintf("%s\n\nFailed to load runtimes: %s", m.static, msg.Resp.Err)
			return m, tea.Quit
		}

		m.runtimes = msg.Resp.Runtimes

		if len(m.runtimes) == 0 {
			m.static = fmt.Sprintf("%s\n\nNo runtimes was found", m.static)
			return m, tea.Quit
		}

		m.runtimesLoaded = true
		return m.incStep()
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m LambdaCreateModel) handleLCLambdaSelectionStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.Runtime = &m.runtimes[m.runtimeList.Cursor()]
			return m.incStep()
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.runtimeList, cmd = m.runtimeList.Update(msg)
	return m, cmd
}

func (m LambdaCreateModel) handleLCNameStep(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m LambdaCreateModel) handleLCTypeStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.LambdaType = m.lambdaTypes[m.typeList.Cursor()].LambdaType
			return m.incStep()
		case tea.KeyEsc:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.typeList, cmd = m.typeList.Update(msg)
	return m, cmd
}

func (m LambdaCreateModel) handleLCLoadingStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LambdaCreateResponseMsg:
		m.resp = msg.Resp
		return m.incStep()
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m *LambdaCreateModel) incStep(cmds ...tea.Cmd) (*LambdaCreateModel, tea.Cmd) {
	if m.step == LCInitStep {
		m.step++
		return m.incStep(m.nameInput.Cursor.SetMode(cursor.CursorBlink), m.nameInput.Focus())
	}

	if m.step == LCCNameStep && m.Name != "" {
		m.step++
		m.nameInput.Blur()
		m.static = fmt.Sprintf("%s\nName: %s", m.static, m.Name)

		return m.incStep(m.RuntimeLister.List(), m.loadingSpinner.Tick)
	}

	if m.step == LCRuntimesLoadingStep && (m.Runtime != nil || m.runtimesLoaded) {
		m.step++

		target := []list.Item{}
		for _, lambda := range m.runtimes {
			target = append(target, runtimeItem{lambda})
		}

		return m.incStep(m.runtimeList.SetItems(target))
	}

	if m.step == LCRuntimeSelectionStep && m.Runtime != nil {
		m.step++
		m.static = fmt.Sprintf("%s\nRuntime: %s", m.static, m.Runtime.Name)

		return m.incStep()
	}

	if m.step == LCTypeStep && m.LambdaType != "" {
		m.step++
		m.static = fmt.Sprintf("%s\nLambda type: %s", m.static, m.LambdaType)

		return m.incStep(m.LambdaCreator.Create(m.Name, m.Runtime.Id, m.LambdaType, m.Path), m.loadingSpinner.Tick)
	}

	if m.step == LCLoadingStep && m.resp != nil {
		m.step++
		if m.resp.Err != nil {
			m.static = fmt.Sprintf("%s\n\nFailed to create lambda: %s", m.static, m.resp.Err)
		} else {
			j, _ := json.MarshalIndent(m.resp.Lambda, "", "  ")
			m.static = fmt.Sprintf("%s\n\n%s", m.static, j)
		}

		return m.incStep(tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m LambdaCreateModel) GetLambda() *api.Lambda {
	if m.resp == nil {
		return nil
	}

	return m.resp.Lambda
}
