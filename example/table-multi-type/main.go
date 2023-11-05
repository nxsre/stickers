package main

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"log"
	"os"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocarina/gocsv"
	"github.com/nxsre/stickers/flexbox"
	"github.com/nxsre/stickers/table"
)

type SampleData struct {
	ID         int    `csv:"id"`
	FirstName  string `csv:"First Name"`
	LastName   string `csv:"Last Name"`
	Age        int    `csv:"Age"`
	Occupation string `csv:"Occupation"`
}

var sampleData []*SampleData

type model struct {
	// 表格
	table *table.Table
	// 表格状态栏
	infoBox *flexbox.FlexBox

	// 编辑器
	//editor *flexbox.FlexBox
	editor *inputsModel

	activeComponent string
	headers         []string
}

func main() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "foo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})

	// read in CSV data
	f, err := os.Open("../sample.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := gocsv.UnmarshalFile(f, &sampleData); err != nil {
		panic(err)
	}

	// 设置表头字段
	headers := []string{"id", "First Name", "Last Name", "Age", "Occupation"}
	// 设置每个字段屏幕宽度占比
	ratio := []int{1, 10, 10, 5, 10}
	// 设置每个字段的最小占位
	minSize := []int{5, 5, 5, 2, 5}

	// 设置每个字段的类型
	var s string
	var i int
	types := []any{i, s, s, i, s}

	m := model{
		table:   table.NewTable(0, 0, headers),
		infoBox: flexbox.New(1, 1),
		//editor:  flexbox.New(0, 10),
		editor: &inputsModel{},
		// 默认激活组件为 table
		activeComponent: "table",
		headers:         headers,
	}

	// 开启选择（原始数据表格最后增加一列选择状态）
	m.table.EnableSelect()

	// 设置状态栏
	infoRow := m.infoBox.NewRow()
	infoRow.AddCells(
		flexbox.NewCell(1, 1).
			SetID("info"),
		//flexbox.NewCell(1, 1).
		//	SetID("info").
		//	SetContent("222222"),
	)
	m.infoBox.AddRows([]*flexbox.Row{infoRow})

	infoboxStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7158e2")).
		Foreground(lipgloss.Color("#ffffff")).Align(lipgloss.Left).Height(1)
	m.infoBox.SetStyle(infoboxStyle)
	// set types
	_, err = m.table.SetTypes(types...)
	if err != nil {
		panic(err)
	}

	// setup dimensions
	m.table.SetRatio(ratio).SetMinWidth(minSize)
	// set style passing
	m.table.SetStylePassing(true)
	// add rows
	// with multi type table we have to convert our rows to []any first which is a bit of a pain
	var orderedRows [][]any
	for _, row := range sampleData {
		orderedRows = append(orderedRows, []any{
			row.ID, row.FirstName, row.LastName, row.Age, row.Occupation,
		})
	}
	m.table.MustAddRows(orderedRows)

	// 编辑器表单
	m.editor = initialInputsModel(&m)

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.activeComponent == "editor" {
		return m.editor.Update(msg)
	}

	// 如果活动组件为空，则设置活动组件为默认的 table
	if m.activeComponent == "" {
		m.activeComponent = "table"
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - m.infoBox.GetHeight() + 1)
		m.infoBox.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		// ctrl+e 编辑选中内容
		case "ctrl+e":
			if m.activeComponent == "table" {
				m.activeComponent = "editor"
			} else {
				m.activeComponent = "table"
			}
		}
		if m.activeComponent == "table" {
			switch msg.String() {
			case "ctrl+r":
			//	TODO: 刷新表格数据
			case "ctrl+c":
				return m, tea.Quit
			case "down":
				m.table.CursorDown()
			case "up":
				m.table.CursorUp()
			case "left":
				m.table.CursorLeft()
			case "right":
				m.table.CursorRight()
			case "ctrl+s":
				x, _ := m.table.GetCursorLocation()
				m.table.OrderByColumn(x)
			case "ctrl+a":
				if m.table.SelectedAll() {
					m.table.UnSelectAll()
				} else {
					m.table.SelectAll()
				}
				log.Println(m.table.GetSelectedRows())
			case " ":
				if m.table.Selected() {
					m.table.UnSelect()
				} else {
					m.table.Select()
				}
				log.Println(m.table.GetSelectedRows())
			// 回车选中当前行同时进入编辑页面，编辑所有选中内容
			case "enter":
				if m.activeComponent == "table" {
					m.activeComponent = "editor"
				} else {
					m.activeComponent = "table"
				}
				m.table.Select()
				return m.editor, tea.EnterAltScreen
				//selectedValue = m.table.GetCursorValue()
				//m.infoBox.GetRow(0).GetCell(1).SetContent("\nselected cell: " + selectedValue)
			case "backspace":
				m.filterWithStr(msg.String())
			// esc 键清空过滤条件
			case "esc":
				m.table.UnsetFilter()
			default:
				if len(msg.String()) == 1 {
					r := msg.Runes[0]
					if unicode.IsLetter(r) || unicode.IsDigit(r) {
						m.filterWithStr(msg.String())
					}
				}
			}
		}

	}

	// 设置状态栏
	m.infoBox.GetRow(0).GetCell(0).SetContent(fmt.Sprintf("已选择: %d", len(m.table.GetSelectedRows())))

	return m, nil
}

func (m *model) filterWithStr(key string) {
	i, s := m.table.GetFilter()
	x, _ := m.table.GetCursorLocation()
	if x != i && key != "backspace" {
		m.table.SetFilter(x, key)
		return
	}
	if key == "backspace" {
		if len(s) == 1 {
			m.table.UnsetFilter()
			return
		} else if len(s) > 1 {
			s = s[0 : len(s)-1]
		} else {
			return
		}
	} else {
		s = s + key
	}
	m.table.SetFilter(i, s)
}

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = buttonStyle.Copy().
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				MarginRight(2).
				Underline(true)
)

func (m *model) View() string {
	// 这里控制显示什么内容
	if m.activeComponent == "table" {
		return lipgloss.JoinVertical(lipgloss.Left, m.table.Render(), m.infoBox.Render())
	}
	if m.activeComponent == "editor" {
		//return lipgloss.JoinVertical(lipgloss.Left, m.editor.Render())
		return m.editor.View()
	}
	{
		//结合 "github.com/pterm/pterm" 包绘图
		//s := &strings.Builder{}
		//pterm.SetDefaultOutput(s)
		//pterm.DefaultBarChart.WithBars([]pterm.Bar{
		//	{Label: "A", Value: 10},
		//	{Label: "B", Value: 20},
		//	{Label: "C", Value: 30},
		//	{Label: "D", Value: 40},
		//	{Label: "E", Value: 50},
		//	{Label: "F", Value: 40},
		//	{Label: "G", Value: 30},
		//	{Label: "H", Value: 20},
		//	{Label: "I", Value: 10},
		//}).WithHeight(5).Render()
		//return s.String()
	}
	return ""
}
