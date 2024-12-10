package wiki

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type WikiTokenType int

const (
	TitleToken WikiTokenType = iota
	LinkToken
	BoldToken
	FileToken
	RedirectToken
)

type DefaultToken struct {
	tokenType WikiTokenType
	text      string
	content   string
	target    string
	id        int
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
func (t DefaultToken) Id() int {
	return t.id
}
func (t DefaultToken) Placeholder() string {
	if t.id == -1 {
		return t.content
	}
	str := fmt.Sprintf("$%-*s_", len(t.Content())-1, strconv.Itoa(t.id))
	str = strings.ReplaceAll(str, " ", "_")
	return str
}

type Parser struct {
	idIncr int
	text   string
	regex  map[WikiTokenType]*regexp.Regexp
	tokens []DefaultToken
}

func NewParser(text string, styles map[WikiTokenType]*lipgloss.Style) Parser {
	p := Parser{
		text: text,
		regex: map[WikiTokenType]*regexp.Regexp{
			TitleToken:    regexp.MustCompile(`==+([a-zA-Z \-\':\(\)]+)==+`),
			LinkToken:     regexp.MustCompile(`\[{2}(?:([\w'\+\- \(\):#]+)\|)?([\w'\- \+:\(\)]+)\]{2}`),
			BoldToken:     regexp.MustCompile(`'''([a-zA-Z ]+)'''`),
			FileToken:     regexp.MustCompile(`\(?\[{2}File:[a-zA-Z ]+\.[a-z]{3}(?:\|[a-zA-Z ]+)*\]{2}\)?`),
			RedirectToken: regexp.MustCompile(`\{\{redirect\|[\w'\-\| \(\)]+\}\}`),
		},
		tokens: []DefaultToken{},
	}
	p.Parse()

	return p
}

func (p *Parser) newTokenForType(tokenType WikiTokenType) func(text []string) DefaultToken {
	return func(text []string) DefaultToken {
		id := p.idIncr
		p.idIncr++
		return DefaultToken{
			tokenType: tokenType,
			text:      text[0],
			content:   text[1],
			target:    "",
			id:        id,
		}
	}
}
func (p *Parser) newHiddenTokenForType(tokenType WikiTokenType) func(text []string) DefaultToken {
	return func(text []string) DefaultToken {
		return DefaultToken{
			tokenType: tokenType,
			text:      text[0],
			content:   "",
			target:    "",
			id:        -100,
		}
	}
}
func (p *Parser) newLinkToken(text []string) DefaultToken {
	target := text[1]
	if target == "" {
		target = text[2]
	}
	id := p.idIncr
	p.idIncr++
	return DefaultToken{
		tokenType: LinkToken,
		text:      text[0],
		content:   text[2],
		target:    target,
		id:        id,
	}
}

func (p *Parser) matchesForRegex(reg *regexp.Regexp, text string, fn func(s []string) DefaultToken) []DefaultToken {
	var tokens []DefaultToken = []DefaultToken{}
	matches := reg.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		tokens = append(tokens, fn(match))
	}

	return tokens
}
func (p *Parser) Tokens() []DefaultToken {
	return p.tokens
}
func (p *Parser) TokenById(id int) *DefaultToken {
	for _, token := range p.tokens {
		if token.Id() == id {
			return &token
		}
	}
	return nil
}
func (p *Parser) TokenByPlaceholder(str string) *DefaultToken {
	for _, token := range p.tokens {
		if token.Placeholder() == str {
			return &token
		}
	}
	return nil
}

func (p *Parser) Parse() {
	tokens := []DefaultToken{}
	for tokenType, regex := range p.regex {
		var fn func(s []string) DefaultToken
		switch tokenType {
		case LinkToken:
			fn = p.newLinkToken
		case FileToken, RedirectToken:
			fn = p.newHiddenTokenForType(tokenType)
		default:
			fn = p.newTokenForType(tokenType)
		}

		p.text = regex.ReplaceAllStringFunc(p.text, func(s string) string {
			match := regex.FindAllStringSubmatch(s, -1)[0]
			if match == nil {
				return s
			}
			token := fn(match)
			tokens = append(tokens, token)
			return token.Placeholder()
		})
	}
	p.tokens = tokens
}
func (p *Parser) Text() string {
	return p.text
}
