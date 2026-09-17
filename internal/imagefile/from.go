package imagefile

import (
	"errors"
	"strings"
)

func (p *parser) from(arguments string) error {
	words := strings.Fields(arguments)
	named := len(words) == 3 && strings.EqualFold(words[1], "AS")

	switch {
	case len(words) == 0:
		return errors.New("FROM needs a base")
	case len(words) == 1:
		p.stages = append(p.stages, Stage{Base: words[0]})
	case named:
		p.stages = append(p.stages, Stage{Base: words[0], Name: words[2]})
	default:
		return errors.New("FROM takes a base and an optional AS name")
	}

	return nil
}
