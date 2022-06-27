package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Reminder struct {
	ID       string `json:"id"`
	Message  string `json:"message"`
	CreateBy string `json:"create_by"`
	CreateAt int64  `json:"create_at"`
	PostID   string `json:"post_id"`
	When     int64  `json:"reminder_at"`
	Type     string `json:"remember_type"`
}

func newReminder(userID, message, postID, rememberType string, when int64) *Reminder {
	return &Reminder{
		ID:       model.NewId(),
		CreateBy: userID,
		CreateAt: model.GetMillis(),
		Message:  message,
		PostID:   postID,
		Type:     rememberType,
		When:     model.GetMillis() + when,
	}
}

func (p *Plugin) TriggerReminders() {
	reminders, err := p.listManager.GetActiveIssues()
	if err != nil {
		return
	}

	for _, reminder := range reminders {
		post, pErr := p.API.GetPost(reminder.PostID)
		if pErr != nil {
			continue
		}

		channel, cErr := p.API.GetChannel(post.ChannelId)
		if cErr != nil {
			continue
		}

		team, tErr := p.API.GetTeam(channel.TeamId)
		if tErr != nil {
			continue
		}

		if channel.IsGroupOrDirect() {
			p.PostBotDM(reminder.CreateBy, fmt.Sprintf("You requested to be reminded about [this post](/%s/pl/%s)!", team.Name, post.Id))
		} else {
			p.PostBotDM(reminder.CreateBy, fmt.Sprintf("You requested to be reminded about [this post](/%s/pl/%s) in ~%s!", team.Name, post.Id, channel.Name))
		}
		_, _ = p.listManager.RemoveIssue(reminder.ID)
	}
}
