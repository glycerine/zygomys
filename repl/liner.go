package zygo

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
)

var history_fn = filepath.Join("~/.zygohist")

var completion_keywords = []string{`(`, `(aget `, `(and `, `(append `, `(apply `, `(array `, `(array? `, `(aset! `, `(assert `, `(begin `, `(bit-and `, `(bit-not `, `(bit-or `, `(bit-xor `, `(break `, `(car `, `(cdr `, `(char? `, `(concat `, `(cond `, `(cons `, `(continue `, `(defmac `, `(defn `, `(dump `, `(empty? `, `(first `, `(float? `, `(fn `, `(for `, `(gensym `, `(hash `, `(hash? `, `(hdel! `, `(hget `, `(hpair `, `(hset! `, `(include `, `(int? `, `(json `, `(keys `, `(len `, `(let `, `(let* `, `(list `, `(list? `, `(macexpand `, `(make-array `, `(map `, `(mdef `, `(mod `, `(msgmap `, `(msgpack `, `(not `, `(not= `, `(now `, `(null? `, `(number? `, `(or `, `(print `, `(print `, `(printf `, `(printf `, `(println `, `(println `, `(quote `, `(raw `, `(raw2str `, `(read `, `(req `, `(rest `, `(second `, `(set `, `(sget `, `(slice `, `(sll `, `(source `, `(source `, `(sra `, `(srl `, `(str `, `(str2sym `, `(string? `, `(sym2str `, `(symbol? `, `(symnum `, `(syntax-quote `, `(timeit `, `(togo `, `(type `, `(unjson `, `(unmsgpack `, `(zero? `, `(!= `, `(* `, `(** `, `(+ `, `(- `, `(-> `, `(/ `, `(< `, `(<= `, `(== `, `(> `, `(>= `, `(\ `}

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
