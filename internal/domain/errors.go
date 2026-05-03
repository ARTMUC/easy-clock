package domain

import "errors"

var (
	ErrEmptyName                 = errors.New("name is required")
	ErrInvalidTimezone           = errors.New("invalid IANA timezone")
	ErrInvalidHourRange          = errors.New("from_hour must be less than to_hour")
	ErrActivityOverlap           = errors.New("activities overlap within the same ring")
	ErrImageRequired             = errors.New("image_path is required")
	ErrInvalidTimeRange          = errors.New("from_time must be before to_time")
	ErrEventProfileXorActivities = errors.New("event must use either profile_id or activities, not both")
	ErrNotFound                  = errors.New("not found")
	ErrOptimisticLock            = errors.New("record was modified by another process")
)
