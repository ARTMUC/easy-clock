package domain

// ClockState is the read model returned by GET /api/clock/:clock_token.
// It is a projection — not an aggregate.
// All resolution is done in the child's LOCAL timezone before this struct is populated.
type ClockState struct {
	Empty          bool
	ProfileID      string
	ProfileName    string
	ProfileColor   string
	ActiveActivity *Activity // activity whose FromHour–ToHour window contains the current local hour
	AllActivities  []Activity
}
