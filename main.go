package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


const height = 5

// var bigChars = map[rune][]string{
// 	'0': {"████", "█  █", "█  █", "█  █", "████"},
// 	'1': {"   █", "   █", "   █", "   █", "   █"},
// 	'2': {"████", "   █", "████", "█   ", "████"},
// 	'3': {"████", "   █", "████", "   █", "████"},
// 	'4': {"█  █", "█  █", "████", "   █", "   █"},
// 	'5': {"████", "█   ", "████", "   █", "████"},
// 	'6': {"████", "█   ", "████", "█  █", "████"},
// 	'7': {"████", "   █", "   █", "   █", "   █"},
// 	'8': {"████", "█  █", "████", "█  █", "████"},
// 	'9': {"████", "█  █", "████", "   █", "████"},
// 	':': {"    ", "  █ ", "    ", "  █ ", "    "},
// }

var bigChars = map[rune][]string{
	'0': {
		"██████",
		"██  ██",
		"██  ██",
		"██  ██",
		"██████",
	},
	'1': {
		"  ██  ", 
		"  ██  ",
		"  ██  ",
		"  ██  ",
		"  ██  ",
	},
	'2': {
		"██████",
		"    ██",
		"██████",
		"██    ",
		"██████",
	},
	'3': {
		"██████",
		"    ██",
		"██████",
		"    ██",
		"██████",
	},
	'4': {
		"██  ██",
		"██  ██",
		"██████",
		"    ██",
		"    ██",
	},
	'5': {
		"██████",
		"██    ",
		"██████",
		"    ██",
		"██████",
	},
	'6': {
		"██████",
		"██    ",
		"██████",
		"██  ██",
		"██████",
	},
	'7': {
		"██████",
		"    ██",
		"    ██",
		"    ██",
		"    ██",
	},
	'8': {
		"██████",
		"██  ██",
		"██████",
		"██  ██",
		"██████",
	},
	'9': {
		"██████",
		"██  ██",
		"██████",
		"    ██",
		"██████",
	},
	':': {
		"      ",
		"  ██  ",
		"      ",
		"  ██  ",
		"      ",
	},
}
func RenderTime(s string) string {
	lines := make([]string, height)
	for idx, char := range s {
		sprite, ok := bigChars[char]
		if !ok {
			continue
		}
		for i := range height {
			lines[i] += sprite[i]
			if idx < len(s)-1 {
				lines[i] += "  " //space between nubers
			}
		}
	}
	// for i := range lines {
	// 	lines[i] = strings.TrimRight(lines[i], " ")
	// }
	return strings.Join(lines, "\n")
}
type sessionState int

const (
	stateSetupStudy sessionState = iota 
	stateSetupRest                      
	stateRunning                        
)

type model struct {
	height    int
	width     int
	
	studyMin  int
	restMin   int
	
	timeLeft  int
	isResting bool        
	state     sessionState 
	
	input     textinput.Model
	err       error
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initModel() model {
	ti := textinput.New()
	ti.Placeholder = "45"
	ti.Focus()
	ti.CharLimit = 3
	ti.Width = 5

	return model{
		state:    stateSetupStudy,
		input:    ti,
		studyMin: 45, // Default
		restMin:  15,  // Default
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		
		case "esc":
			// reset
			m.state = stateSetupStudy
			m.input.SetValue("")
			m.input.Placeholder = "25"
			m.input.Focus()
			m.err = nil
			return m, nil

		case "enter":
			if m.state == stateSetupStudy || m.state == stateSetupRest {
				// parsing input
				val := m.input.Value()
				// default if empty
				num := 0
				var err error
				
				if val == "" {
					if m.state == stateSetupStudy { num = 45 } else { num = 15 }
				} else {
					num, err = strconv.Atoi(val)
				}

				if err != nil {
					m.err = err
					return m, nil
				}
				m.err = nil

				// Gestione transizione stati
				switch m.state {
				case stateSetupStudy:
					m.studyMin = num
					m.state = stateSetupRest 
					m.input.SetValue("")
					m.input.Placeholder = "15"
					return m, nil
				case stateSetupRest:
					m.restMin = num
					// start timer
					m.state = stateRunning
					m.isResting = false //always start with studytime
					m.timeLeft = m.studyMin * 60
					m.input.Blur()
					return m, tick()
				}
			}
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tickMsg:
		if m.state == stateRunning {
			if m.timeLeft > 0 {
				m.timeLeft--
				return m, tick()
			} else {
				// switch mode
				m.isResting = !m.isResting 
				
				if m.isResting {
					// start rest
					m.timeLeft = m.restMin * 60
				} else {
					// start study
					m.timeLeft = m.studyMin * 60
				}
				
				// Opzionale: suonare una campanella (bell character) ??????
				// fmt.Print("\a") 
				
				return m, tick() // restart timer
			}
		}
	}

	if m.state == stateSetupStudy || m.state == stateSetupRest {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.width == 0 { return "Loading..." }

	var content string
	
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	studyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	restStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("154")).Bold(true)

	switch m.state {
	case stateSetupStudy:
		content = lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("POMODORO SETUP"),
			"\nMinutes of study:",
			m.input.View(),
			"\n(Enter to confirm)",
		)

	case stateSetupRest:
		content = lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("POMODORO SETUP"),
			"\nMinutes of rest:",
			m.input.View(),
			"\n(Enter to confirm)",
		)

	case stateRunning:
		minutes := m.timeLeft / 60
		seconds := m.timeLeft % 60
		timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)
		
		art := RenderTime(timeStr)
		
		var statusText string
		var styledArt string

		if m.isResting {
			statusText = " TIME TO REST "
			styledArt = restStyle.Render(art)
			statusText = restStyle.Render(statusText)
		} else {
			statusText = " FOCUS TIME "
			styledArt = studyStyle.Render(art)
			statusText = studyStyle.Render(statusText)
		}

		info := fmt.Sprintf("\n%s\n(ESC to reset, q to quit)", statusText)
		content = lipgloss.JoinVertical(lipgloss.Center, styledArt, info)
	}

	// fix to center better on odds number of spaces in vertical
	contentHeight := lipgloss.Height(content)
	if (m.height - contentHeight) % 2 != 0 {
		content = "\n" + content
	}

	// fix to center better on odds number of spaces in horizontal
	contentWidth := lipgloss.Width(content)
	paddingFixX := ""
	if (m.width - contentWidth) % 2 != 0 {
		paddingFixX = " " 
	}
	finalContent := lipgloss.JoinHorizontal(lipgloss.Center, paddingFixX, content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, finalContent)
}

func main() {
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
