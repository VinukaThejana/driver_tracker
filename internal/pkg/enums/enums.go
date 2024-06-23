// Package enums contains various enumerations
package enums

// Env is a struct that is used to contain various environments
type Env string

const (
	// Dev represents the local development environment
	Dev Env = "dev"
	// Stg represents the remote testing environment (this is basically used for testing (unit & intergration))
	Stg Env = "stg"
	// Prd represents the production environment
	Prd Env = "prd"
)
