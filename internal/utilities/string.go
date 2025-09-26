package utilities

import (
	"maps"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

func LicenseHeader(license string) string {
	if license == "" {
		return ""
	}
	return "/*\n" + license + "*/\n\n"
}

func UpperCaseFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func SnakeToCamel(in string) string {
	tmp := []rune(in)
	tmp[0] = unicode.ToUpper(tmp[0])
	for i := 0; i < len(tmp); i++ {
		if tmp[i] == '_' {
			tmp[i+1] = unicode.ToUpper(tmp[i+1])
			tmp = append(tmp[:i], tmp[i+1:]...)
			i--
		}
	}
	return string(tmp)
}

func CamelToSnake(in string) string {
	tmp := []rune(in)
	sb := strings.Builder{}
	sb.Grow(len(tmp))

	ucSequenceLength := 0 //Special semantics for consecutive UC characters

	for i := 0; i < len(tmp); i++ {
		if unicode.IsUpper(tmp[i]) || (ucSequenceLength > 0 && unicode.IsNumber(tmp[i])) {
			ucSequenceLength++

			if i == 0 {
				sb.WriteRune(unicode.ToLower(tmp[i]))
			} else if ucSequenceLength == 1 {
				sb.WriteRune('_')
				sb.WriteRune(unicode.ToLower(tmp[i]))
			} else if i+1 >= len(tmp) {
				sb.WriteRune(unicode.ToLower(tmp[i]))
			} else if unicode.IsUpper(tmp[i+1]) || unicode.IsNumber(tmp[i+1]) {
				sb.WriteRune(unicode.ToLower(tmp[i]))
			} else {
				sb.WriteRune('_')
				sb.WriteRune(unicode.ToLower(tmp[i]))
			}
		} else {
			ucSequenceLength = 0
			sb.WriteRune(tmp[i])
		}
	}
	return sb.String()
}

func CommentSerializer(lineComment string, preComments *strings.Builder) string {
	if preComments.Len() == 0 {
		return lineComment
	}
	defer preComments.Reset()
	if lineComment == "" {
		return preComments.String()
	}
	return lineComment + `. ` + preComments.String()
}

type StringSet map[string]struct{}

func (s StringSet) Add(strs ...string) {
	for _, str := range strs {
		s[str] = struct{}{}
	}
}

func (s StringSet) AddFrom(s2 StringSet) {
	for k := range s2 {
		s[k] = struct{}{}
	}
}

func (s StringSet) ToSlice() []string {
	return slices.Collect(maps.Keys(map[string]struct{}(s)))
}

func (s StringSet) ToSortedSlice() []string {
	vals := slices.Collect(maps.Keys(map[string]struct{}(s)))
	slices.Sort(vals)
	return vals
}
