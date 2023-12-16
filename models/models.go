package models

type Cage struct {
	Label        string `json:"label"`
	Occupancy    int    `json:"occupancy"`
	MaxOccupancy int    `json:"maxOccupancy"`
	HasPower     bool   `json:"hasPower"`
}

type Dinosaur struct {
	Name    string  `json:"name"`
	Species string  `json:"species"`
	Diet    string  `json:"diet"`
	Cage    *string `json:"cage,omitempty"`
}

type AddDinosaurToCageRequest struct {
	Name string `json:"name"`
}

type UpdateCagePowerStatusRequest struct {
	HasPower bool `json:"hasPower"`
}

type DinosaurFilter struct {
	Species             *string
	Diet                *string
	NeedsCageAssignment *bool
}

type CageFilter struct {
	HasPower *bool
}

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}
