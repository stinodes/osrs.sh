package articlepane

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"osrs.sh/wiki/ssh/src/cmd"
	"osrs.sh/wiki/ssh/src/style"
	"osrs.sh/wiki/ssh/src/utils"
	"osrs.sh/wiki/ssh/src/wiki"
)

type ActionInput string

const (
	Up       ActionInput = "k"
	Down     ActionInput = "j"
	Left     ActionInput = "h"
	Right    ActionInput = "h"
	Top      ActionInput = "gg"
	Bottom   ActionInput = "G"
	NextLink ActionInput = "l"
	PrevLink ActionInput = "h"
	Confirm  ActionInput = "enter"
)

type styles struct {
	body     lipgloss.Style
	title    lipgloss.Style
	bold     lipgloss.Style
	link     lipgloss.Style
	selected lipgloss.Style
	content  lipgloss.Style
	lineCol  lipgloss.Style
}
type Model struct {
	r           *lipgloss.Renderer
	width       int
	height      int
	styles      styles
	tokenStyles map[wiki.WikiTokenType]*lipgloss.Style

	page   *wiki.Page
	parser wiki.Parser

	buffer        []string
	scrollPos     int
	content       string
	selectedToken int
}

var numberRegex = regexp.MustCompile(`(\d+)`)

const numberWidth = 5

func New(renderer *lipgloss.Renderer, w int, h int) Model {
	s := styles{
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
		selected: renderer.
			NewStyle().
			Background(style.SelectedBackground),
		content: renderer.NewStyle().Width(w - numberWidth),
		lineCol: renderer.NewStyle().
			MaxWidth(numberWidth).
			Foreground(style.SubtleForeground),
	}
	return Model{
		r:      renderer,
		styles: s,
		tokenStyles: map[wiki.WikiTokenType]*lipgloss.Style{
			wiki.TitleToken: &s.title,
			wiki.LinkToken:  &s.link,
			wiki.BoldToken:  &s.bold,
		},
		page:          nil,
		parser:        wiki.Parser{},
		buffer:        []string{},
		scrollPos:     0,
		content:       "",
		selectedToken: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) SetPage(page *wiki.Page) Model {
	log.Info("SetPage", "page", page)
	m.page = page
	m.parser = wiki.NewParser(
		page.WikiText,
		map[wiki.WikiTokenType]*lipgloss.Style{
			wiki.TitleToken: &m.styles.title,
			wiki.BoldToken:  &m.styles.bold,
			wiki.LinkToken:  &m.styles.link,
		},
	)
	return m
}
func (m *Model) Resize(width, height int) {
	m.width = width
	m.height = height
	m.styles.content.Width(width - numberWidth)
	m.scrollPos = m.constrainScrollPos(m.scrollPos)
}

func (m *Model) contentLength() int {
	return len(strings.Split(
		m.styles.content.Render(m.parser.Text()),
		"\n",
	),
	)
}

func (m Model) constrainScrollPos(pos int) int {
	if pos < 0 {
		pos = 0
	}
	if pos > m.contentLength() {
		pos = m.contentLength() - 1
	}
	return pos
}
func (m *Model) Scroll(delta int) {
	if delta == 0 {
		delta = 1
	}
	m.scrollPos = m.constrainScrollPos(m.scrollPos + delta)
}
func (m *Model) ScrollUp(delta int) {
	if delta == 0 {
		delta = 1
	}
	m.Scroll(-delta)
}
func (m *Model) ScrollTo(line int) {
	if line == 0 {
		line = m.contentLength() - m.height/2
	}
	m.scrollPos = m.constrainScrollPos(line)
}
func (m *Model) ScrollToTop(_ int) {
	m.ScrollTo(1)
}
func (m *Model) ScrollToBottom(_ int) {
	m.ScrollTo(m.contentLength() - m.height/2)
}
func (m *Model) ScrollToToken(token wiki.DefaultToken) {
	line := m.lineFor(token.Placeholder())
	if !m.isInView(token.Placeholder()) {
		offset := -5

		if line > m.scrollPos {
			offset = -m.height - offset
		}

		m.ScrollTo(line + offset)
	}
}
func (m *Model) SelectToken(token wiki.DefaultToken) {
	m.selectedToken = token.Id()
	m.ScrollToToken(token)
}
func (m *Model) NextLink(id int) {
	cur := m.parser.TokenById(m.selectedToken)
	linkVisible := cur != nil && m.isInView(cur.Placeholder())

	if !linkVisible {
		viewableContent := m.viewableContent()
		for i := 0; i < len(m.parser.Tokens()); i++ {
			token := m.parser.TokenById(i)
			if token != nil &&
				token.TokenType() == wiki.LinkToken &&
				strings.Contains(viewableContent, token.Placeholder()) {
				m.SelectToken(*token)
				break
			}
		}

		return
	}

	tokens := m.parser.Tokens()
	for i := cur.Id() + 1; i < len(m.parser.Tokens()); i++ {
		if tokens[i].TokenType() == wiki.LinkToken {
			m.SelectToken(tokens[i])
			break
		}
	}
}
func (m *Model) PrevLink(id int) {
	cur := m.parser.TokenById(m.selectedToken)
	linkVisible := cur != nil && m.isInView(cur.Placeholder())

	if !linkVisible {
		viewableContent := m.viewableContent()
		for i := len(m.parser.Tokens()) - 1; i >= 0; i-- {
			token := m.parser.TokenById(i)
			if token != nil &&
				token.TokenType() == wiki.LinkToken &&
				strings.Contains(viewableContent, token.Placeholder()) {
				m.SelectToken(*token)
				break
			}
		}

		return
	}

	tokens := m.parser.Tokens()
	for i := cur.Id() - 1; i >= 0; i++ {
		if tokens[i].TokenType() == wiki.LinkToken {
			m.SelectToken(tokens[i])
			break
		}
	}
}

func matches(input string, match ActionInput) bool {
	return strings.HasPrefix(input, string(match))
}
func (m *Model) Push(input string) tea.Cmd {
	m.buffer = append([]string{input}, m.buffer...)
	bufferString := strings.Join(m.buffer, "")
	cur := bufferString

	var action func(n int)

	log.Info("Push", "input", input, "buffer", m.buffer)

	switch {
	case matches(cur, Up):
		action = m.ScrollUp
	case matches(cur, Down):
		action = m.Scroll
	case matches(cur, Top):
		action = m.ScrollToTop
	case matches(cur, Bottom):
		action = m.ScrollTo
	case matches(cur, NextLink):
		action = m.NextLink
	case matches(cur, PrevLink):
		action = m.PrevLink
	case matches(cur, Confirm):
		token := m.parser.TokenById(m.selectedToken)
		if token != nil {
			return cmd.OpenArticleWithNameCmd(token.Target())
		}
	}

	if action != nil {
		numberMatches := numberRegex.FindAllStringSubmatch(cur, -1)
		num := 0
		if len(numberMatches) > 0 {
			numberStr := numberMatches[0][len(numberMatches)-1]
			num, _ = strconv.Atoi(utils.ReverseString(numberStr))
		}
		action(num)
		m.buffer = []string{}
	}

	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.Push(msg.String())
	}

	return m, cmd
}

func (m Model) lineCol() string {
	lines := ""
	for i := m.scrollPos; i < m.scrollPos+m.height; i++ {
		lines += fmt.Sprintf("%-*s\n", numberWidth, strconv.Itoa(i))
	}
	return m.styles.lineCol.
		MaxHeight(m.height).
		Render(lines)
}
func (m Model) lineFor(text string) int {
	before := strings.Split(
		m.renderedContent(),
		text,
	)[0]

	return len(strings.Split(before, "\n")) - 1
}
func (m Model) isInView(text string) bool {
	return strings.Contains(m.viewableContent(), text)
}
func (m Model) viewableContent() string {
	lines := strings.Split(
		m.renderedContent(), "\n")
	if len(lines) <= 1 {
		return ""
	}
	start := m.scrollPos
	end := m.constrainScrollPos(m.scrollPos + m.height)
	return strings.Join(lines[start:end], "\n")
}
func (m Model) renderedContent() string {
	return m.styles.content.Render(m.parser.Text())
}

func (m Model) View() string {
	if m.page == nil {
		return "Loading..."
	}

	c := m.viewableContent()

	for _, token := range m.parser.Tokens() {
		style := m.tokenStyles[token.TokenType()]
		tokenContent := token.Content()
		if style != nil {
			tokenContent = style.Render(token.Content())
			if m.selectedToken == token.Id() {
				tokenContent = m.styles.selected.Render(tokenContent)
			}
		}
		c = strings.Replace(c, token.Placeholder(), tokenContent, 1)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.lineCol(),
		lipgloss.NewStyle().Render(c),
	)
}
