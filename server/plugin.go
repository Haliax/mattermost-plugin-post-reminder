package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	// WSEventConfigUpdate is the WebSocket event to update the configurations on webapp
	WSEventConfigUpdate = "config_update"
)

// ListManager represents the logic on the lists
type ListManager interface {
	AddIssue(userID, message, postID, rememberType string, when int64) (*Reminder, error)
	GetActiveIssues() ([]*Reminder, error)
	GetUserName(userID string) string
	RemoveIssue(issueID string) (*Reminder, error)
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	BotUserID string

	running bool

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	listManager ListManager
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

	p.Run()

	p.listManager = NewListManager(p.API)

	return nil
}

func (p *Plugin) OnDeactivate() error {
	return nil
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/add":
		p.handleAdd(w, r)
	default:
		http.NotFound(w, r)
	}
}

type addAPIRequest struct {
	Message      string `json:"message"`
	RememberAt   string `json:"remember_at"`
	PostID       string `json:"post_id"`
	RememberType string `json:"reminder_type"`
}

func (p *Plugin) handleAdd(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-ID")
	if userID == "" {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	var addRequest *addAPIRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&addRequest)
	if err != nil {
		p.API.LogError("Unable to decode JSON err=" + err.Error())
		p.handleErrorWithCode(w, http.StatusBadRequest, "Unable to decode JSON", err)
		return
	}

	var time, convErr = strconv.ParseInt(addRequest.RememberAt, 10, 64)
	if convErr != nil {
		p.API.LogError("Unable to add issue err=" + convErr.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to add issue", convErr)
		return
	}

	_, err = p.listManager.AddIssue(userID, addRequest.Message, addRequest.PostID, addRequest.RememberType, time)
	if err != nil {
		p.API.LogError("Unable to add issue err=" + err.Error())
		p.handleErrorWithCode(w, http.StatusInternalServerError, "Unable to add issue", err)
		return
	}
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

func (p *Plugin) handleErrorWithCode(w http.ResponseWriter, code int, errTitle string, err error) {
	w.WriteHeader(code)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   errTitle,
		Details: err.Error(),
	})
	_, _ = w.Write(b)
}
