package utilities

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperCaseFirst(t *testing.T) {
	in := "testString"
	expected := "TestString"
	assert.Equal(t, expected, UpperCaseFirst(in))
}

func TestSnakeToCamel(t *testing.T) {
	in := "test_string"
	expected := "TestString"
	assert.Equal(t, expected, SnakeToCamel(in))
}

func TestCamelToSnake(t *testing.T) {
	in := "TestString"
	expected := "test_string"
	assert.Equal(t, expected, CamelToSnake(in))

	in = "TestSString"
	expected = "test_s_string"
	assert.Equal(t, expected, CamelToSnake(in))
}

func TestCommentSerializer(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		pre          string
		want         string
		wantAfterLen int // expected builder length after call
	}{
		{
			name:         "No preComments -> returns line only",
			line:         "hello",
			pre:          "",
			want:         "hello",
			wantAfterLen: 0, // stays empty (no Reset called)
		},
		{
			name:         "Line empty -> returns preComments and clears builder",
			line:         "",
			pre:          "pre 1\npre 2",
			want:         "pre 1\npre 2",
			wantAfterLen: 0, // cleared by deferred Reset
		},
		{
			name:         "Both present -> concatenates with . and clears builder",
			line:         "line comment",
			pre:          "pre comment block",
			want:         "line comment. pre comment block",
			wantAfterLen: 0, // cleared by deferred Reset
		},
		{
			name:         "Line already ends with period -> no dedup of dot",
			line:         "ends with.",
			pre:          "pre",
			want:         "ends with.. pre",
			wantAfterLen: 0, // cleared by deferred Reset
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			b.WriteString(tc.pre)

			got := CommentSerializer(tc.line, &b)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
			if b.Len() != tc.wantAfterLen {
				t.Fatalf("builder len after call: want %d, got %d", tc.wantAfterLen, b.Len())
			}
		})
	}
}
