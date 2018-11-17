package domain

type Poll struct {
	Subject     string              `json:"subject"`
	CreatedAt   int64               `json:"created_at"`
	Items       []string            `json:"items"`
	CreatedBy   string              `json:"created_by"`
	IsPublished bool                `json:"is_published"`
	Votes       map[string][]string `json:"votes"`
}
