package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type AuthDb struct{}

func (a *AuthDb) Login(tenant string) *LoginRepository {
	baseRepo := odm.AbstractRepository[models.LoginModel]{
		Database:       tenant + "_auth",
		CollectionName: "login",
	}

	return &LoginRepository{baseRepo}
}

func (a *AuthDb) Profile(tenant string) *ProfileRepository {
	baseRepo := odm.AbstractRepository[models.ProfileModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile",
	}
	return &ProfileRepository{baseRepo}
}

func (a *AuthDb) Tenant() *TenantRepository {
	baseRepo := odm.AbstractRepository[models.TenantModel]{
		Database:       "global",
		CollectionName: "tenant",
	}
	return &TenantRepository{baseRepo}
}

func (a *AuthDb) ProfileMaster(tenant string) *ProfileMasterRepository {
	baseRepo := odm.AbstractRepository[models.ProfileMasterModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile_master",
	}

	return &ProfileMasterRepository{baseRepo}
}

func (a *AuthDb) DeviceInstance(tenant string) *DeviceInstanceRepository {
	baseRepo := odm.AbstractRepository[models.DeviceInstanceModel]{
		Database:       tenant + "_auth",
		CollectionName: "device_instance",
	}

	return &DeviceInstanceRepository{baseRepo}
}
