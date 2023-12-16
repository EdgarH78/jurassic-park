package models

import "errors"

var (
	EntityNotFound             = errors.New("Entity Not Found")
	InvalidDinosaurSpecies     = errors.New("Invalid Dinosaur Species")
	EntityAlreadyExists        = errors.New("Entity already exists")
	CageCapacityExceeded       = errors.New("Cage capacity exceeded")
	IncompatibleSpecies        = errors.New("Incompatible Species")
	IncompatibleCagePowerState = errors.New("Incompatible Cage Power State")
)
