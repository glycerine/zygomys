package zygo

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
)

var history_fn = filepath.Join("~/.zygohist")

var completion_keywords = []string{"and", "or", "cond", "quote", "mdef", "fn", "defn", "begin", "let", "let*", "defmac", "assert", "macexpand", "syntax-quote", "include", "source", "req", "for", "set", "break", "continue", "now", "time"}

type Prompter struct {
	prompt   string
	prompter *liner.State
}

func NewPrompter() *Prompter {
	p := &Prompter{
		prompt:   "zygo> ",
		prompter: liner.NewLiner(),
	}

	p.prompter.SetCtrlCAborts(true)
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
	if prompt == nil {
		line, err = p.prompter.Prompt(p.prompt)
	} else {
		line, err = p.prompter.Prompt(*prompt)
	}
	if err == nil {
		p.prompter.AppendHistory(line)
		return line, nil
	} else if err == liner.ErrPromptAborted {
		log.Print("Aborted")
	} else {
		log.Print("Error reading line: ", err)
	}
	return "", err
}
