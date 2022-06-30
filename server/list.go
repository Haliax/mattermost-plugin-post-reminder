package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// ListStore represents the KVStore operations for lists
type ListStore interface {
	AddReminder(issue *Reminder) error
	GetReminder(issueID string) (*Reminder, error)
	RemoveReminder(issueID string) error
	GetAndRemoveReminder(issueID string) (*Reminder, error)
	GetList() ([]*ReminderRef, error)

	AddReference(remindDate int64, issueID string) error
	RemoveReference(issueID string) error
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
	issue := newReminder(userID, message, postID, when)

	if err := l.store.AddReminder(issue); err != nil {
		return nil, err
	}

	if err := l.store.AddReference(issue.When, issue.ID); err != nil {
		if rollbackError := l.store.RemoveReminder(issue.ID); rollbackError != nil {
			l.api.LogError("cannot rollback issue after add error, Err=", err.Error())
		}
		return nil, err
	}

	return issue, nil
}

func (l *listManager) RemoveIssue(issueID string) (outIssue *Reminder, outErr error) {
	ir, _ := l.store.GetReminder(issueID)
	if ir == nil {
		return nil, fmt.Errorf("cannot find element")
	}

	if err := l.store.RemoveReference(issueID); err != nil {
		return nil, err
	}

	issue, err := l.store.GetAndRemoveReminder(issueID)
	if err != nil {
		l.api.LogError("cannot remove issue, Err=", err.Error())
	}

	return issue, nil
}

func (l *listManager) GetActiveIssues() ([]*Reminder, error) {
	refs, err := l.store.GetList()
	if err != nil {
		return nil, err
	}

	now := model.GetMillis()

	reminders := []*Reminder{}
	for _, ref := range refs {
		if ref.ReminderDate > now {
			continue
		}

		reminder, err := l.store.GetReminder(ref.ReminderID)
		if err != nil {
			continue
		}

		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

func (l *listManager) GetUserName(userID string) string {
	user, err := l.api.GetUser(userID)
	if err != nil {
		return "Someone"
	}
	return user.Username
}
