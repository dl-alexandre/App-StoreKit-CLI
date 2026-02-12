package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/itchyny/gojq"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
	FormatRaw   Format = "raw"
)

func Render(w io.Writer, data any, format Format, jqExpr string, columns []string) error {
	if jqExpr != "" {
		filtered, err := applyJQ(data, jqExpr)
		if err != nil {
			return err
		}
		data = filtered
	}

	switch format {
	case FormatTable:
		return renderTable(w, data, columns)
	case FormatRaw:
		return renderRaw(w, data)
	case FormatJSON, "":
		return renderJSON(w, data)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func renderJSON(w io.Writer, data any) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	_, err := io.Copy(w, buf)
	return err
}

func renderRaw(w io.Writer, data any) error {
	switch v := data.(type) {
	case []byte:
		_, err := w.Write(v)
		return err
	case string:
		_, err := io.WriteString(w, v)
		return err
	default:
		return renderJSON(w, data)
	}
}

func renderTable(w io.Writer, data any, columns []string) error {
	rows, ok := data.([]map[string]any)
	if !ok {
		return errors.New("table format requires array of objects")
	}
	if len(rows) == 0 {
		return nil
	}

	keys := make([]string, 0)
	if len(columns) > 0 {
		keys = append(keys, columns...)
	} else {
		keys = make([]string, 0, len(rows[0]))
		for key := range rows[0] {
			keys = append(keys, key)
		}
		sort.Strings(keys)
	}

	t := table.NewWriter()
	t.SetOutputMirror(w)

	header := make(table.Row, 0, len(keys))
	for _, key := range keys {
		header = append(header, key)
	}
	t.AppendHeader(header)

	for _, row := range rows {
		record := make(table.Row, 0, len(keys))
		for _, key := range keys {
			record = append(record, row[key])
		}
		t.AppendRow(record)
	}

	t.Render()
	return nil
}

func applyJQ(data any, expr string) (any, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, err
	}

	iter := query.Run(data)
	var results []any
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		results = append(results, v)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
