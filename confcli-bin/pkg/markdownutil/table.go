package markdownutil

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var tableFollowingList = regexp.MustCompile(`^[ \t]*(?:[-*+]|[0-9]+[.)])[ \t]+`)

// SeparateTableLists keeps a list immediately following a pipe table from being
// parsed as more table rows. Only an explicit table header/delimiter and bordered
// rows trigger the repair; cell contents and code examples remain unchanged.
func SeparateTableLists(source string) string {
	lines := strings.Split(source, "\n")
	masked := strings.Split(maskTableSource(source), "\n")
	result := make([]string, 0, len(lines))
	inTable := false
	tableIndent := 0
	for i, line := range lines {
		current := strings.TrimSpace(masked[i])
		if inTable && !strings.HasPrefix(current, "|") {
			// A dedented list item already closes a properly nested table.
			if tableFollowingList.MatchString(masked[i]) && tableLineIndent(masked[i]) >= tableIndent {
				blank := ""
				if i > 0 && strings.HasSuffix(lines[i-1], "\r") {
					blank = "\r"
				}
				result = append(result, blank)
			}
			inTable = false
		}
		if i > 0 && strings.HasPrefix(strings.TrimSpace(masked[i-1]), "|") && isTableDelimiter(current) {
			inTable = true
			tableIndent = tableLineIndent(masked[i])
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// Raw HTML, including <pre> examples, is already responsible for its own block
// boundaries. Do not insert Markdown separators inside it.
func maskTableSource(source string) string {
	masked := []byte(MaskCode(source))
	doc := goldmark.DefaultParser().Parse(text.NewReader(masked))
	mask := func(start, end int) {
		for i := start; i < end; i++ {
			if masked[i] != '\n' && masked[i] != '\r' {
				masked[i] = ' '
			}
		}
	}
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		block, ok := n.(*ast.HTMLBlock)
		if entering && ok {
			for i := 0; i < block.Lines().Len(); i++ {
				segment := block.Lines().At(i)
				mask(segment.Start, segment.Stop)
			}
			if block.HasClosure() {
				mask(block.ClosureLine.Start, block.ClosureLine.Stop)
			}
		}
		return ast.WalkContinue, nil
	})
	return string(masked)
}

func tableLineIndent(line string) int {
	indent := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			indent++
		case '\t':
			indent += 4 - indent%4
		default:
			return indent
		}
	}
	return indent
}

func isTableDelimiter(line string) bool {
	if !strings.HasPrefix(line, "|") {
		return false
	}
	cells := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|"), "|")
	for _, cell := range cells {
		cell = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(cell), ":"), ":")
		if cell == "" || strings.Trim(cell, "-") != "" {
			return false
		}
	}
	return true
}
