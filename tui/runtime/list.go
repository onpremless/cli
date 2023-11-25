package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	api "github.com/onpremless/go-client"
)

type RuntimeListResponseMsg struct {
	Resp *RuntimeListResponse
}

type RuntimeListResponse struct {
	Runtimes []api.Runtime
	Err      error
}

type RuntimeLister interface {
	List() tea.Cmd
}

type RuntimeListModel struct {
	Lister RuntimeLister

	resp *RuntimeListResponse

	loadingSpinner spinner.Model
}

func InitRuntimeListModel(m *RuntimeListModel) *RuntimeListModel {
	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m RuntimeListModel) Init() tea.Cmd {
	return tea.Batch(m.Lister.List(), m.loadingSpinner.Tick)
}

func (m RuntimeListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case RuntimeListResponseMsg:
		m.resp = msg.Resp
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m RuntimeListModel) View() string {
	if m.resp == nil {
		return fmt.Sprintf("%s Query runtimes...", m.loadingSpinner.View())
	}

	if m.resp.Err != nil {
		return fmt.Sprintf("Failed to create runtime: %s\n", m.resp.Err)
	} else {
		j, _ := json.MarshalIndent(m.resp.Runtimes, "", "  ")
		return string(j) + "\n"
	}
}
