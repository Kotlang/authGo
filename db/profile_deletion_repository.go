package db

import (
	"github.com/Kotlang/authGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type ProfileDeletionRepositoryInterface interface {
	odm.BootRepository[models.ProfileDeletionModel]
	GetProfileDeletionRequests(PageSize, PageNumber int64) []models.ProfileDeletionModel
}

type ProfileDeletionRepository struct {
	odm.UnimplementedBootRepository[models.ProfileDeletionModel]
}

func (p *ProfileDeletionRepository) GetProfileDeletionRequests(PageSize, PageNumber int64) []models.ProfileDeletionModel {
	skip := PageNumber * PageSize

	resultChan, errChan := p.Find(nil, nil, PageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case <-errChan:
		return []models.ProfileDeletionModel{}
	}
}
