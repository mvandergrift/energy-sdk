package repo

import (
	"github.com/jinzhu/gorm"
	"github.com/mvandergrift/energy-sdk/model"
)

type gormMeasureRepo struct {
	Cn *gorm.DB
}

func NewMeasureRepo(cn *gorm.DB) MeasureRepo {
	return &gormMeasureRepo{Cn: cn}
}

func (db *gormMeasureRepo) Save(measure *model.Measure) error {
	m := &model.Measure{}
	db.Cn.First(m, "external_id = ?", *measure.ExternalID)
	measure.ID = m.ID // update ID with identity from DB
	return db.Cn.Save(&measure).Error
}
