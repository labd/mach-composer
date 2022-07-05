package generator

import (
	"fmt"
	"regexp"
	"strings"
)

var varRegex = regexp.MustCompile(`\${(component(?:\.[^\}]+)+)}`)

func ParseTemplateVariable(val string) (string, error) {
	matches := varRegex.FindAllStringSubmatch(val, 20)
	if len(matches) == 0 {
		return val, nil
	}

	for _, match := range matches {
		parts := strings.SplitN(match[1], ".", 3)
		if len(parts) < 3 {
			return "", fmt.Errorf(
				"invalid variable '%s'; "+
					"When using a ${component...} variable it has to consist of 2 parts; "+
					"component-name.output-name",
				match[1])
		}

		replacement := fmt.Sprintf("${module.%s.%s}", parts[1], parts[2])
		val = strings.ReplaceAll(val, match[0], replacement)
	}

	return val, nil
}
