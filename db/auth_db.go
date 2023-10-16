package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type AuthDbInterface interface {
	Login(tenant string) LoginRepositoryInterface
	Profile(tenant string) ProfileRepositoryInterface
	Tenant() TenantRepositoryInterface
	ProfileMaster(tenant string) ProfileMasterRepositoryInterface
}

type AuthDb struct{}

func (a *AuthDb) Login(tenant string) LoginRepositoryInterface {
	baseRepo := odm.AbstractRepository[models.LoginModel]{
		Database:       tenant + "_auth",
		CollectionName: "login",
	}

	return &LoginRepository{baseRepo}
}

func (a *AuthDb) Profile(tenant string) ProfileRepositoryInterface {
	baseRepo := odm.AbstractRepository[models.ProfileModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile",
	}
	return &ProfileRepository{baseRepo}
}

func (a *AuthDb) Tenant() TenantRepositoryInterface {
	baseRepo := odm.AbstractRepository[models.TenantModel]{
		Database:       "global",
		CollectionName: "tenant",
	}
	return &TenantRepository{baseRepo}
}

func (a *AuthDb) ProfileMaster(tenant string) ProfileMasterRepositoryInterface {
	baseRepo := odm.AbstractRepository[models.ProfileMasterModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile_master",
	}

	return &ProfileMasterRepository{baseRepo}
}
