package repo

import "github.com/mvandergrift/energy-sdk/model"

type WorkoutRepo interface {
	Save(*model.Workout) error
}

type MeasureRepo interface {
	Save(*model.Measure) error
}
