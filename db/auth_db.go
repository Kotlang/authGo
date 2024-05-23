package db

import (
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type AuthDbInterface interface {
	Login(tenant string) LoginRepositoryInterface
	Profile(tenant string) ProfileRepositoryInterface
	Tenant() TenantRepositoryInterface
	ProfileMaster(tenant string) ProfileMasterRepositoryInterface
	Lead(tenant string) LeadRepositoryInterface
}

type AuthDb struct{}

func ProvideAuthDb() AuthDbInterface {
	return &AuthDb{}
}

func (a *AuthDb) Login(tenant string) LoginRepositoryInterface {
	baseRepo := odm.UnimplementedBootRepository[LoginModel]{
		Database:       tenant + "_auth",
		CollectionName: "login",
	}

	return &LoginRepository{baseRepo}
}

func (a *AuthDb) Profile(tenant string) ProfileRepositoryInterface {
	baseRepo := odm.UnimplementedBootRepository[ProfileModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile",
	}
	return &ProfileRepository{baseRepo}
}

func (a *AuthDb) Tenant() TenantRepositoryInterface {
	baseRepo := odm.UnimplementedBootRepository[TenantModel]{
		Database:       "global",
		CollectionName: "tenant",
	}
	return &TenantRepository{baseRepo}
}

func (a *AuthDb) ProfileMaster(tenant string) ProfileMasterRepositoryInterface {
	baseRepo := odm.UnimplementedBootRepository[ProfileMasterModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile_master",
	}

	return &ProfileMasterRepository{baseRepo}
}

func (a *AuthDb) Lead(tenant string) LeadRepositoryInterface {
	baseRepo := odm.UnimplementedBootRepository[LeadModel]{
		Database:       tenant + "_auth",
		CollectionName: "lead",
	}

	return &LeadRepository{baseRepo}
}
