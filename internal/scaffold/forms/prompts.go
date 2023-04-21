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

// func promptMultiple(prompt *textinput.TextInput, another *confirmation.Confirmation) ([]string, error) {
// 	var values []string
// 	askAgain := true
// 	for askAgain {
// 		value, err := prompt.RunPrompt()
// 		if err != nil {
// 			return nil, err
// 		}
// 		values = append(values, value)
// 		askAgain, err = another.RunPrompt()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return values, nil
// }

type Prompter[T any] interface {
	RunPrompt() (T, error)
}

type optionalPrompt[T any] struct {
	enable *confirmation.Confirmation
	prompt Prompter[T]
}

func (p *optionalPrompt[T]) RunPrompt() (T, error) {
	var zero T
	enabled, err := p.enable.RunPrompt()
	if err != nil {
		return zero, err
	}
	if !enabled {
		return zero, nil
	}
	value, err := p.prompt.RunPrompt()
	if err != nil {
		return zero, err
	}
	return value, nil
}
