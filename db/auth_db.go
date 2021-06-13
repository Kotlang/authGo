package db

import (
	"reflect"

	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type AuthDb struct{}

func (a *AuthDb) Login(tenant string) *LoginRepository {
	repo := odm.AbstractRepository{
		CollectionName: "login_" + tenant,
		Model:          reflect.TypeOf(models.LoginModel{}),
	}
	return &LoginRepository{repo}
}

func (a *AuthDb) Profile(tenant string) *ProfileRepository {
	repo := odm.AbstractRepository{
		CollectionName: "profile_" + tenant,
		Model:          reflect.TypeOf(models.ProfileModel{}),
	}
	return &ProfileRepository{repo}
}

func (a *AuthDb) Tenant() *TenantRepository {
	repo := odm.AbstractRepository{
		CollectionName: "tenant",
		Model:          reflect.TypeOf(models.TenantModel{}),
	}
	return &TenantRepository{repo}
}
