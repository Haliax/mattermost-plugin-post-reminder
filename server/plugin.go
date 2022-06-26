package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

const (
	// WSEventRefresh is the WebSocket event for refreshing the Todo list
	// WSEventRefresh = "refresh"

	// WSEventConfigUpdate is the WebSocket event to update the Todo list's configurations on webapp
	WSEventConfigUpdate = "config_update"
)

// ListManager represents the logic on the lists
// type ListManager interface {
// }

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	BotUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// listManager ListManager
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		return err
	}

	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    "post-reminder",
		DisplayName: "Post Reminder",
		Description: "Created by the Post Reminder plugin.",
	}, []plugin.EnsureBotOption{
		plugin.ProfileImagePath("public/app-bar-icon.png"),
	}...)
	if err != nil {
		return errors.Wrap(err, "failed to ensure post reminder bot account")
	}
	p.BotUserID = botID

	return p.API.RegisterCommand(getCommand())
}

func (p *Plugin) OnDeactivate() error {
	return nil
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// Publish a WebSocket event to update the client config of the plugin on the webapp end.
func (p *Plugin) sendConfigUpdateEvent() {
	clientConfigMap := map[string]interface{}{
		"hide_team_sidebar": p.configuration.HideTeamSidebar,
	}

	p.API.PublishWebSocketEvent(
		WSEventConfigUpdate,
		clientConfigMap,
		&model.WebsocketBroadcast{},
	)
}
