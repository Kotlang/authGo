// Package db provides database functionalities and access methods for authentication-related data.
package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

// AuthDb represents a database structure to access authentication-related collections.
type AuthDb struct{}

// Login creates and returns a repository to access the login collection for a specific tenant.
// The database name is constructed by appending "_auth" to the tenant string.
// Returns a LoginRepository pointing to the tenant-specific "login" collection.
func (a *AuthDb) Login(tenant string) *LoginRepository {
	baseRepo := odm.AbstractRepository[models.LoginModel]{
		Database:       tenant + "_auth",
		CollectionName: "login",
	}

	return &LoginRepository{baseRepo}
}

// Profile creates and returns a repository to access the profile collection for a specific tenant.
// The database name is constructed by appending "_auth" to the tenant string.
// Returns a ProfileRepository pointing to the tenant-specific "profile" collection.
func (a *AuthDb) Profile(tenant string) *ProfileRepository {
	baseRepo := odm.AbstractRepository[models.ProfileModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile",
	}
	return &ProfileRepository{baseRepo}
}

// Tenant creates and returns a repository to access the global "tenant" collection.
// This function points to a global database named "global".
// Returns a TenantRepository pointing to the "tenant" collection in the global database.
func (a *AuthDb) Tenant() *TenantRepository {
	baseRepo := odm.AbstractRepository[models.TenantModel]{
		Database:       "global",
		CollectionName: "tenant",
	}
	return &TenantRepository{baseRepo}
}

// ProfileMaster creates and returns a repository to access the profile master collection for a specific tenant.
// The database name is constructed by appending "_auth" to the tenant string.
// Returns a ProfileMasterRepository pointing to the tenant-specific "profile_master" collection.
func (a *AuthDb) ProfileMaster(tenant string) *ProfileMasterRepository {
	baseRepo := odm.AbstractRepository[models.ProfileMasterModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile_master",
	}

	return &ProfileMasterRepository{baseRepo}
}
