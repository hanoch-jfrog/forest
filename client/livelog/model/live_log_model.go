package model

type Config struct {
	LogFileNames      []string `json:"logs,omitempty"`
	RefreshRateMillis int64    `json:"refresh_rate_millis,omitempty"`
}

type Data struct {
	LogFileModified int64  `json:"last_update_modified,omitempty"`
	Timestamp       int64  `json:"last_update_label,omitempty"`
	Content         string `json:"log_content,omitempty"`
	PageMarker      int64  `json:"file_size,omitempty"`
}

type ServiceNodes struct {
	Nodes []struct {
		NodeID string `json:"node_id"`
	} `json:"nodes"`
}
