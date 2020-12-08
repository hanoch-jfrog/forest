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
		URL            string `json:"url"`
		Version        string `json:"version"`
		Status         string `json:"status"`
		ServiceName    string `json:"service_name"`
		ServiceID      string `json:"service_id"`
		NodeID         string `json:"node_id"`
		LastHeartbeat  int64  `json:"last_heartbeat"`
		HeartbeatStale bool   `json:"heartbeat_stale"`
		StartTime      int64  `json:"start_time"`
		StatusDetails  string `json:"status_details"`
	} `json:"nodes"`
}
