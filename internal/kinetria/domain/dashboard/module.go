package dashboard

import "go.uber.org/fx"

var Module = fx.Module("dashboard",
	fx.Provide(
		NewGetUserProfileUC,
		NewGetTodayWorkoutUC,
		NewGetWeekProgressUC,
		NewGetWeekStatsUC,
	),
)
