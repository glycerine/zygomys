package zygo

import (
	"sort"
	"strings"

	"github.com/glycerine/liner"
)

// filled at init time based on BuiltinFunctions
var completion_keywords = []string{`(`}

var math_funcs = []string{`* `, `** `, `+ `, `- `, `-> `, `/ `, `< `, `<= `, `== `, `> `, `>= `, `\ `}

func init() {
	// fill in our auto-complete keywords
	sortme := []*SymtabE{}
	for f, _ := range BuiltinFunctions {
		sortme = append(sortme, &SymtabE{Key: f})
	}
	sort.Sort(SymtabSorter(sortme))
	for i := range sortme {
		completion_keywords = append(completion_keywords, "("+sortme[i].Key)
	}

	for i := range math_funcs {
		completion_keywords = append(completion_keywords, "("+math_funcs[i])
	}
}

type Prompter struct {
	prompt   string
	prompter *liner.State
	origMode liner.ModeApplier
	rawMode  liner.ModeApplier
}

// complete phrases that start with '('
func MyWordCompleter(line string, pos int) (head string, c []string, tail string) {

	beg := []rune(line[:pos])
	end := line[pos:]
	Q("\nline = '%s' pos=%v\n", line, pos)
	Q("\nbeg = '%v'\nend = '%s'\n", string(beg), end)
	// find most recent paren in beg
	n := len(beg)
	last := n - 1
	var i int
	var p int = -1
outer:
	for i = last; i >= 0; i-- {
		Q("\nbeg[i=%v] is '%v'\n", i, string(beg[i]))
		switch beg[i] {
		case ' ':
			break outer
		case '(':
			p = i
			Q("\n found paren at p = %v\n", i)
			break outer
		}
	}
	Q("p=%d\n", p)
	prefix := string(beg)
	extendme := ""
	if p == 0 {
		prefix = ""
		extendme = string(beg)
	} else if p > 0 {
		prefix = string(beg[:p])
		extendme = string(beg[p:])
	}
	Q("prefix = '%s'\nextendme = '%s'\n", prefix, extendme)

	for _, n := range completion_keywords {
		if strings.HasPrefix(n, strings.ToLower(extendme)) {
			Q("n='%s' has prefix  = '%s'\n", n, extendme)
			c = append(c, n)
		}
	}

	return prefix, c, end
}

func NewPrompter() *Prompter {
	origMode, err := liner.TerminalMode()
	if err != nil {
		panic(err)
	}

	p := &Prompter{
		prompt:   "zygo> ",
		prompter: liner.NewLiner(),
		origMode: origMode,
	}

	rawMode, err := liner.TerminalMode()
	if err != nil {
		panic(err)
	}
	p.rawMode = rawMode

	p.prompter.SetCtrlCAborts(false)
	p.prompter.SetWordCompleter(liner.WordCompleter(MyWordCompleter))

	return p
}

func (p *Prompter) Close() {
	defer p.prompter.Close()
}

func (p *Prompter) Getline(prompt *string) (line string, err error) {
	applyErr := p.rawMode.ApplyMode()
	if applyErr != nil {
		panic(applyErr)
	}
	defer func() {
		applyErr := p.origMode.ApplyMode()
		if applyErr != nil {
			panic(applyErr)
		}
	}()

	if prompt == nil {
		line, err = p.prompter.Prompt(p.prompt)
	} else {
		line, err = p.prompter.Prompt(*prompt)
	}
	if err == nil {
		p.prompter.AppendHistory(line)
		return line, nil
	}
	return "", err
}
