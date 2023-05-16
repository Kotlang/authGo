package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type DeviceInstanceRepository struct {
	odm.AbstractRepository[models.DeviceInstanceModel]
}
