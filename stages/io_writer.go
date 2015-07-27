// Package stages holds PipelineStage implementations that
// are generic and potentially useful across any ETL project.
package stages

import (
	"io"
	"github.com/DailyBurn/ratchet/data"
	"github.com/DailyBurn/ratchet/logger"
	"github.com/DailyBurn/ratchet/util"
)

// IoWriter is a stage that wraps any io.Writer objects.
// It can be used to write data out to a File, os.Stdout, or
// any other task that can be supported via io.Writer.
type IoWriter struct {
	Writer io.Writer
}

// NewIoWriter returns a new IoWriter wrapping the given io.Writer object
func NewIoWriter(writer io.Writer) *IoWriter {
	return &IoWriter{Writer: writer}
}

// HandleData - see interface for documentation.
func (w *IoWriter) HandleData(d data.JSON, outputChan chan data.JSON, killChan chan error) {
	bytesWritten, err := w.Writer.Write(d)
	util.KillPipelineIfErr(err, killChan)
	logger.Debug("IoWriter:", bytesWritten, "bytes written")
}

// Finish - see interface for documentation.
func (w *IoWriter) Finish(outputChan chan data.JSON, killChan chan error) {
	if outputChan != nil {
		close(outputChan)
	}
}

func (w *IoWriter) String() string {
	return "IoWriter"
}
