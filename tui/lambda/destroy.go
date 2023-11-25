package lambda

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type LambdaDestroyResponseMsg struct {
	Resp *LambdaDestroyResponse
}

type LambdaDestroyResponse struct {
	Err error
}

type LambdaDestroyer interface {
	Destroy(id string) tea.Cmd
}

type LambdaDestroyModel struct {
	LambdaID  string
	Destroyer LambdaDestroyer

	resp *LambdaDestroyResponse

	loadingSpinner spinner.Model
}

func InitLambdaDestroyModel(m *LambdaDestroyModel) *LambdaDestroyModel {
	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m LambdaDestroyModel) Init() tea.Cmd {
	return tea.Batch(m.Destroyer.Destroy(m.LambdaID), m.loadingSpinner.Tick)
}

func (m LambdaDestroyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case LambdaDestroyResponseMsg:
		m.resp = msg.Resp
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m LambdaDestroyModel) View() string {
	if m.resp == nil {
		return fmt.Sprintf("%s Destroying lambda...", m.loadingSpinner.View())
	}

	if m.resp.Err != nil {
		return fmt.Sprintf("Failed to destroy lambda: %s\n", m.resp.Err)
	} else {
		return "Lambda has been destroyed\n"
	}
}
