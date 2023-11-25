package lambda

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	api "github.com/onpremless/go-client"
)

type LambdaStartResponseMsg struct {
	Resp *LambdaStartResponse
}

type LambdaStartResponse struct {
	Lambda *api.Lambda
	Err    error
}

type LambdaStarter interface {
	Start(id string) tea.Cmd
}

type LambdaStartModel struct {
	LambdaID string
	Starter  LambdaStarter

	resp *LambdaStartResponse

	loadingSpinner spinner.Model
}

func InitLambdaStartModel(m *LambdaStartModel) *LambdaStartModel {
	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m LambdaStartModel) Init() tea.Cmd {
	return tea.Batch(m.Starter.Start(m.LambdaID), m.loadingSpinner.Tick)
}

func (m LambdaStartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case LambdaStartResponseMsg:
		m.resp = msg.Resp
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m LambdaStartModel) View() string {
	if m.resp == nil {
		return fmt.Sprintf("%s Starting lambda...", m.loadingSpinner.View())
	}

	if m.resp.Err != nil {
		return fmt.Sprintf("Failed to start lambda: %s\n", m.resp.Err)
	} else {
		j, _ := json.MarshalIndent(m.resp.Lambda, "", "  ")
		return string(j) + "\n"
	}
}
