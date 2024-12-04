package cmd

import tea "github.com/charmbracelet/bubbletea"

type Search struct {
	Query string
}

func SearchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		return Search{
			Query: query,
		}
	}
}

type OpenArticle struct {
	PageId int
	Name   string
}

func OpenArticleWithIdCmd(pageId int) tea.Cmd {
	return func() tea.Msg {
		return OpenArticle{
			PageId: pageId,
		}
	}
}
func OpenArticleWithNameCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return OpenArticle{
			Name: name,
		}
	}
}
