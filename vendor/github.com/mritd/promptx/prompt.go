package promptx

import (
	"text/template"

	"github.com/mritd/promptx/util"
	"github.com/mritd/readline"
)

const (
	DefaultPrompt         = "»"
	DefaultErrorMsgPrefix = "✘ "
	DefaultAskTpl         = "{{ . | cyan }} "
	DefaultPromptTpl      = "{{ . | green }} "
	DefaultInvalidTpl     = "{{ . | red }} "
	DefaultValidTpl       = "{{ . | green }} "
	DefaultErrorMsgTpl    = "{{ . | red }} "
)

type Prompt struct {
	Config
	Ask     string
	Prompt  string
	FuncMap template.FuncMap

	isFirstRun bool

	ask      *template.Template
	prompt   *template.Template
	valid    *template.Template
	invalid  *template.Template
	errorMsg *template.Template
}

type Config struct {
	AskTpl        string
	PromptTpl     string
	ValidTpl      string
	InvalidTpl    string
	ErrorMsgTpl   string
	CheckListener func(line []rune) error
}

func NewDefaultConfig(check func(line []rune) error) Config {
	return Config{
		AskTpl:        DefaultAskTpl,
		PromptTpl:     DefaultPromptTpl,
		InvalidTpl:    DefaultInvalidTpl,
		ValidTpl:      DefaultValidTpl,
		ErrorMsgTpl:   DefaultErrorMsgTpl,
		CheckListener: check,
	}
}

func NewDefaultPrompt(check func(line []rune) error, ask string) Prompt {
	return Prompt{
		Ask:     ask,
		Prompt:  DefaultPrompt,
		FuncMap: FuncMap,
		Config:  NewDefaultConfig(check),
	}
}

func (p *Prompt) prepareTemplates() {

	var err error
	p.ask, err = template.New("").Funcs(FuncMap).Parse(p.AskTpl)
	util.CheckAndExit(err)
	p.prompt, err = template.New("").Funcs(FuncMap).Parse(p.PromptTpl)
	util.CheckAndExit(err)
	p.valid, err = template.New("").Funcs(FuncMap).Parse(p.ValidTpl)
	util.CheckAndExit(err)
	p.invalid, err = template.New("").Funcs(FuncMap).Parse(p.InvalidTpl)
	util.CheckAndExit(err)
	p.errorMsg, err = template.New("").Funcs(FuncMap).Parse(p.ErrorMsgTpl)
	util.CheckAndExit(err)

}

func (p *Prompt) Run() string {
	p.isFirstRun = true
	p.prepareTemplates()

	displayPrompt := append(util.Render(p.prompt, p.Prompt), util.Render(p.ask, p.Ask)...)
	validPrompt := append(util.Render(p.valid, p.Prompt), util.Render(p.ask, p.Ask)...)
	invalidPrompt := append(util.Render(p.invalid, p.Prompt), util.Render(p.ask, p.Ask)...)

	l, err := readline.NewEx(&readline.Config{
		Prompt:                 string(displayPrompt),
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
	})
	util.CheckAndExit(err)

	filterInput := func(r rune) (rune, bool) {

		switch r {
		// block CtrlZ feature
		case readline.CharCtrlZ:
			return r, false
		default:
			return r, true
		}
	}

	l.Config.FuncFilterInputRune = filterInput

	l.Config.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		// Real-time verification
		if err = p.CheckListener(line); err != nil {
			l.SetPrompt(string(invalidPrompt))
			l.Refresh()
		} else {
			l.SetPrompt(string(validPrompt))
			l.Refresh()
		}
		return nil, 0, false
	})
	defer l.Close()

	// read line
	for {
		if !p.isFirstRun {
			l.Write([]byte(moveUp))
		}
		s, err := l.Readline()
		util.CheckAndExit(err)
		if err = p.CheckListener([]rune(s)); err != nil {
			l.Write([]byte(clearLine))
			l.Write([]byte(string(util.Render(p.errorMsg, DefaultErrorMsgPrefix+err.Error()))))
			p.isFirstRun = false
		} else {
			return s
		}
	}
}
