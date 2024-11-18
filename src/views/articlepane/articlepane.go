package articlepane

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"osrs.sh/wiki/ssh/src/style"
	"osrs.sh/wiki/ssh/src/wiki"
)

type styles struct {
	body  lipgloss.Style
	title lipgloss.Style
	bold  lipgloss.Style
	link  lipgloss.Style
}
type Model struct {
	r        *lipgloss.Renderer
	width    int
	height   int
	styles   styles
	page     *wiki.Page
	parser   wiki.Parser
	viewport viewport.Model
}

func New(renderer *lipgloss.Renderer, w int, h int) Model {
	viewport := viewport.New(w, h)
	return Model{
		r: renderer,
		styles: styles{
			body: renderer.
				NewStyle().
				Foreground(style.PrimaryForeground),
			title: renderer.
				NewStyle().
				Foreground(style.AccentForeground).
				Bold(true).
				Underline(true),
			bold: renderer.
				NewStyle().
				Bold(true),
			link: renderer.
				NewStyle().
				Foreground(style.LinkForeground),
		},
		page:     nil,
		viewport: viewport,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.YPosition = 0
	m.viewport.Width = width
	m.viewport.Height = height
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case *wiki.Page:
		m.page = msg
		m.parser = wiki.NewParser(
			msg.WikiText,
			map[wiki.WikiTokenType]*lipgloss.Style{
				wiki.TitleToken: &m.styles.title,
				wiki.BoldToken:  &m.styles.bold,
				wiki.LinkToken:  &m.styles.link,
			},
		)
		m.viewport.SetContent(m.parser.Format())
	}

	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}
func (m Model) removeImages(text string) string {
	regex := regexp.MustCompile(`(\(?\[{2}File:[a-zA-Z ]+\.[a-z]{3}(?:\|[a-zA-Z ]+)*\]{2}\)?)`)
	return regex.ReplaceAllString(text, "")
}
func (m Model) parseWikiLinks(text string) string {
	regex := regexp.MustCompile(`(\[{2}[a-zA-Z ]+\]{2})`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		s = strings.ReplaceAll(s, "[", "")
		s = strings.ReplaceAll(s, "]", "")
		return m.styles.link.Render(s)
	})
}
func (m Model) parseWikiBoldText(text string) string {
	regex := regexp.MustCompile(`('''[a-zA-Z ]+''')`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		return m.styles.bold.Render(strings.ReplaceAll(s, "'", ""))
	})
}
func (m Model) parseWikiHeaders(text string) string {
	regex := regexp.MustCompile(`(==+[a-zA-Z ]+==+)`)
	return regex.ReplaceAllStringFunc(text, func(s string) string {
		return m.styles.title.Render(strings.ReplaceAll(s, "=", ""))
	})

}
func (m *Model) parseWikiText(text string) string {
	text = m.parseWikiHeaders(text)
	text = m.parseWikiBoldText(text)
	text = m.parseWikiLinks(text)
	text = m.removeImages(text)
	return text
}

func (m Model) View() string {
	if m.page == nil {
		return "Loading..."
	}

	return m.viewport.View()
}
