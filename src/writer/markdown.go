package writer

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/skuid/changelog/src/changelog"
	"github.com/skuid/changelog/src/linkStyle"
)

func formatCommits(repo string, style linkStyle.Style, commits changelog.Commits) string {
	if len(commits) == 1 {
		return commits[0].Summary(repo, style)
	}
	var response []string
	for _, commit := range commits {
		response = append(response, fmt.Sprintf("  * %s", commit.Summary(repo, style)))
	}
	return fmt.Sprintf("\n%s", strings.Join(response, "\n"))
}

// MarkdownWriter writes a Markdown changelog
type MarkdownWriter struct {
	Writer io.Writer
}

const changeLog = `<a name="{{.version }}"></a>
##{{if .patchVersion}}#{{end}} {{.version}} ({{.date}}){{$style := .style}}{{$repo := .repo }}{{ $sectionMap := .sectionMap}}

{{- range $i, $section := .order}}
{{ $items  := index $sectionMap $section }}{{ $itemLen := len $items}}
{{- if gt $itemLen 0 }}
### {{ $section  }}
{{ range $component, $commits  := $items }}
* **{{ $component }}:** {{formatCommits $repo $style $commits}}{{end}}
{{- end}}
{{- end}}
`

// Generate writes a changelog to its embedded io.Writer
//
// yes the arguments are gross, I'm going to break out the fields needed off of
// changelog.changelog rather than pass around a huge object
func (m MarkdownWriter) Generate(c changelog.ChangeLog, style linkStyle.Style, sectionMap changelog.SectionMap) error {

	t, err := template.New("changeLog").Funcs(
		map[string]interface{}{
			"formatCommits": formatCommits,
		},
	).Parse(changeLog)

	if err != nil {
		return errors.WithStack(err)
	}

	data := map[string]interface{}{
		"version":      c.Version,
		"patchVersion": c.PatchVersion,
		"style":        style,
		"date":         time.Now().Format("2006-01-02"),
		"sectionMap":   sectionMap.Sections,
		"order":        sectionMap.Order(),
		"repo":         c.Repo,
	}

	return errors.WithStack(t.Execute(m.Writer, data))
}
