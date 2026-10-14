package converters

import (
	"strings"
	"testing"

	"confcli/pkg/converters/md"
	"golang.org/x/net/html"
)

func TestExportKeepsTableBoundaryBeforeResumedList(t *testing.T) {
	source := `<ol start="10"><li>Check request</li><li>Add details:</li></ol><table><tr><th>Field</th><th>Value</th></tr><tr><td>bankId</td><td>Identifier</td></tr></table><ol start="12"><li>Find receiver<ul><li>Use identifier</li></ul></li><li>Send response</li></ol>`
	markdown, err := StorageToMarkdownAdvanced(source, "")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := md.ToStorage([]byte(markdown))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(markdown, "|\n\n12.") {
		t.Fatalf("table/list boundary removed:\n%s", markdown)
	}
	assertResumedListOutsideTable(t, storage)
}

func assertResumedListOutsideTable(t *testing.T, storage string) {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(storage))
	if err != nil {
		t.Fatal(err)
	}
	rows, resumed := 0, false
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, inTable bool) {
		inTable = inTable || n.Data == "table"
		if n.Data == "tr" {
			rows++
		}
		if n.Data == "ol" {
			for _, a := range n.Attr {
				if a.Key == "start" && a.Val == "12" && !inTable {
					resumed = true
				}
			}
		}
		if n.Type == html.TextNode && (strings.Contains(n.Data, "Find receiver") || strings.Contains(n.Data, "Use identifier") || strings.Contains(n.Data, "Send response")) && inTable {
			t.Fatalf("list text became a table cell: %s", storage)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, inTable)
		}
	}
	walk(doc, false)
	if rows != 2 || !resumed {
		t.Fatalf("expected two table rows and resumed numbering at 12: %s", storage)
	}
}

func TestImportSeparatesUnindentedTableFromResumedList(t *testing.T) {
	markdown := "10. Check request\n11. Add details:\n\n| Field | Value |\n| --- | --- |\n| bankId | Identifier |\n12. Find receiver\n    - Use identifier\n13. Send response\n"
	storage, err := md.ToStorage([]byte(markdown))
	if err != nil {
		t.Fatal(err)
	}
	assertResumedListOutsideTable(t, storage)
}
