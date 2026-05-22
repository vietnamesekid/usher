package ui

import (
	"github.com/AlecAivazis/survey/v2"
)

type Prompt struct{}

func NewPrompt() *Prompt { return &Prompt{} }

func (p *Prompt) AskString(question, defaultVal string) string {
	var answer string
	q := &survey.Input{
		Message: question,
		Default: defaultVal,
	}
	_ = survey.AskOne(q, &answer)
	return answer
}

func (p *Prompt) AskSecret(question string) string {
	var answer string
	q := &survey.Password{
		Message: question,
	}
	_ = survey.AskOne(q, &answer)
	return answer
}

func (p *Prompt) AskSelect(question string, options []string) string {
	var answer string
	q := &survey.Select{
		Message: question,
		Options: options,
	}
	_ = survey.AskOne(q, &answer)
	return answer
}

func (p *Prompt) AskConfirm(question string) bool {
	var answer bool
	q := &survey.Confirm{
		Message: question,
	}
	_ = survey.AskOne(q, &answer)
	return answer
}

func (p *Prompt) AskMultiSelect(question string, options []string) []string {
	return p.AskMultiSelectWithDefaults(question, options, nil)
}

func (p *Prompt) AskMultiSelectWithDefaults(question string, options, defaults []string) []string {
	var answers []string
	q := &survey.MultiSelect{
		Message: question,
		Options: options,
		Default: defaults,
	}
	_ = survey.AskOne(q, &answers)
	return answers
}
