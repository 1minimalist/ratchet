package stages

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/DailyBurn/ratchet/data"
	"github.com/DailyBurn/ratchet/util"
)

// CSVWriter is handles converting data.JSON objects into CSV format,
// and writing them to the given io.Writer. The Data
// must be a valid JSON object or a slice of valid JSON objects.
// If you already have Data formatted as a CSV string you can
// use an IoWriter instead.
type CSVWriter struct {
	writer        *csv.Writer
	WriteHeader   bool
	headerWritten bool
}

// NewCSVWriter returns a new CSVWriter wrapping the given io.Writer object
func NewCSVWriter(w io.Writer) *CSVWriter {
	return &CSVWriter{writer: csv.NewWriter(w), WriteHeader: true, headerWritten: false}
}

// ProcessData - see interface in stages.go for documentation.
func (w *CSVWriter) ProcessData(d data.JSON, outputChan chan data.JSON, killChan chan error) {
	// use util helper to convert Data into []map[string]interface{}
	objects, err := data.ObjectsFromJSON(d)
	util.KillPipelineIfErr(err, killChan)

	rows := [][]string{}
	if w.WriteHeader && !w.headerWritten {
		header := []string{}
		for k := range objects[0] {
			header = append(header, k)
		}
		rows = append(rows, header)
		w.headerWritten = true
	}

	for _, object := range objects {
		row := []string{}
		for _, v := range object {
			row = append(row, fmt.Sprintf("%v", v))
		}
		rows = append(rows, row)
	}

	err = w.writer.WriteAll(rows)
	util.KillPipelineIfErr(err, killChan)
}

// Finish - see interface for documentation.
func (w *CSVWriter) Finish(outputChan chan data.JSON, killChan chan error) {
	if outputChan != nil {
		close(outputChan)
	}
}

func (w *CSVWriter) String() string {
	return "CSVWriter"
}
