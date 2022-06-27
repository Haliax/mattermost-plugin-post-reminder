package main

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	// MyListKey is the key used to store the list of the owned todos
	MyListKey = ""
)

// ListStore represents the KVStore operations for lists
type ListStore interface {
	// Issue related function
	AddIssue(issue *Reminder) error
	RemoveIssue(issueID string) error

	// Issue References related functions

	// AddReference creates a new IssueRef with the issueID, foreignUSerID and foreignIssueID, and stores it
	// on the listID for userID.
	AddReference(userID, issueID, listID string) error
}

type listManager struct {
	store ListStore
	api   plugin.API
}

// NewListManager creates a new listManager
func NewListManager(api plugin.API) ListManager {
	return &listManager{
		store: NewListStore(api),
		api:   api,
	}
}

func (l *listManager) AddIssue(userID, message, postID string, when int64) (*Reminder, error) {
	issue := newReminder(message, postID, when)

	if err := l.store.AddIssue(issue); err != nil {
		return nil, err
	}

	if err := l.store.AddReference(userID, issue.ID, MyListKey); err != nil {
		if rollbackError := l.store.RemoveIssue(issue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback issue after add error, Err=", err.Error())
		}
		return nil, err
	}

	return issue, nil
}

func (l *listManager) GetUserName(userID string) string {
	user, err := l.api.GetUser(userID)
	if err != nil {
		return "Someone"
	}
	return user.Username
}
