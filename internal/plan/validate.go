package plan

import "github.com/The127/miso/internal/imagefile"

// Validate checks that the stages make sense together.
func Validate(stages []imagefile.Stage) error {
	known := map[string]map[string]bool{}
	written := map[string]bool{}
	for _, stage := range stages {
		if err := checkName(stage, known); err != nil {
			return err
		}

		if err := checkReferences(stage, known); err != nil {
			return err
		}

		if err := checkChecks(stage); err != nil {
			return err
		}

		if err := checkOutputs(stage, written); err != nil {
			return err
		}

		if stage.Name != "" {
			known[stage.Name] = outputNames(stage)
		}
	}

	return nil
}

func at(line int, err error) error {
	return &imagefile.Error{Line: line, Err: err}
}
