package web

type HistorySlot struct {
	Time   string           `json:"time"`
	Counts map[string]int64 `json:"counts"`
	Total  int64            `json:"total"`
}