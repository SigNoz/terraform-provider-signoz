package model

const ()

var ()

// Dashboard model.
type Dashboard struct {
	UUID     string `json:"uuid"`
	CreateAt string `json:"createAt,omitempty"`
	CreateBy string `json:"createBy,omitempty"`
	UpdateAt string `json:"updateAt,omitempty"`
	UpdateBy string `json:"updateBy,omitempty"`

	Title       string `json:"title"`
	Description string `json:"description"`

	Tags []string `json:"tags"`

	Layout  []string `json:"layout"`
	Widgets []string `json:"widgets"`

	Variables map[string]interface{} `json:"variables"`
}
