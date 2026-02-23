package vos

type MuscleGroup string

const (
	MuscleGroupChest     MuscleGroup = "chest"
	MuscleGroupBack      MuscleGroup = "back"
	MuscleGroupLegs      MuscleGroup = "legs"
	MuscleGroupShoulders MuscleGroup = "shoulders"
	MuscleGroupArms      MuscleGroup = "arms"
	MuscleGroupCore      MuscleGroup = "core"
	MuscleGroupFullBody  MuscleGroup = "full_body"
)

func (m MuscleGroup) String() string {
	return string(m)
}

func (m MuscleGroup) IsValid() bool {
	switch m {
	case MuscleGroupChest, MuscleGroupBack, MuscleGroupLegs, MuscleGroupShoulders,
		MuscleGroupArms, MuscleGroupCore, MuscleGroupFullBody:
		return true
	}
	return false
}
