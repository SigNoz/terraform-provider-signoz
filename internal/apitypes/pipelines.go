package apitypes

// Hand-written wire types for the logs pipelines API. The API lives in
// SigNoz's legacy query-service handler and is absent from the generated
// OpenAPI spec, so it has no generated client; requests go through
// apiclients.WrappedClient.Do. Field shapes mirror
// github.com/SigNoz/signoz/pkg/types/pipelinetypes.

// PipelinetypesPostablePipelines is the POST /api/v1/logs/pipelines body. The
// posted list fully replaces the previous pipeline set as a new config version.
type PipelinetypesPostablePipelines struct {
	Pipelines []PipelinetypesPostablePipeline `json:"pipelines"`
}

type PipelinetypesPostablePipeline struct {
	OrderID     int                             `json:"orderId"`
	Name        string                          `json:"name"`
	Alias       string                          `json:"alias"`
	Description *string                         `json:"description,omitempty"`
	Enabled     bool                            `json:"enabled"`
	Filter      *V3FilterSet                    `json:"filter"`
	Config      []PipelinetypesPipelineOperator `json:"config"`
}

type PipelinetypesGettablePipeline struct {
	Id          string                          `json:"id"`
	OrderID     int                             `json:"orderId"`
	Enabled     bool                            `json:"enabled"`
	Name        string                          `json:"name"`
	Alias       string                          `json:"alias"`
	Description *string                         `json:"description,omitempty"`
	Filter      *V3FilterSet                    `json:"filter"`
	Config      []PipelinetypesPipelineOperator `json:"config"`
}

// PipelinetypesPipelineOperator is one processor in a pipeline's config. Which
// optional fields apply depends on Type (grok_parser wants Pattern,
// regex_parser wants Regex, …) — the server validates the combination.
type PipelinetypesPipelineOperator struct {
	Type       string                  `json:"type"`
	ID         string                  `json:"id,omitempty"`
	Output     string                  `json:"output,omitempty"`
	OnError    *string                 `json:"on_error,omitempty"`
	If         *string                 `json:"if,omitempty"`
	OrderID    int                     `json:"orderId"`
	Enabled    bool                    `json:"enabled"`
	Name       *string                 `json:"name,omitempty"`
	ParseTo    *string                 `json:"parse_to,omitempty"`
	Pattern    *string                 `json:"pattern,omitempty"`
	Regex      *string                 `json:"regex,omitempty"`
	ParseFrom  *string                 `json:"parse_from,omitempty"`
	TraceId    *PipelinetypesParseFrom `json:"trace_id,omitempty"`
	SpanId     *PipelinetypesParseFrom `json:"span_id,omitempty"`
	TraceFlags *PipelinetypesParseFrom `json:"trace_flags,omitempty"`
	Field      *string                 `json:"field,omitempty"`
	Value      *string                 `json:"value,omitempty"`
	From       *string                 `json:"from,omitempty"`
	To         *string                 `json:"to,omitempty"`
	Expr       *string                 `json:"expr,omitempty"`
	Fields     *[]string               `json:"fields,omitempty"`
	Default    *string                 `json:"default,omitempty"`
	Layout     *string                 `json:"layout,omitempty"`
	LayoutType *string                 `json:"layout_type,omitempty"`

	EnableFlattening      bool                 `json:"enable_flattening,omitempty"`
	EnablePaths           bool                 `json:"enable_paths,omitempty"`
	PathPrefix            *string              `json:"path_prefix,omitempty"`
	Mapping               *map[string][]string `json:"mapping,omitempty"`
	OverwriteSeverityText bool                 `json:"overwrite_text,omitempty"`
}

type PipelinetypesParseFrom struct {
	ParseFrom string `json:"parse_from"`
}

// V3FilterSet is the pipeline filter, from
// github.com/SigNoz/signoz/pkg/query-service/model/v3.
type V3FilterSet struct {
	Operator string         `json:"op,omitempty"`
	Items    []V3FilterItem `json:"items"`
}

type V3FilterItem struct {
	Key      V3AttributeKey `json:"key"`
	Value    interface{}    `json:"value"`
	Operator string         `json:"op"`
}

type V3AttributeKey struct {
	Key      string `json:"key"`
	DataType string `json:"dataType,omitempty"`
	Type     string `json:"type,omitempty"`
	IsColumn bool   `json:"isColumn,omitempty"`
	IsJSON   bool   `json:"isJSON,omitempty"`
}

// LogparsingpipelinePipelinesResponse is the `data` payload returned by both
// GET /api/v1/logs/pipelines/{version} and POST /api/v1/logs/pipelines.
// Version metadata is flattened in when a config version exists and absent
// before the first save.
type LogparsingpipelinePipelinesResponse struct {
	Version   *int                            `json:"version,omitempty"`
	Pipelines []PipelinetypesGettablePipeline `json:"pipelines"`
}

// LogparsingpipelineApiResponse is the legacy `{status, data}` success
// envelope the pipelines endpoints respond with.
type LogparsingpipelineApiResponse struct {
	Status string                              `json:"status"`
	Data   LogparsingpipelinePipelinesResponse `json:"data"`
}
