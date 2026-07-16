package ui

import "strings"

type navigationHint struct {
	key         string
	description string
}

func renderNavigationHints(hints ...navigationHint) string {
	if len(hints) == 0 {
		return ""
	}

	parts := make([]string, 0, len(hints))
	for _, hint := range hints {
		parts = append(parts, NavigationHintKeyStyle.Render(hint.key)+NavigationHintStyle.Render(" - "+hint.description+";"))
	}

	return strings.Join(parts, " ")
}
