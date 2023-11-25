package lambda

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	api "github.com/onpremless/go-client"
)

type LambdaListResponseMsg struct {
	Resp *LambdaListResponse
}

type LambdaListResponse struct {
	Lambdas []api.Lambda
	Err     error
}

type LambdaLister interface {
	List() tea.Cmd
}

type LambdaListModel struct {
	Lister LambdaLister

	resp *LambdaListResponse

	loadingSpinner spinner.Model
}

func InitLambdaListModel(m *LambdaListModel) *LambdaListModel {
	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m LambdaListModel) Init() tea.Cmd {
	return tea.Batch(m.Lister.List(), m.loadingSpinner.Tick)
}

func (m LambdaListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case LambdaListResponseMsg:
		m.resp = msg.Resp
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m LambdaListModel) View() string {
	if m.resp == nil {
		return fmt.Sprintf("%s Query lambdas...", m.loadingSpinner.View())
	}

	if m.resp.Err != nil {
		return fmt.Sprintf("Failed to create lambda: %s\n", m.resp.Err)
	} else {
		j, _ := json.MarshalIndent(m.resp.Lambdas, "", "  ")
		return string(j) + "\n"
	}
}
