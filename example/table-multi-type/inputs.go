package main

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	saveButton   = "[ 保存 ]"
	cancelButton = "[ 取消 ]"
)

type inputsModel struct {
	model      *model
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
}

func initialInputsModel(main *model) *inputsModel {
	m := inputsModel{
		model:  main,
		inputs: make([]textinput.Model, 3),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Prompt = "账号: "
			t.Placeholder = "Nickname"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Prompt = "邮箱: "
			t.Placeholder = "Email"
			t.CharLimit = 64
		case 2:
			t.Prompt = "密码: "
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return &m
}

func (m inputsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.model.activeComponent = ""
			return m.model, tea.ClearScreen
		case "ctrl+c":
			return m, tea.Quit
		// 设置光标模式
		case "ctrl+r":
			//m.cursorMode++
			//if m.cursorMode > cursor.CursorHide {
			//	m.cursorMode = cursor.CursorBlink
			//}
			//cmds := make([]tea.Cmd, len(m.inputs))
			//for i := range m.inputs {
			//	cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			//}
			//return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down", "left", "right":
			s := msg.String()

			if s == "enter" {
				// 如果当前焦点是最后一个组件
				if m.focusIndex == len(m.inputs) {
					for k, v := range m.inputs {
						log.Println(k, v.Value())
					}
					log.Println("保存")
					m.model.activeComponent = ""
					return m.model, tea.ClearScreen
				}

				if m.focusIndex == len(m.inputs)+1 {
					for k, v := range m.inputs {
						log.Println(k, v.Value())
					}
					log.Println("取消")
					m.model.activeComponent = ""
					return m.model, tea.ClearScreen
				}
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" || s == "left" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			// len(m.inputs)+1 为 inputs 数量 + 两个按钮
			if m.focusIndex > len(m.inputs)+1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) + 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *inputsModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputsModel with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m inputsModel) View() string {
	var b strings.Builder

	for _, v := range m.model.table.GetSelectedRows() {
		fmt.Fprintf(&b, "%s\n", v[1])
	}

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	fmt.Fprintf(&b, "\n")
	buttons := []string{blurredStyle.Render(saveButton), blurredStyle.Render(cancelButton)}

	if m.focusIndex == len(m.inputs)+1 {
		buttons = []string{blurredStyle.Copy().Render(saveButton), focusedStyle.Copy().Render(cancelButton)}
	}

	if m.focusIndex == len(m.inputs) {
		buttons = []string{focusedStyle.Copy().Render(saveButton), blurredStyle.Copy().Render(cancelButton)}
	}

	for _, button := range buttons {
		fmt.Fprintf(&b, "\t%s\t", button)
	}
	fmt.Fprintf(&b, "\n")

	// 显示当前光标模式
	//b.WriteString(helpStyle.Render("cursor mode is "))
	//b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	//b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
