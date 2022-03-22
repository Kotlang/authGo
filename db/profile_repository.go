package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type ProfileRepository struct {
	odm.AbstractRepository[models.ProfileModel]
}
