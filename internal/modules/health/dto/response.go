package dto

// HealthResponse - ответ health check
type HealthResponse struct {
	Status   string                 `json:"status"`
	Data     map[string]interface{} `json:"data"`
}

// HealthData - детали health check
type HealthData struct {
	Service  string `json:"service"`
	Database string `json:"database"`
	Version  string `json:"version,omitempty"`
}
