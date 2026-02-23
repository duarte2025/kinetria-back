package vos

type ExerciseCategory string

const (
	ExerciseCategoryStrength    ExerciseCategory = "strength"
	ExerciseCategoryCardio      ExerciseCategory = "cardio"
	ExerciseCategoryFlexibility ExerciseCategory = "flexibility"
	ExerciseCategoryBalance     ExerciseCategory = "balance"
)

func (c ExerciseCategory) String() string {
	return string(c)
}

func (c ExerciseCategory) IsValid() bool {
	switch c {
	case ExerciseCategoryStrength, ExerciseCategoryCardio, ExerciseCategoryFlexibility, ExerciseCategoryBalance:
		return true
	}
	return false
}
