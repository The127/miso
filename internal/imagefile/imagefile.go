package imagefile

import "strings"

// Stage is a FROM line and the instructions that follow it.
type Stage struct {
	Name     string
	Base     string
	Commands []string
}

// Parse reads a build file into its stages, in order.
func Parse(source string) []Stage {
	var stages []Stage
	for _, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, "FROM ") {
			base, name := from(line)
			stages = append(stages, Stage{Name: name, Base: base})
		}
		if command, ok := strings.CutPrefix(line, "RUN "); ok {
			current := &stages[len(stages)-1]
			current.Commands = append(current.Commands, command)
		}
	}
	return stages
}

func from(line string) (base, name string) {
	fields := strings.Fields(line)
	base = fields[1]
	if len(fields) > 3 {
		name = fields[3]
	}
	return base, name
}
