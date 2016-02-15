package html2mdutil

import (
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"regexp"
	"strings"
	"unicode"
)

func replaceTagsBlocks(c string, n *html.Node) string {
	if len(c) == 0 {
		return ""
	}

	return fmt.Sprintf("\n\n%s\n\n", c)
}

func replaceTagsBr(c string, n *html.Node) string {
	return "  \n"
}

func replaceTagsH(c string, n *html.Node) string {
	prefix := "#"
	count := 1

	switch n.DataAtom {
	case atom.H2:
		count = 2
	case atom.H3:
		count = 3
	case atom.H4:
		count = 4
	case atom.H5:
		count = 5
	case atom.H6:
		count = 6
	}

	return fmt.Sprintf("\n\n%s %s\n\n", strings.Repeat(prefix, count), c)
}

func replaceTagsHr(c string, n *html.Node) string {
	return "\n\n* * *\n\n"
}

func replaceTagsEmI(c string, n *html.Node) string {
	return fmt.Sprintf("_%s_", c)
}

func replaceTagsStrongB(c string, n *html.Node) string {
	return fmt.Sprintf("**%s**", c)
}

func replaceTagsCodePre(c string, n *html.Node) string {
	if n.DataAtom == atom.Pre && n.FirstChild != nil &&
		n.FirstChild.DataAtom == atom.Code && n.FirstChild.FirstChild != nil {
		return fmt.Sprintf("\n\n    %s\n\n", strings.Replace(n.FirstChild.FirstChild.Data, "\n", "\n    ", -1))
	}

	if n.DataAtom == atom.Code {
		siblings := n.PrevSibling == nil && n.NextSibling == nil
		code := n.Parent != nil && n.Parent.DataAtom == atom.Pre && siblings

		if !code && len(c) > 0 {
			return fmt.Sprintf("`%s`", c)
		}
	}

	return c
}

func replaceTagsA(c string, n *html.Node) string {
	var href, title string

	for _, a := range n.Attr {
		if a.Key == "href" {
			href = a.Val
			continue
		}

		if a.Key == "title" {
			title = " " + a.Val
			continue
		}
	}

	if len(href) != 0 {
		return fmt.Sprintf("[%s](%s%s)", c, href, title)
	} else {
		return href
	}
}

func replaceTagsImg(c string, n *html.Node) string {
	var src, alt, title string

	for _, a := range n.Attr {
		if a.Key == "src" {
			src = a.Val
			continue
		}

		if a.Key == "alt" {
			alt = a.Val
			continue
		}

		if a.Key == "title" {
			title = " " + a.Val
			continue
		}
	}

	if len(src) != 0 {
		return fmt.Sprintf("![%s](%s%s)", alt, src, title)
	} else {
		return src
	}
}

func replaceTagsBlockquote(c string, n *html.Node) string {
	reMultiNl := regexp.MustCompile(`\n\n\n+`)
	reBlockquote := regexp.MustCompile(`(?m)^`)

	data := strings.TrimSpace(c)
	data = reMultiNl.ReplaceAllString(data, "\n\n")
	data = reBlockquote.ReplaceAllString(data, "> ")

	return fmt.Sprintf("\n\n%s\n\n", data)
}

func replaceTagsLi(c string, n *html.Node) string {
	data := strings.TrimLeftFunc(c, unicode.IsSpace)
	data = strings.Replace(data, "\n", "\n    ", -1)

	pref := "*   "

	p := n.Parent

	if p != nil {
		i := 0
		for c := p.FirstChild; c != nil; c = c.NextSibling {
			if c.DataAtom == atom.Li {
				i++
				if c == n {
					break
				}
			}
		}

		if p.DataAtom == atom.Ol {
			pref = fmt.Sprintf("%d.   ", i)
		}
	}

	return pref + data
}

func replaceTagsUlOl(c string, n *html.Node) string {
	l := []string{}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == atom.Li {
			l = append(l, c.Data)
		}
	}

	if n.Parent != nil && n.Parent.DataAtom == atom.Li {
		return "\n" + strings.Join(l, "\n")
	} else {
		return fmt.Sprintf("\n\n%s\n\n", strings.Join(l, "\n"))
	}
}

func replaceTagsOther(c string, n *html.Node) string {
	return fmt.Sprintf("<%s>%s</%s>", n.Data, c, n.Data)
}

func replaceText(c string, n *html.Node) string {
	return n.Data
}
