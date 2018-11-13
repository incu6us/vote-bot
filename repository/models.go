package repository

type Poll struct {
	CreatedAt int64    `json:"created_at"`
	Items     []string `json:"items"`
	Kind      string   `json:"kind"`
	Subject   string   `json:"subject"`
}
