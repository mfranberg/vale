package lint

import (
	"regexp"
	"strings"

	"github.com/niklasfasching/go-org/org"

	"github.com/errata-ai/vale/v3/internal/core"
)

var orgConverter = org.New()
var orgWriter = org.NewHTMLWriter()

var orgExample = "\n#+BEGIN_EXAMPLE\n$1\n#+END_EXAMPLE\n"

var reOrgAttribute = regexp.MustCompile(`(#(?:\+| )[^\s]+:.+)`)
var reOrgProps = regexp.MustCompile(`(:PROPERTIES:\n.+\n:END:)`)
var reOrgSrc = regexp.MustCompile(`(?i)#\+BEGIN_SRC .+`)

type ExtendedHTMLWriter struct {
	*org.HTMLWriter
}

func (w *ExtendedHTMLWriter) WriteComment(n org.Comment) {
	w.HTMLWriter.WriteString("<!-- ")
	w.HTMLWriter.WriteString(n.Content)
	w.HTMLWriter.WriteString(" -->\n")
}

func (l Linter) lintOrg(f *core.File) error {
	extendedWriter := &ExtendedHTMLWriter{orgWriter}
	orgWriter.ExtendingWriter = extendedWriter

	s := reOrgAttribute.ReplaceAllString(f.Content, "\n=$1=\n")
	s = reOrgProps.ReplaceAllString(s, orgExample)

	s, err := l.applyPatterns(s, orgExample, "=$1=", f.NormedPath)
	if err != nil {
		return err
	}

	// We don't want to find matches in `begin_src` lines.
	body := reOrgSrc.ReplaceAllStringFunc(f.Content, func(m string) string {
		return strings.Repeat("*", len(m))
	})

	doc := orgConverter.Parse(strings.NewReader(s), f.Path)
	// We don't want to introduce any *new* content into our HTML,
	// so we clear the outline.
	doc.Outline.Children = nil

	html, err := doc.Write(orgWriter)
	if err != nil {
		return err
	}

	f.Content = body
	return l.lintHTMLTokens(f, []byte(html), 0)
}
