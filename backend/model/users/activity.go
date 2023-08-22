package users

import (
	"github.com/pocketbase/pocketbase/models"
)

// activity to category mapping
type AvailableActivity string

const (
	AvailableActivityOptimization AvailableActivity = "optimization"
	AvailableActivityUpdateEtf    AvailableActivity = "update_etf"
)

type UserActivity struct {
	models.BaseModel

	RelatedUser  string            `db:"related_user" json:"related_user"`
	ActivityType AvailableActivity `db:"activity_type" json:"activity_type"`
	Details      []byte            `db:"details" json:"details"`
}

func (m *UserActivity) TableName() string {
	return "user_activity" // the name of your collection
}

var _ models.Model = (*UserActivity)(nil)
