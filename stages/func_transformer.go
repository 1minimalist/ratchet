package stages

import "github.com/DailyBurn/ratchet/data"

// FuncTransformer is a stage that executes the given function on each data
// payload, sending the resuling data to the next stage of processing.
//
// While FuncTransformer is useful for simple data transformation, more
// complicated tasks justify building a custom implementation of PipelineStage.
type FuncTransformer struct {
	foo  func(d data.JSON) data.JSON
	Name string // can be set for more useful log output
}

func NewFuncTransformer(foo func(d data.JSON) data.JSON) *FuncTransformer {
	return &FuncTransformer{foo: foo}
}

func (t *FuncTransformer) ProcessData(d data.JSON, outputChan chan data.JSON, killChan chan error) {
	outputChan <- t.foo(d)
}

func (t *FuncTransformer) Finish(outputChan chan data.JSON, killChan chan error) {
	if outputChan != nil {
		close(outputChan)
	}
}

func (t *FuncTransformer) String() string {
	if t.Name != "" {
		return t.Name
	}
	return "FuncTransformer"
}
