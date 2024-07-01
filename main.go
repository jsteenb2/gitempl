package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"unicode"
	
	"github.com/conventionalcommit/parser"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

func main() {
	if err := newCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newCmd() *cobra.Command {
	c := new(cli)
	return c.newCmd()
}

type cli struct {
	dir  string
	tmpl string
}

func (c *cli) newCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:          "gitempl [--dir|-d] $OPTIONAL_FILENAME",
		Short:        "a simple doc generator that is conventional commit aware",
		RunE:         c.runE,
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		Example: `  # execute from root of git repo with inline template writes to stdout
> gitempl <<EOF
{{ range .Commits }}
    {{ .Author }} commited by
    {{ .Hash }} commit hash
    {{ .Message }} commit message full
{{ end }}
EOF

# execute with a template file and write to file
> gitempl -t $FILE_TEMPLATE $FILE

# execute with a git repo in arbitrary directory (not in PWD) with template file
# writes to stdout
> gitempl -d $PATH_TO_GIT_REPO -t $FILE_TEMPLATE
`,
	}
	
	cmd.Flags().StringVarP(&c.dir, "dir", "d", ".", "directory of git repo")
	cmd.Flags().StringVarP(&c.tmpl, "template", "t", "", "template to execute; defaults to stdin")
	
	return &cmd
}

func (c *cli) runE(cmd *cobra.Command, args []string) error {
	r, err := git.PlainOpen(c.dir)
	if err != nil {
		return err
	}
	
	commits, err := parseGitTemplVars(r)
	if err != nil {
		return err
	}
	
	t, err := c.template(cmd.InOrStdin())
	if err != nil {
		return err
	}
	
	var file string
	if len(args) > 0 {
		file = args[0]
	}
	
	w, closeFn, err := c.output(file, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	defer closeFn() // in case of early exit
	
	err = t.Execute(w, input{Commits: commits})
	if err != nil {
		return err
	}
	
	return closeFn()
}

func (c *cli) template(stdin io.Reader) (*template.Template, error) {
	var (
		b   []byte
		err error
	)
	if c.tmpl != "" {
		b, err = os.ReadFile(c.tmpl)
	} else {
		b, err = io.ReadAll(stdin)
	}
	if err != nil {
		return nil, err
	}
	
	return template.
		New("template").
		Funcs(funcMap).
		Parse(string(b))
}

func (c *cli) output(file string, w io.Writer) (io.Writer, func() error, error) {
	if file == "" {
		closeFn := func() error { return nil }
		return w, closeFn, nil
	}
	
	f, err := os.CreateTemp("", "")
	if err != nil {
		return nil, nil, err
	}
	
	var closed bool
	closeFn := func() error {
		if closed {
			return nil
		}
		if err := f.Close(); err != nil {
			return err
		}
		err := os.Rename(f.Name(), file)
		closed = true
		return err
	}
	
	return f, closeFn, nil
}

var (
	statRegex      = regexp.MustCompile(`(?P<file>[\w.\-_/]+)\s*|\s*(?P<count>\d+)\s*(?P<additions>\+*)(?P<removals>-*)`)
	statGroupNames = statRegex.SubexpNames()
	
	spaceRegex   = regexp.MustCompile(`\s+`)
	nonwordRegex = regexp.MustCompile(`[^0-9a-z-]+`)
)

var funcMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"markdownHeaderLink": func(s string) string {
		s = strings.ToLower(s)
		s = spaceRegex.ReplaceAllString(s, "-")
		s = nonwordRegex.ReplaceAllString(s, "")
		return s
	},
	"statsHTML": statsHTML,
	"title": func(s string) string {
		if s == "" {
			return ""
		}
		sep := " "
		ss := strings.SplitN(s, " ", 2)
		r := []rune(ss[0])
		r[0] = unicode.ToUpper(r[0])
		if len(ss) == 1 {
			return string(r)
		}
		return string(r) + sep + ss[1]
	},
}

func statsHTML(in string) string {
	var (
		sb         strings.Builder
		line       string
		firstWrite bool
	)
	writeLine := func(line string) {
		if !firstWrite {
			sb.WriteString("| File | Count | Diff |\n")
			sb.WriteString("| ------ | ------ | ------ |\n")
			firstWrite = true
		}
		if !strings.HasSuffix(line, "|") {
			line += " |"
		}
		sb.WriteString(line + "\n")
	}
	for _, matches := range statRegex.FindAllStringSubmatch(in, -1) {
		for groupIdx, match := range matches {
			if groupIdx == 0 || match == "" {
				continue
			}
			switch statGroupNames[groupIdx] {
			case "file":
				if line != "" {
					writeLine(line)
				}
				line = fmt.Sprintf(`| [%[1]s](%[1]s) `, match)
			case "count":
				line += fmt.Sprintf(`| **%s** | `, match)
			case "additions":
				line += fmt.Sprintf(`<span style="color:green">%s</span>`, match)
			case "removals":
				line += fmt.Sprintf(`<span style="color:red">%s</span>`, match)
			}
		}
	}
	if line != "" {
		writeLine(line)
	}
	return sb.String()
}

type input struct {
	Commits commitSlc
}

type commitSlc []commit

func (c commitSlc) DropByField(field, value string) commitSlc {
	matchFn := fieldsMatcherGen(field, false)
	return c.filter(func(c commit) bool {
		return matchFn(c, value)
	})
}

func (c commitSlc) KeepByField(field, value string) commitSlc {
	matchFn := fieldsMatcherGen(field, true)
	return c.filter(func(c commit) bool {
		return matchFn(c, value)
	})
}

func (c commitSlc) DropByNote(noteType, value string) commitSlc {
	return c.filter(func(c commit) bool {
		for _, n := range c.CC.Notes {
			if n.Type == noteType && n.Value == value {
				return false
			}
		}
		return true
	})
}

func (c commitSlc) KeepByNote(noteType, value string) commitSlc {
	return c.filter(func(c commit) bool {
		for _, n := range c.CC.Notes {
			if n.Type == noteType && n.Value == value {
				return true
			}
		}
		return false
	})
}

func (c commitSlc) filter(filterFn func(commit) bool) commitSlc {
	var out commitSlc
	for _, com := range c {
		if filterFn(com) {
			out = append(out, com)
		}
	}
	return out
}

func fieldsMatcherGen(field string, keep bool) func(c commit, v string) bool {
	cmpFn := func(a, b string) bool { return a == b }
	if !keep {
		cmpFn = func(a, b string) bool { return a != b }
	}
	return func(c commit, v string) bool {
		
		switch field {
		case "Author":
			return cmpFn(c.Author, v)
		case "Scope":
			return cmpFn(c.CC.Scope, v)
		case "Type":
			return cmpFn(c.CC.Type, v)
		default:
			return !keep
		}
	}
}

type (
	commit struct {
		Author    string
		Hash      string
		HashShort string
		Message   string
		Stats     string
		CC        conventional
	}
	
	conventional struct {
		Body   string
		Desc   string
		Footer string
		Header string
		Notes  noteSlc
		Scope  string
		Type   string
	}
	
	note struct {
		Type  string
		Value string
	}
)

type noteSlc []note

func (n noteSlc) KeepByType(nType string) noteSlc {
	var out noteSlc
	for _, note := range n {
		if note.Type == nType {
			out = append(out, note)
		}
	}
	return out
}

func parseGitTemplVars(r *git.Repository) ([]commit, error) {
	iter, err := r.Log(&git.LogOptions{})
	if err != nil {
		return nil, err
	}
	
	p := parser.New()
	
	var commits []commit
	err = iter.ForEach(func(c *object.Commit) error {
		com := commit{
			Message: c.Message,
			Hash:    c.Hash.String(),
		}
		if maxLen := 7; len(com.Hash) > maxLen {
			com.HashShort = com.Hash[:maxLen]
		}
		
		fi, err := c.Stats()
		if err != nil {
			return err
		}
		com.Stats = fi.String()
		
		if cc, err := p.Parse(c.Message); err == nil {
			var notes []note
			for _, n := range cc.Notes() {
				notes = append(notes, note{
					Type:  n.Token(),
					Value: n.Value(),
				})
			}
			
			com.CC = conventional{
				Body:   cc.Body(),
				Desc:   cc.Description(),
				Footer: cc.Footer(),
				Header: cc.Header(),
				Notes:  notes,
				Scope:  cc.Scope(),
				Type:   cc.Type(),
			}
		}
		
		commits = append(commits, com)
		return nil
	})
	
	slices.Reverse(commits)
	return commits, err
}
