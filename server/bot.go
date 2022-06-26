package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

// PostBotDM posts a DM as the cloud bot user.
func (p *Plugin) PostBotDM(userID string, message string) {
	p.createBotPostDM(&model.Post{
		UserId:  p.BotUserID,
		Message: message,
	}, userID)
}

func (p *Plugin) createBotPostDM(post *model.Post, userID string) {
	channel, appError := p.API.GetDirectChannel(userID, p.BotUserID)

	if appError != nil {
		p.API.LogError("Unable to get direct channel for bot err=" + appError.Error())
		return
	}
	if channel == nil {
		p.API.LogError("Could not get direct channel for bot and user_id=%s", userID)
		return
	}

	post.ChannelId = channel.Id
	_, appError = p.API.CreatePost(post)

	if appError != nil {
		p.API.LogError("Unable to create bot post DM err=" + appError.Error())
	}
}
