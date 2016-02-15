// Html to md utils package.
package html2mdutil

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type replaceFn func(c string, n *html.Node) string

type htmlReplace struct {
	nodePtr  *html.Node
	replacer replaceFn
}

type parser struct {
	r     io.Reader
	w     io.Writer
	strip bool
}

// Reads html input from r.
// Writes md output to w.
// strip - option for stripping unknown html tags.
// Assumes that its input is in utf8.
// Return error on any write error encountered.
func Process(r io.Reader, w io.Writer, strip bool) error {
	p := &parser{
		r:     r,
		w:     w,
		strip: strip,
	}

	err := p.process()

	return err
}

func (p *parser) process() error {
	doc, err := html.Parse(p.r)

	if err != nil {
		return err
	}

	replacers := prepareChildrenReplacers(doc, p.strip)

	// run all replacers in back order
	for i := len(replacers) - 1; i >= 0; i-- {
		hr := replacers[i]
		l, t := getNodeSpaces(hr.nodePtr) // determine the spacing for current node
		replaced := hr.replacer(getNodeContent(hr.nodePtr, true), hr.nodePtr)
		hr.nodePtr.Data = l + replaced + t
	}

	out := getNodeContent(doc, true)
	out = normalizeMd(out)

	_, err = p.w.Write([]byte(out))

	return err
}

func getNodeContent(n *html.Node, normalize bool) string {
	var out string

	if n.Type == html.TextNode {
		if normalize {
			return normalizeText(n.Data)
		} else {
			return n.Data
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if len(c.Data) == 0 {
			continue
		}

		switch c.Type {
		case html.ElementNode:
			out += c.Data
		case html.TextNode:
			if normalize {
				out += normalizeText(c.Data)
			} else {
				out += c.Data
			}
		default:
			continue
		}
	}

	return out
}

func normalizeMd(s string) string {
	reMultiNl := regexp.MustCompile(`\n\n\n+`)
	reMultiNlW := regexp.MustCompile(`\n\s+\n`)
	reMultiLTR := regexp.MustCompile(`^[\t\r\n]+|[\t\r\n\s]+$`)

	out := reMultiLTR.ReplaceAllString(s, "")
	out = reMultiNlW.ReplaceAllString(out, "\n\n")
	out = reMultiNl.ReplaceAllString(out, "\n\n")

	return out
}

func normalizeText(s string) string {
	reMultiWs := regexp.MustCompile(`\s+`)
	out := reMultiWs.ReplaceAllString(s, " ")
	return strings.TrimSpace(out)
}

func prepareChildrenReplacers(n *html.Node, strip bool) []htmlReplace {
	t := []htmlReplace{}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			switch c.DataAtom {
			case atom.Br:
				t = append(t, htmlReplace{c, replaceTagsBr})
			case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
				t = append(t, htmlReplace{c, replaceTagsH})
			case atom.Hr:
				t = append(t, htmlReplace{c, replaceTagsHr})
			case atom.I, atom.Em:
				t = append(t, htmlReplace{c, replaceTagsEmI})
			case atom.B, atom.Strong:
				t = append(t, htmlReplace{c, replaceTagsStrongB})
			case atom.Code, atom.Pre:
				t = append(t, htmlReplace{c, replaceTagsCodePre})
			case atom.A:
				t = append(t, htmlReplace{c, replaceTagsA})
			case atom.Blockquote:
				t = append(t, htmlReplace{c, replaceTagsBlockquote})
			case atom.Li:
				t = append(t, htmlReplace{c, replaceTagsLi})
			case atom.Ul, atom.Ol:
				t = append(t, htmlReplace{c, replaceTagsUlOl})
			case atom.Address, atom.Article, atom.Aside, atom.Audio, atom.Body, atom.Canvas, atom.Center, atom.Dd,
				atom.Dir, atom.Div, atom.Dl, atom.Dt, atom.Fieldset, atom.Figcaption, atom.Figure, atom.Footer,
				atom.Form, atom.Frameset, atom.Header, atom.Hgroup, atom.Html, atom.Isindex, atom.Menu, atom.Nav,
				atom.Noframes, atom.Noscript, atom.Output, atom.P, atom.Span, atom.Section, atom.Table, atom.Tbody,
				atom.Td, atom.Tfoot, atom.Th, atom.Thead, atom.Tr:
				t = append(t, htmlReplace{c, replaceTagsBlocks})
			case 0: // unknown html atoms
				if !strip {
					t = append(t, htmlReplace{c, replaceTagsOther})
				} else {
					c.Data = ""
				}
			default: // ignore other known atoms
				c.Data = ""
			}
		} else if c.Type == html.TextNode {
			t = append(t, htmlReplace{c, replaceText})
		} else { // ignore all non text/element nodes
			c.Data = ""
		}

		if c.FirstChild != nil { // append children nodes
			t = append(t, prepareChildrenReplacers(c, strip)...)
		}
	}

	return t
}

func isBlock(a atom.Atom) bool {
	var blocks = map[atom.Atom]bool{
		atom.Address:    true,
		atom.Article:    true,
		atom.Aside:      true,
		atom.Audio:      true,
		atom.Blockquote: true,
		atom.Body:       true,
		atom.Canvas:     true,
		atom.Center:     true,
		atom.Dd:         true,
		atom.Dir:        true,
		atom.Div:        true,
		atom.Dl:         true,
		atom.Dt:         true,
		atom.Fieldset:   true,
		atom.Figcaption: true,
		atom.Figure:     true,
		atom.Footer:     true,
		atom.Form:       true,
		atom.Frameset:   true,
		atom.H1:         true,
		atom.H2:         true,
		atom.H3:         true,
		atom.H4:         true,
		atom.H5:         true,
		atom.H6:         true,
		atom.Header:     true,
		atom.Hgroup:     true,
		atom.Hr:         true,
		atom.Html:       true,
		atom.Isindex:    true,
		atom.Li:         true,
		atom.Menu:       true,
		atom.Nav:        true,
		atom.Noframes:   true,
		atom.Noscript:   true,
		atom.Ol:         true,
		atom.Output:     true,
		atom.P:          true,
		atom.Pre:        true,
		atom.Section:    true,
		atom.Table:      true,
		atom.Tbody:      true,
		atom.Td:         true,
		atom.Tfoot:      true,
		atom.Th:         true,
		atom.Thead:      true,
		atom.Tr:         true,
		atom.Ul:         true,
	}

	return blocks[a]
}

func checkNodeSiblingWs(n *html.Node, next bool) bool {
	var s *html.Node
	var content string
	var rn rune

	if next {
		for s = n.NextSibling; s != nil; s = s.NextSibling {
			if s.Type == html.ElementNode || s.Type == html.TextNode {
				break
			}
		}
	} else {
		for s = n.PrevSibling; s != nil; s = s.PrevSibling {
			if s.Type == html.ElementNode || s.Type == html.TextNode {
				break
			}
		}
	}

	if s == nil {
		return false
	}

	if isBlock(s.DataAtom) {
		return false
	}

	if s.Type == html.TextNode {
		content = s.Data
	} else {
		content = getNodeContent(s, true)
	}

	if len(content) == 0 {
		return false
	}

	if next {
		rn, _ = utf8.DecodeRuneInString(content[0:1])
	} else {
		rn, _ = utf8.DecodeLastRuneInString(content)
	}

	if unicode.IsSpace(rn) && len(content) > 1 {
		return true
	}

	return false
}

func getNodeSpaces(n *html.Node) (l string, t string) {
	c := getNodeContent(n, false)

	if len(c) == 0 {
		return
	}

	if isBlock(n.DataAtom) {
		return
	}

	sL, _ := utf8.DecodeRuneInString(c[0:1])
	sT, _ := utf8.DecodeLastRuneInString(c)

	if unicode.IsSpace(sL) || checkNodeSiblingWs(n, false) {
		l = " "
	}

	if unicode.IsSpace(sT) || checkNodeSiblingWs(n, true) {
		t = " "
	}

	return
}
