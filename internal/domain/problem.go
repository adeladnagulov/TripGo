package domain

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Code     string `json:"code,omitempty"`
	Instance string `json:"instance,omitempty"`
}
