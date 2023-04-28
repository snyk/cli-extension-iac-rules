package forms

import (
	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/textinput"
)

type multiplePrompt struct {
	prompt  *textinput.TextInput
	another *confirmation.Confirmation
}

func (p *multiplePrompt) RunPrompt() ([]string, error) {
	var values []string
	askAgain := true
	for askAgain {
		value, err := p.prompt.RunPrompt()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		askAgain, err = p.another.RunPrompt()
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}
