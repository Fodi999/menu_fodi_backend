package dto

type CreateBusinessRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	City        string `json:"city"`
	OwnerID     string `json:"owner_id"`
}

type UpdateBusinessRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	City        *string `json:"city"`
	IsActive    *bool   `json:"isActive"`
}
