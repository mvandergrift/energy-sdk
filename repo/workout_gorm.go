package repo

import (
	"github.com/jinzhu/gorm"
	"github.com/mvandergrift/energy-sdk/model"
)

type gormCompanyRepo struct {
	Cn *gorm.DB
}

func NewWorkoutRepo(cn *gorm.DB) WorkoutRepo {
	return &gormCompanyRepo{Cn: cn}
}

func (db *gormCompanyRepo) Save(workout *model.Workout) error {
	w := &model.Workout{} // initialize before use to allow use of TableName receiver
	db.Cn.First(w, "external_id = ?", *workout.ExternalID)
	workout.ID = w.ID // update workout.ID with identity from DB
	return db.Cn.Save(&workout).Error
}
