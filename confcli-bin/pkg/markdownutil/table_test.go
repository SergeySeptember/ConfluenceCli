package markdownutil

import (
	"strings"
	"testing"
)

func TestSeparateTableLists(t *testing.T) {
	table := "| Field | Value |\n| :--- | ---: |\n| id | value |"
	for _, tc := range []struct {
		name, source, want string
	}{
		{"resumed ordered list", table + "\n12. Continue\n    - Detail", table + "\n\n12. Continue\n    - Detail"},
		{"correctly nested table", "11. Details:\n\n    " + strings.ReplaceAll(table, "\n", "\n    ") + "\n12. Continue", "11. Details:\n\n    " + strings.ReplaceAll(table, "\n", "\n    ") + "\n12. Continue"},
		{"unordered list", table + "\n- Continue", table + "\n\n- Continue"},
		{"already separated", table + "\n\n12. Continue", table + "\n\n12. Continue"},
		{"pipe in list text", table + "\n12. Use a | b", table + "\n\n12. Use a | b"},
		{"cell text retained", table + "\n| 12. Cell text | - Cell text |", table + "\n| 12. Cell text | - Cell text |"},
		{"not a table", "| Text |\n| not a delimiter |\n12. Continue", "| Text |\n| not a delimiter |\n12. Continue"},
		{"fenced example", "````text\n" + table + "\n12. Continue\n```\n````", "````text\n" + table + "\n12. Continue\n```\n````"},
		{"raw HTML example", "<pre>\n" + table + "\n12. Continue\n</pre>", "<pre>\n" + table + "\n12. Continue\n</pre>"},
		{"indented example", "    " + strings.ReplaceAll(table, "\n", "\n    ") + "\n    12. Continue", "    " + strings.ReplaceAll(table, "\n", "\n    ") + "\n    12. Continue"},
		{"CRLF", strings.ReplaceAll(table+"\n12. Continue\n", "\n", "\r\n"), strings.ReplaceAll(table+"\n\n12. Continue\n", "\n", "\r\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := SeparateTableLists(tc.source)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			if SeparateTableLists(got) != got {
				t.Fatal("repair must be idempotent")
			}
		})
	}
}
