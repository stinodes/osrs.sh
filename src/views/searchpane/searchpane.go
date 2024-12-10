package searchpane

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"osrs.sh/wiki/ssh/src/cmd"
	"osrs.sh/wiki/ssh/src/style"
	"osrs.sh/wiki/ssh/src/wiki"
)

type item struct {
	title string
	desc  string
	id    int
}

func (i item) ID() int {
	return i.id
}
func (i item) Title() string {
	return i.title
}
func (i item) Description() string {
	return i.desc
}
func (i item) FilterValue() string {
	return ""
}

type Model struct {
	r       *lipgloss.Renderer
	results *wiki.QueryResult
	list    list.Model
}

func itemStyles(renderer *lipgloss.Renderer) (s list.DefaultItemStyles) {
	s.NormalTitle = renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})

	s.SelectedTitle = renderer.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(style.AccentForeground).
		Foreground(style.AccentForeground).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedTitle.
		Foreground(style.AccentForeground)

	s.DimmedTitle = renderer.NewStyle().
		Foreground(style.PrimaryForeground).
		Padding(0, 0, 0, 2)

	s.DimmedDesc = s.DimmedTitle.
		Foreground(style.DimmedForeground)

	s.FilterMatch = renderer.NewStyle().Underline(true)

	return s
}
func New(renderer *lipgloss.Renderer, w int, h int) Model {
	items := []list.Item{}
	delegate := list.NewDefaultDelegate()
	delegate.Styles = itemStyles(renderer)

	list := list.New(items, delegate, w, h)
	list.SetShowTitle(false)
	list.SetShowStatusBar(false)
	list.SetShowPagination(false)
	list.SetShowFilter(false)
	list.SetShowHelp(false)
	return Model{
		r:       renderer,
		results: nil,
		list:    list,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SelectedResult() int {
	return m.list.SelectedItem().(item).id
}

func (m *Model) Resize(width, height int) {
	m.list.SetSize(width, height)

}
func (m *Model) setResults(results *wiki.QueryResult) {
	pages := results.Query.Search
	var items []list.Item = []list.Item{}
	for _, result := range pages {
		items = append(items, item{
			title: result.Title,
			desc:  result.Snippet,
			id:    result.PageID,
		})

	}

	m.results = results
	m.list.SetItems(items)

}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var command tea.Cmd
	switch msg := msg.(type) {
	case *wiki.QueryResult:
		m.results = msg
		m.setResults(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, cmd.OpenArticleWithIdCmd(m.SelectedResult())
		}
	}

	m.list, command = m.list.Update(msg)

	return m, command
}

func (m Model) View() string {
	return m.list.View()
}
