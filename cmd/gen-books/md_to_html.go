package main

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/kjk/u"
	"github.com/microcosm-cc/bluemonday"
)

var (
	htmlFormatter  *html.Formatter
	highlightStyle *chroma.Style
)

func init() {
	htmlFormatter = html.New(html.WithClasses(), html.TabWidth(2))
	u.PanicIf(htmlFormatter == nil, "couldn't create html formatter")
	styleName := "monokailight"
	highlightStyle = styles.Get(styleName)
	u.PanicIf(highlightStyle == nil, "didn't find style '%s'", styleName)
}

// based on https://github.com/alecthomas/chroma/blob/master/quick/quick.go
func htmlHighlight(w io.Writer, source, lang, defaultLang string) error {
	if lang == "" {
		lang = defaultLang
	}
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return htmlFormatter.Format(w, highlightStyle, it)
}

func isArticleOrChapterLink(s string) bool {
	return strings.HasPrefix(s, "a-") || strings.HasPrefix(s, "ch-")
}

var didPrint = false

func printKnownURLS(a []string) {
	if didPrint {
		return
	}
	didPrint = true
	fmt.Printf("%d known urls\n", len(a))
	for _, s := range a {
		fmt.Printf("%s\n", s)
	}
}

// turn partial url like "a-20381" into a full url like "a-20381-installing"
func fixupURL(uri string, knownURLS []string) string {
	if !isArticleOrChapterLink(uri) {
		return uri
	}
	for _, known := range knownURLS {
		if uri == known {
			return uri
		}
		if strings.HasPrefix(known, uri) {
			//fmt.Printf("fixupURL: %s => %s\n", uri, known)
			return known
		}
	}
	fmt.Printf("fixupURL: didn't fix up: %s\n", uri)
	//printKnownURLS(knownURLS)
	return uri
}

// d might be empty, just lang or lang:github link
func parseCodeBlockInfo(d []byte) (string, string) {
	//fmt.Printf("d: %s\n", string(d))
	if len(d) == 0 {
		return "", ""
	}
	s := strings.TrimSpace(string(d))
	if len(s) == 0 {
		return "", ""
	}
	parts := strings.Split(s, "|")
	if len(parts) == 1 {
		// only lang
		return parts[0], ""
	}
	u.PanicIf(len(parts) != 2, "len(parts) is %d, expected 2, d: '%s'", len(parts), string(d))
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// gross hack. we need change html generated by chroma
func fixupHTMLCodeBlock(s string, gitHubLoc string) string {
	if gitHubLoc == "" {
		return s
	}
	// gitHubLoc is sth. like github.com/essentialbooks/books/books/go/main.go
	fileName := path.Base(gitHubLoc)
	html := fmt.Sprintf(`<div class="code-box">
%s
<div class="code-box-nav">
	<div class="code-box-file-name">
		<a href="%s" target="_blank">%s</a>
	</div>
</div>`, s, gitHubLoc, fileName)
	return html
}

// knownUrls is a list of chapter/article urls in the form "a-20381-installing", "ch-198-getting-started"
func makeRenderHookCodeBlock(defaultLang string, book *Book) mdhtml.RenderNodeFunc {
	return func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {

		if codeBlock, ok := node.(*ast.CodeBlock); ok {
			lang, gitHubLoc := parseCodeBlockInfo(codeBlock.Info)
			//fmt.Printf("lang: %s, gitHubLoc: %s\n", lang, gitHubLoc)
			if false {
				fmt.Printf("lang: '%s', code: %s\n", lang, string(codeBlock.Literal[:16]))
				io.WriteString(w, "\n"+`<pre class="chroma"><code>`)
				mdhtml.EscapeHTML(w, codeBlock.Literal)
				io.WriteString(w, "</code></pre>\n")
			} else {
				//fmt.Printf("\n----\n%s\n----\n", string(codeBlock.Literal))
				var tmp bytes.Buffer
				htmlHighlight(&tmp, string(codeBlock.Literal), lang, defaultLang)
				d := tmp.Bytes()
				s := fixupHTMLCodeBlock(string(d), gitHubLoc)
				io.WriteString(w, s)
			}
			return ast.GoToNext, true
		} else if link, ok := node.(*ast.Link); ok {
			// fix up the url if it's a prefix of known url and let original code to render it
			dest := string(link.Destination)
			link.Destination = []byte(fixupURL(dest, book.knownUrls))
			return ast.GoToNext, false
		} else {
			return ast.GoToNext, false
		}
	}
}

func markdownToUnsafeHTML(md []byte, defaultLang string, book *Book) []byte {
	extensions := parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings |
		parser.NoEmptyLineBeforeBlock
	parser := parser.NewWithExtensions(extensions)

	htmlFlags := mdhtml.Smartypants |
		mdhtml.SmartypantsFractions |
		mdhtml.SmartypantsDashes |
		mdhtml.SmartypantsLatexDashes
	htmlOpts := mdhtml.RendererOptions{
		Flags:          htmlFlags,
		RenderNodeHook: makeRenderHookCodeBlock(defaultLang, book),
	}
	renderer := mdhtml.NewRenderer(htmlOpts)
	return markdown.ToHTML(md, parser, renderer)
}

func markdownToHTML(d []byte, defaultLang string, book *Book) string {
	unsafe := markdownToUnsafeHTML(d, defaultLang, book)
	policy := bluemonday.UGCPolicy()
	policy.AllowStyling()
	policy.RequireNoFollowOnFullyQualifiedLinks(false)
	policy.RequireNoFollowOnLinks(false)
	policy.AllowAttrs("target").OnElements("a")
	res := policy.SanitizeBytes(unsafe)
	return string(res)
}
