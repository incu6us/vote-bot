package domain

import "fmt"

type Poll struct {
	Subject   string              `json:"subject"`
	CreatedAt int64               `json:"created_at"`
	Items     []string            `json:"items"`
	CreatedBy string              `json:"created_by"`
	Votes     map[string][]string `json:"votes"`
}

func (p Poll) String() string {
	return fmt.Sprintf("{ Subject: '%s', CreatedAt: %d, Items: %q, CreatedBy: '%s', Votes: %+v",
		p.Subject, p.CreatedAt, p.Items, p.CreatedBy, p.Votes)
}
