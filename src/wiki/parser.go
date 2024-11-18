package wiki

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type WikiTokenType int

const (
	TitleToken WikiTokenType = iota
	LinkToken
	BoldToken
	FileToken
)

type DefaultToken struct {
	tokenType WikiTokenType
	text      string
	content   string
	target    string
}

func (t DefaultToken) TokenType() WikiTokenType {
	return t.tokenType
}
func (t DefaultToken) Text() string {
	return t.text
}
func (t DefaultToken) Content() string {
	return t.content
}
func (t DefaultToken) Target() string {
	return t.target
}

func newTokenForType(tokenType WikiTokenType) func(text []string) DefaultToken {
	return func(text []string) DefaultToken {
		return DefaultToken{
			tokenType: tokenType,
			text:      text[0],
			content:   text[1],
			target:    "",
		}
	}
}
func newFileToken(text []string) DefaultToken {
	return DefaultToken{
		tokenType: FileToken,
		text:      text[0],
		content:   "",
		target:    "",
	}
}
func newLinkToken(text []string) DefaultToken {
	target := text[1]
	if target == "" {
		target = text[2]
	}
	return DefaultToken{
		tokenType: LinkToken,
		text:      text[0],
		content:   text[2],
		target:    target,
	}
}

type Parser struct {
	text   string
	regex  map[WikiTokenType]*regexp.Regexp
	styles map[WikiTokenType]*lipgloss.Style
	tokens []DefaultToken
}

func NewParser(text string, styles map[WikiTokenType]*lipgloss.Style) Parser {
	p := Parser{
		text:   text,
		styles: styles,
		regex: map[WikiTokenType]*regexp.Regexp{
			TitleToken: regexp.MustCompile(`==+([a-zA-Z ]+)==+`),
			LinkToken:  regexp.MustCompile(`\[{2}(?:([\w'\- ]+)\|)?([\w'\- ]+)\]{2}`),
			BoldToken:  regexp.MustCompile(`'''([a-zA-Z ]+)'''`),
			FileToken:  regexp.MustCompile(`\(?\[{2}File:[a-zA-Z ]+\.[a-z]{3}(?:\|[a-zA-Z ]+)*\]{2}\)?`),
		},
		tokens: []DefaultToken{},
	}
	p.Parse()

	return p
}

func (p *Parser) matchesForRegex(reg *regexp.Regexp, text string, fn func(s []string) DefaultToken) []DefaultToken {
	var tokens []DefaultToken = []DefaultToken{}
	matches := reg.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		tokens = append(tokens, fn(match))
	}

	return tokens
}
func (p *Parser) Parse() {
	tokens := []DefaultToken{}
	for tokenType, regex := range p.regex {
		var fn func(s []string) DefaultToken
		switch tokenType {
		case LinkToken:
			fn = newLinkToken
		case FileToken:
			fn = newFileToken
		default:
			fn = newTokenForType(tokenType)
		}
		newTokens := p.matchesForRegex(regex, p.text, fn)
		tokens = append(tokens, newTokens...)
	}
	p.tokens = tokens
}
func (p *Parser) Format() string {
	text := p.text
	log.Info("Tokens", "n", len(p.tokens))
	for _, token := range p.tokens {
		content := token.Content()
		if p.styles[token.TokenType()] != nil {
			content = p.styles[token.TokenType()].Render(content)
		}
		text = strings.ReplaceAll(
			text,
			token.Text(),
			content,
		)
	}
	return text
}
