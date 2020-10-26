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
	err := db.Cn.First(w, "(user_id = ? AND activity_id = ? AND attribute_id = ? AND model_id = ? AND start_time = ?) OR (external_id = ?)",
		workout.UserID,
		workout.ActivityID,
		workout.AttributeID,
		workout.ModelID,
		workout.StartTime,
		*workout.ExternalID).Error

	if err != nil && err.Error() != "record not found" {
		return err
	}

	workout.ID = w.ID // update workout.ID with identity from DB
	return db.Cn.Save(&workout).Error
}
