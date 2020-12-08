package model

type Config struct {
	LogFileNames      []string `json:"logs,omitempty"`
	RefreshRateMillis int64    `json:"refresh_rate_millis,omitempty"`
}

type Data struct {
	Content    string `json:"log_content,omitempty"`
	PageMarker int64  `json:"file_size,omitempty"`
}

type ServiceNodes struct {
	Nodes []ServiceNode `json:"nodes"`
}

type ServiceNode struct {
	NodeId string `json:"node_id"`
}
