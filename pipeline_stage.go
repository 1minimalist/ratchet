package ratchet

import (
	"github.com/DailyBurn/ratchet/data"
)

// PipelineStage is the interface used to process data within a Pipeline.
// The Pipeline is responsible for passing data between each stage,
// and each stage will receive data in it's ProcessData function and send along
// a new data payload on it's output channel.
//
// When the stage is finished sending all of it's data, it should close the output
// channel (the exception being when it is the final stage in the pipeline, where outputChan
// will be nil). This will almost always occur in the stage's Finish() function, which is called
// after the previous stage has closed it's ouput channel (i.e. is finished sending data).
// Finish() can also be used to send a final Data payload in use-cases where the stage
// is batching up multiple Data payloads and needs to send an aggregated Data object
// once all processing is complete.
//
// If an unexpected error occurs, it should be sent to the killChan to halt
// pipeline execution.
//
// To outline:
//
//   * Initial PipelineStage:
//     * Will receive a "GO" in ProcessData when the Pipeline is Run.
//     * Should send one or more data payloads on it's outputChan.
//     * Should close(outputChan) when done sending data (typically done in Finish).
//   * Intermediate PipelineStage:
//     * Will receive a call to ProcessData for each data payload sent from the preceding PipelineStage.
//     * Should send one or more data payloads on it's outputChan.
//     * Will receive a call to Finish when the preceding stage is completed.
//     * Should close(outputChan) when done sending data (typically done in Finish).
//   * Final PipelineStage:
//     * Will receive a call to ProcessData for each data payload sent from the preceding PipelineStage.
//     * Should handle writing data out to a final location, but should NOT send to outputChan (it will be nil).
//     * Will receive a call to Finish when the preceding stage is completed.
//     * Should close(outputChan) when done handling data (typically done in Finish).
type PipelineStage interface {
	// ProcessData will be called for each data sent on the previous stage's outputChan
	ProcessData(d data.JSON, outputChan chan data.JSON, killChan chan error)

	// Finish will be called after the previous stage has closed it's outputChan
	// and won't be sending any more data. So, Finish() will be be called after
	// the last call to ProcessData().
	//
	// *Note: If the PipelineStage instance receiving the Finish() call is the last
	// stage in the pipeline, outputChan will be nil.*
	Finish(outputChan chan data.JSON, killChan chan error)
}

// ConcurrentPipelineStage is a PipelineStage that also defines
// a level of concurrency. For example, if Concurrency() returns 2,
// then the pipeline will allow the stage to execute 2 ProcessData()
// calls concurrently.
//
// NOTE: The order of data processing is maintained, meaning that
// when a stage receives ProcessData calls with d1, d2, ..., the resulting data
// payloads sent on the outputChan will be sent in the same order as received.
type ConcurrentPipelineStage interface {
	PipelineStage
	Concurrency() int
}

// IsConcurrent returns true if the given PipelineStage implements ConcurrentPipelineStage
func IsConcurrent(s PipelineStage) bool {
	_, ok := interface{}(s).(ConcurrentPipelineStage)
	return ok
}
