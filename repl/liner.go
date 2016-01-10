package zygo

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
)

var history_fn = filepath.Join("~/.zygohist")

var completion_keywords = []string{"(and ", "(or ", "(cond ", "(quote ", "(mdef ", "(fn ", "(defn ", "(begin ", "(let ", "(let* ", "(defmac ", "(assert ", "(macexpand ", "(syntax-quote ", "(include ", "(source ", "(req ", "(for ", "(set ", "(break ", "(continue ", "(now ", "(timeit "}

type Prompter struct {
	prompt   string
	prompter *liner.State
	origMode liner.ModeApplier
	rawMode  liner.ModeApplier
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
	//p.prompter.SetTabCompletionStyle(liner.TabPrints)

	p.prompter.SetCompleter(func(line string) (c []string) {
		for _, n := range completion_keywords {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	if f, err := os.Open(history_fn); err == nil {
		p.prompter.ReadHistory(f)
		f.Close()
	}

	return p
}

func (p *Prompter) Close() {
	defer p.prompter.Close()
	if f, err := os.Create(history_fn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		p.prompter.WriteHistory(f)
		f.Close()
	}
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
