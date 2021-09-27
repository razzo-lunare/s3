package str

import (
	"regexp"
)

// RegexGroups returns the group matches for the incoming regex as a map
func RegexGroups(newRegex *regexp.Regexp, newStr string) map[string]string {
	subexValues := newRegex.FindStringSubmatch(newStr)
	subexpNames := newRegex.SubexpNames()

	regexGroups := map[string]string{}
	for index := range subexpNames {
		if index == 0 {
			continue
		}
		if index >= len(subexValues) {
			break
		}

		if subexpNames[index] != "" && subexValues[index] != "" {
			regexGroups[subexpNames[index]] = subexValues[index]
		}
	}

	return regexGroups
}
