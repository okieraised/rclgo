package gogen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	blacklistRe *regexp.Regexp
)

func init() { prepareBlacklistMatchingRegexp() }

// Build one union regex with a named group per rule: (?P<r0>pat0)|(?P<r1>pat1)|...
func prepareBlacklistMatchingRegexp() {
	if len(blacklistedMessages) == 0 {
		blacklistRe = nil
		return
	}
	var b strings.Builder
	for i, pat := range blacklistedMessages {
		if i > 0 {
			b.WriteByte('|')
		}
		// Keep each rule as-is (it’s already a regex). Named groups let us
		// reliably identify which rule matched even if a rule has its own groups.
		_, _ = fmt.Fprintf(&b, "(?P<r%d>%s)", i, pat)
	}
	re, err := regexp.Compile(b.String())
	if err != nil {
		panic(fmt.Errorf("invalid blacklist pattern: %w", err))
	}
	blacklistRe = re
}

func blacklisted(path string) (bool, string) {
	if blacklistRe == nil {
		return false, ""
	}
	sub := blacklistRe.FindStringSubmatch(path)
	if sub == nil {
		return false, ""
	}
	names := blacklistRe.SubexpNames()

	for i := len(sub) - 1; i >= 1; i-- {
		if sub[i] == "" {
			continue
		}
		// Our named groups are r0, r1, ...
		if strings.HasPrefix(names[i], "r") {
			if idx, err := strconv.Atoi(names[i][1:]); err == nil && idx >= 0 && idx < len(blacklistedMessages) {
				return true, blacklistedMessages[idx]
			}
		}
	}
	return true, "" // matched but couldn't resolve which (shouldn't happen)
}
