// © 2023 Snyk Limited All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
