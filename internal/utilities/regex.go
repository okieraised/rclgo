package utilities

import (
	"regexp"
	"strings"
)

var (
	normalizeMsgDefaultArrayValueRE = regexp.MustCompile(`(?m)^\[|]$`)
	reStripSingleQuoteEdges         = regexp.MustCompile(`(?m)^'|'$`)
	reEscapeDoubleQuotes            = regexp.MustCompile(`\\?"`)
	reUnescapeSingleQuotes          = regexp.MustCompile(`\\?'`)
	reStripDoubleQuoteEdges         = regexp.MustCompile(`(?m)^"|"$`)
	reSrvMsgSuffix                  = regexp.MustCompile(`_(?:Request|Response)$`)
	reActionMsgSuffix               = regexp.MustCompile(`_(?:Goal|Result|Feedback|SendGoal_Request|SendGoal_Response|GetResult_Request|GetResult_Response|FeedbackMessage)$`)
	reActionSrvSuffix               = regexp.MustCompile(`_(?:SendGoal|GetResult)$`)
	reRclRetPrefix                  = regexp.MustCompile(`^RCL_RET_`)
	reRmwRetPrefix                  = regexp.MustCompile(`^RMW_RET_`)
)

func normalizeMsgDefaultArrayValue(s string) string {
	return normalizeMsgDefaultArrayValueRE.ReplaceAllString(s, "")
}

func defaultValueSanitizer_(ros2type, defaultValue string) string {
	switch ros2type {
	case "string", "wstring", "U16String":
		if defaultValue != "" {
			defaultValue = reStripDoubleQuoteEdges.ReplaceAllString(defaultValue, "")
			defaultValue = reStripSingleQuoteEdges.ReplaceAllString(defaultValue, "")
			defaultValue = reEscapeDoubleQuotes.ReplaceAllString(defaultValue, `\"`)
			defaultValue = reUnescapeSingleQuotes.ReplaceAllString(defaultValue, `'`)
			defaultValue = `"` + defaultValue + `"`
		} else {
			defaultValue = `""`
		}
	}
	return defaultValue
}

func DefaultValueSanitizer(ros2type, defaultValue string) string {
	switch ros2type {
	case "string", "wstring", "U16String":
		if defaultValue != "" {
			defaultValue = reStripDoubleQuoteEdges.ReplaceAllString(defaultValue, "")
		}
	}
	return defaultValueSanitizer_(ros2type, defaultValue)
}

func splitCSVLike(s string) []string {
	var out []string
	var b strings.Builder
	inSingle, inDouble, escape := false, false, false

	for _, r := range s {
		switch {
		case escape:
			b.WriteRune(r)
			escape = false
			continue
		case r == '\\' && (inSingle || inDouble):
			escape = true
			continue
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			b.WriteRune(r) // keep quotes for sanitizer
			continue
		case r == '"' && !inSingle:
			inDouble = !inDouble
			b.WriteRune(r) // keep quotes for sanitizer
			continue
		case r == ',' && !inSingle && !inDouble:
			out = append(out, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		default:
			b.WriteRune(r)
		}
	}
	out = append(out, strings.TrimSpace(b.String()))
	return out
}

func splitMsgDefaultArrayValues(ros2type, defaultsField string) []string {
	s := strings.TrimSpace(normalizeMsgDefaultArrayValue(defaultsField))
	if s == "" {
		return nil
	}
	values := splitCSVLike(s)

	switch ros2type {
	case "string", "wstring", "U16String":
		for i := range values {
			values[i] = defaultValueSanitizer_(ros2type, values[i])
		}
	}
	return values
}

func SrvNameFromSrvMsgName(s string) string {
	return reSrvMsgSuffix.ReplaceAllString(s, "")
}

func ActionNameFromActionMsgName(s string) string {
	return reActionMsgSuffix.ReplaceAllString(s, "")
}

func ActionNameFromActionSrvName(s string) string {
	return reActionSrvSuffix.ReplaceAllString(s, "")
}

func CReturnCodeNameToGo(n string) string {
	n = reRclRetPrefix.ReplaceAllString(n, "")
	n = reRmwRetPrefix.ReplaceAllString(n, "RMW_")
	return SnakeToCamel(strings.ToLower(n))
}
