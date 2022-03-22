package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type AuthDb struct{}

func (a *AuthDb) Login(tenant string) *LoginRepository {
	repo := odm.AbstractRepository[models.LoginModel]{
		Database:       tenant + "_auth",
		CollectionName: "login",
	}
	return &LoginRepository{repo}
}

func (a *AuthDb) Profile(tenant string) *ProfileRepository {
	repo := odm.AbstractRepository[models.ProfileModel]{
		Database:       tenant + "_auth",
		CollectionName: "profile",
	}
	return &ProfileRepository{repo}
}

func (a *AuthDb) Tenant() *TenantRepository {
	repo := odm.AbstractRepository[models.TenantModel]{
		Database:       "global",
		CollectionName: "tenant",
	}
	return &TenantRepository{repo}
}
