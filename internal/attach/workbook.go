package attach

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// sheetXML is the part of a worksheet that holds its cells.
type sheetXML struct {
	Rows []struct {
		Cells []struct {
			Type   string `xml:"t,attr"`
			Value  string `xml:"v"`
			Inline string `xml:"is>t"`
		} `xml:"c"`
	} `xml:"sheetData>row"`
}

// sharedXML is the workbook's table of shared strings.
type sharedXML struct {
	Items []struct {
		Text string   `xml:"t"`
		Runs []string `xml:"r>t"`
	} `xml:"si"`
}

// workbookText returns each sheet as tab-separated rows.
func workbookText(z *zip.Reader) (string, error) {
	shared, err := sharedStrings(z)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, name := range numbered(z, "xl/worksheets/sheet", ".xml") {
		var sheet sheetXML
		if err := decodeZipped(z, name, &sheet); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "## Sheet %d\n", i+1)
		for _, row := range sheet.Rows {
			cells := make([]string, len(row.Cells))
			for j, c := range row.Cells {
				cells[j] = cellText(c.Type, c.Value, c.Inline, shared)
			}
			b.WriteString(strings.Join(cells, "\t") + "\n")
		}
		b.WriteString("\n")
	}
	return b.String(), nil
}

func cellText(kind, value, inline string, shared []string) string {
	switch kind {
	case "s":
		if i, err := strconv.Atoi(value); err == nil && i >= 0 && i < len(shared) {
			return shared[i]
		}
	case "inlineStr":
		return inline
	}
	return value
}

// sharedStrings reads the shared string table, which a workbook may not
// have.
func sharedStrings(z *zip.Reader) ([]string, error) {
	var table sharedXML
	if _, err := z.Open("xl/sharedStrings.xml"); err != nil {
		return nil, nil
	}
	if err := decodeZipped(z, "xl/sharedStrings.xml", &table); err != nil {
		return nil, err
	}
	out := make([]string, len(table.Items))
	for i, item := range table.Items {
		out[i] = item.Text + strings.Join(item.Runs, "")
	}
	return out, nil
}

func decodeZipped(z *zip.Reader, name string, into any) error {
	r, err := openZipped(z, name)
	if err != nil {
		return err
	}
	defer r.Close()
	if err := xml.NewDecoder(r).Decode(into); err != nil {
		return fmt.Errorf("unreadable %s: %w", name, err)
	}
	return nil
}
