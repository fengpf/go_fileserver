package go_fileserver

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

var csvContentType = []string{"text/csv; charset=utf-8"}

// CSV Comma-Separated Values struct.
type CSV struct {
	Data  [][]string
	Title string
}

// Render (CSV) writes data with CSV ContentType.
func (c CSV) Render(w io.Writer) (err error) {
	writer := csv.NewWriter(w)

	bomUtf8 := []byte{0xEF, 0xBB, 0xBF}
	writer.Write([]string{string(bomUtf8[:])})

	writer.WriteAll(c.Data)
	if err = writer.Error(); err != nil {
		err = errors.WithStack(err)
	}

	writer.Flush()
	return
}

// WriteContentType write CSV ContentType.
func (c CSV) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = csvContentType
	}
	if c.Title != "" {
		header["Content-Disposition"] = append(
			header["Content-Disposition"],
			fmt.Sprintf("attachment; filename=\"%s.csv\"", c.Title),
		)
	}
}
