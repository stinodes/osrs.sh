package layout

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"osrs.sh/wiki/ssh/src/cmd"
	"osrs.sh/wiki/ssh/src/style"
	"osrs.sh/wiki/ssh/src/views/articlepane"
	"osrs.sh/wiki/ssh/src/views/homepane"
	"osrs.sh/wiki/ssh/src/views/searchpane"
	"osrs.sh/wiki/ssh/src/wiki"
)

type styles struct {
	main         lipgloss.Style
	contentFrame lipgloss.Style
}
type keys struct {
	Search key.Binding
	Enter  key.Binding
	Quit   key.Binding
	Cancel key.Binding
}

type contentPane int

const (
	homePane contentPane = iota
	searchPane
	articlePane
)

type Model struct {
	r             *lipgloss.Renderer
	styles        styles
	keys          keys
	width         int
	height        int
	title         string
	showSearchBar bool
	searchInput   textinput.Model
	queryResult   *wiki.QueryResult
	currentPane   contentPane
	panes         map[contentPane]tea.Model
}

var DefaultKeys = keys{
	Search: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "Open search"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Confirm"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
}

func New(r *lipgloss.Renderer) Model {
	ti := textinput.New()
	ti.Placeholder = "Search"
	ti.CharLimit = 64
	ti.Width = 20

	m := Model{
		r: r,
		styles: styles{
			contentFrame: r.NewStyle().
				Foreground(style.PrimaryForeground).
				Border(lipgloss.NormalBorder(), true).
				BorderForeground(style.BorderForeground).
				Padding(0, 1),
		},
		keys: DefaultKeys,

		width:  0,
		height: 0,

		title: "osrs.sh - ssh wiki",

		showSearchBar: false,
		searchInput:   ti,

		currentPane: homePane,
		panes:       map[contentPane]tea.Model{},
	}
	m.setPane(homePane)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) contentSize() (w int, h int) {
	return m.width - m.styles.contentFrame.GetHorizontalFrameSize(),
		m.height - m.styles.contentFrame.GetVerticalFrameSize() - lipgloss.Height(m.styles.contentFrame.Render("test"))

}
func (m *Model) resize(w int, h int) *Model {
	m.width = w
	m.height = h
	m.searchInput.Width = w / 2

	contentW, contentH := m.contentSize()
	if m.panes[searchPane] != nil {
		m.panes[searchPane].(*searchpane.Model).Resize(contentW, contentH)
	}

	return m
}

func (m *Model) initPane(pane contentPane) {
	w, h := m.contentSize()
	switch pane {
	case searchPane:
		search := searchpane.New(m.r, w, h)
		w, h := m.contentSize()
		search.Resize(w, h)
		m.panes[pane] = search
	case articlePane:
		article := articlepane.New(m.r, w, h)
		w, h := m.contentSize()
		article.Resize(w, h)
		m.panes[pane] = article
	default:
		m.panes[pane] = homepane.New()
	}
}
func (m *Model) setPane(pane contentPane) {
	if m.panes[pane] == nil {
		m.initPane(pane)
	}
	m.currentPane = pane
}
func (m *Model) confirmSearch(query string) tea.Cmd {
	return func() tea.Msg {
		result, err := wiki.Search(query)
		if err != nil {
			log.Error("Error searching wiki", "err", err)
			return nil
		}

		return result
	}
}
func (m *Model) fetchPage(msg cmd.OpenArticle) tea.Cmd {
	return func() tea.Msg {
		result, err := wiki.ParsePage(msg)
		if err != nil {
			log.Error("Error fetching page", "err", err)
			return nil
		}
		return result
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keys := m.keys
	var command tea.Cmd

	if m.searchInput.Focused() {
		m.searchInput, command = m.searchInput.Update(msg)
	}
	if m.panes[m.currentPane] != nil {
		m.panes[m.currentPane], command = m.panes[m.currentPane].Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.resize(msg.Width, msg.Height), nil
	case cmd.OpenArticle:
		m.setPane(articlePane)
		return m, tea.Batch(
			m.fetchPage(msg),
		)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Search):
			m.showSearchBar = true
			m.searchInput.Focus()
		case key.Matches(msg, keys.Enter):
			if m.searchInput.Focused() {
				m.setPane(searchPane)
				m.searchInput.Blur()
				return m, tea.Batch(
					m.confirmSearch(m.searchInput.Value()),
				)
			}
		case key.Matches(msg, keys.Cancel):
			m.showSearchBar = false
			m.searchInput.Blur()
		}

	}

	return m, command
}

func (m Model) View() string {

	topBarStyle := m.styles.contentFrame

	topBarContent := m.title
	if m.showSearchBar {
		topBarContent = m.searchInput.View()
	}

	topBar := topBarStyle.Render(
		lipgloss.PlaceHorizontal(
			m.width-topBarStyle.GetHorizontalFrameSize(),
			lipgloss.Left,
			topBarContent,
		),
	)

	bodyContent := ""
	if m.panes[m.currentPane] != nil {
		bodyContent = m.panes[m.currentPane].View()
	}
	body := m.styles.contentFrame.Render(
		lipgloss.Place(
			m.width-m.styles.contentFrame.GetHorizontalFrameSize(),
			m.height-m.styles.contentFrame.GetVerticalFrameSize()-lipgloss.Height(topBar),
			lipgloss.Left,
			lipgloss.Top,
			bodyContent,
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		topBar,
		body,
	)
}
