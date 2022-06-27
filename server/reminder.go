package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type Reminder struct {
	ID       string `json:"id"`
	Message  string `json:"message"`
	CreateAt int64  `json:"create_at"`
	PostID   string `json:"post_id"`
	When     int64  `json:"reminder_at"`
}

func newReminder(message string, postID string, when int64) *Reminder {
	return &Reminder{
		ID:       model.NewId(),
		CreateAt: model.GetMillis(),
		Message:  message,
		PostID:   postID,
		When:     model.GetMillis() + when,
	}
}
