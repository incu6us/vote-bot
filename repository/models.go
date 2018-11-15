package repository

type Poll struct {
	Subject   string   `json:"subject"`
	CreatedAt int64    `json:"created_at"`
	Items     []string `json:"items"`
	Owner     string   `json:"owner"`
}
