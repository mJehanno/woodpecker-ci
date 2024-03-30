package tracer

import (
	"strconv"
	"time"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
)

var CustomTracer = pipeline.TraceFunc(func(state *pipeline.State) error {
	if state.Process.Exited {
		return nil
	}
	if state.Pipeline.Step.Environment == nil {
		return nil
	}
	state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "success"
	state.Pipeline.Step.Environment["CI_PIPELINE_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
	state.Pipeline.Step.Environment["CI_PIPELINE_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

	state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "success"
	state.Pipeline.Step.Environment["CI_STEP_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
	state.Pipeline.Step.Environment["CI_STEP_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

	if state.Pipeline.Error != nil {
		state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "failure"
		state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "failure"
	}
	return nil
})
