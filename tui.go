package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aryanA101a/villi/ui"
	"github.com/aryanA101a/villi/utils"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)
var titleStyle=lipgloss.NewStyle().Background(lipgloss.Color("63")).Bold(true).Align(lipgloss.Center).Width(maxWidth+15).Render
var borderStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("63")).Render
var downloadPercentageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true).Render
var helpStyle = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#626262")).Render
var metaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Align(lipgloss.Center).Width(maxWidth + 15).Render

const (
	padding  = 2
	maxWidth = 80
)

type progressErrMsg struct{ err error }

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type model struct {
	meta        ui.Meta
	progress    ui.Progress
	progressBar progress.Model
	err         error
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progressBar.Width = msg.Width - padding*2 - 8
		if m.progressBar.Width > maxWidth {
			m.progressBar.Width = maxWidth
		}
		return m, nil

	case progressErrMsg:
		m.err = msg.err
		return m, tea.Quit

	case ui.Progress:
		m.progress = ui.Progress{
			Ratio:      msg.Ratio,
			Downloaded: msg.Downloaded,
		}

		var cmds []tea.Cmd

		if msg.Ratio >= 1.0 {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.progressBar.SetPercent(float64(msg.Ratio)))
		return m, tea.Batch(cmds...)

	case ui.Status:
		m.meta.Status = msg
		return m, nil
	case ui.FileSize:
		m.meta.FileSize = msg
		return m, nil
	case ui.ConnectedPeers:
		m.meta.ConnectedPeers = msg
		return m, nil
	case ui.Peers:
		m.meta.Peers = msg
		return m, nil
	case ui.FileName:
		m.meta.FileName= msg
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.err != nil {
		return "Error downloading: " + m.err.Error() + "\n"
	}
	meta := fmt.Sprintf("%s/%s 🔽 | %d/%d peers     status:%s", utils.ConvertToHumanReadable(m.progress.Downloaded), m.meta.FileSize, m.meta.ConnectedPeers, m.meta.Peers, m.meta.Status)
	percentage := fmt.Sprintf("   %s", strconv.FormatFloat(m.progress.Ratio*100, 'f', 2, 64)) + "%"
	pad := strings.Repeat(" ", padding)
	return borderStyle(titleStyle(string(m.meta.FileName))+"\n\n" +
		pad + m.progressBar.View() + downloadPercentageStyle(percentage) + pad + "\n\n" + metaStyle(meta) + "\n\n" +
		pad + helpStyle("Press any key to quit"))
}
