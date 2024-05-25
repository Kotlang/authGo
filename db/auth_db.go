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
	baseRepo := odm.NewUnimplementedBootRepository[LoginModel](
		odm.WithDatabase(tenant+"_auth"),
		odm.WithCollectionName("login"),
	)

	return &LoginRepository{baseRepo}
}

func (a *AuthDb) Profile(tenant string) ProfileRepositoryInterface {
	baseRepo := odm.NewUnimplementedBootRepository[ProfileModel](
		odm.WithDatabase(tenant+"_auth"),
		odm.WithCollectionName("profile"),
	)
	return &ProfileRepository{baseRepo}
}

func (a *AuthDb) Tenant() TenantRepositoryInterface {
	baseRepo := odm.NewUnimplementedBootRepository[TenantModel](
		odm.WithDatabase("global"),
		odm.WithCollectionName("tenant"),
	)
	return &TenantRepository{baseRepo}
}

func (a *AuthDb) ProfileMaster(tenant string) ProfileMasterRepositoryInterface {
	baseRepo := odm.NewUnimplementedBootRepository[ProfileMasterModel](
		odm.WithDatabase(tenant+"_auth"),
		odm.WithCollectionName("profile_master"),
	)

	return &ProfileMasterRepository{baseRepo}
}

func (a *AuthDb) Lead(tenant string) LeadRepositoryInterface {
	baseRepo := odm.NewUnimplementedBootRepository[LeadModel](
		odm.WithDatabase(tenant+"_auth"),
		odm.WithCollectionName("lead"),
	)

	return &LeadRepository{baseRepo}
}
