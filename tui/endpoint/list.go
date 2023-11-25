package endpoint

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	api "github.com/onpremless/go-client"
)

type EndpointListResponseMsg struct {
	Resp *EndpointListResponse
}

type EndpointListResponse struct {
	Endpoints []api.Endpoint
	Err       error
}

type EndpointLister interface {
	List() tea.Cmd
}

type EndpointListModel struct {
	Lister EndpointLister

	resp *EndpointListResponse

	loadingSpinner spinner.Model
}

func InitEndpointListModel(m *EndpointListModel) *EndpointListModel {
	m.loadingSpinner = spinner.New()

	m.loadingSpinner.Spinner = spinner.Dot

	return m
}

func (m EndpointListModel) Init() tea.Cmd {
	return tea.Batch(m.Lister.List(), m.loadingSpinner.Tick)
}

func (m EndpointListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case EndpointListResponseMsg:
		m.resp = msg.Resp
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	return m, cmd
}

func (m EndpointListModel) View() string {
	if m.resp == nil {
		return fmt.Sprintf("%s Query endpoints...", m.loadingSpinner.View())
	}

	if m.resp.Err != nil {
		return fmt.Sprintf("Failed to create endpoint: %s\n", m.resp.Err)
	} else {
		j, _ := json.MarshalIndent(m.resp.Endpoints, "", "  ")
		return string(j) + "\n"
	}
}
