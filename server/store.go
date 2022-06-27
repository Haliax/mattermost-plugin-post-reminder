package main

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

const (
	// StoreRetries is the number of retries to use when storing lists fails on a race
	StoreRetries = 3
	// StoreListKey is the key used to store lists in the plugin KV store. Still "order" for backwards compatibility.
	StoreListKey = "order"
	// StoreIssueKey is the key used to store issues in the plugin KV store. Still "item" for backwards compatibility.
	StoreIssueKey = "item"
	// StoreReminderKey is the key used to store the last time a user was reminded
	// StoreReminderKey = "reminder"
	// StoreReminderEnabledKey is the key used to store the user preference of auto daily reminder
	// StoreReminderEnabledKey = "reminder_enabled"

	// StoreAllowIncomingTaskRequestsKey is the key used to store user preference for wallowing any incoming requests
	// StoreAllowIncomingTaskRequestsKey = "allow_incoming_task"
)

type ReminderRef struct {
	ReminderID string `json:"issue_id"`
}

func listKey(userID string, listID string) string {
	return fmt.Sprintf("%s_%s%s", StoreListKey, userID, listID)
}

func issueKey(issueID string) string {
	return fmt.Sprintf("%s_%s", StoreIssueKey, issueID)
}

type listStore struct {
	api plugin.API
}

func (l listStore) AddIssue(issue *Reminder) error {
	jsonIssue, jsonErr := json.Marshal(issue)
	if jsonErr != nil {
		return jsonErr
	}

	appErr := l.api.KVSet(issueKey(issue.ID), jsonIssue)
	if appErr != nil {
		return errors.New(appErr.Error())
	}

	return nil
}

func (l *listStore) RemoveIssue(issueID string) error {
	appErr := l.api.KVDelete(issueKey(issueID))
	if appErr != nil {
		return errors.New(appErr.Error())
	}

	return nil
}

func (l listStore) AddReference(userID, issueID, listID string) error {
	for i := 0; i < StoreRetries; i++ {
		list, originalJSONList, err := l.getList(userID, listID)
		if err != nil {
			return err
		}

		for _, ir := range list {
			if ir.ReminderID == issueID {
				return errors.New("issue id already exists in list")
			}
		}

		list = append(list, &ReminderRef{
			ReminderID: issueID,
		})

		ok, err := l.saveList(userID, listID, list, originalJSONList)
		if err != nil {
			return err
		}

		// If err is nil but ok is false, then something else updated the installs between the get and set above
		// so we need to try again, otherwise we can return
		if ok {
			return nil
		}
	}

	return errors.New("unable to store installation")
}

// NewListStore creates a new listStore
func NewListStore(api plugin.API) ListStore {
	return &listStore{
		api: api,
	}
}

func (l *listStore) GetList(userID, listID string) ([]*ReminderRef, error) {
	irs, _, err := l.getList(userID, listID)
	return irs, err
}

func (l *listStore) getList(userID, listID string) ([]*ReminderRef, []byte, error) {
	originalJSONList, err := l.api.KVGet(listKey(userID, listID))
	if err != nil {
		return nil, nil, err
	}

	if originalJSONList == nil {
		return []*ReminderRef{}, originalJSONList, nil
	}

	var list []*ReminderRef
	jsonErr := json.Unmarshal(originalJSONList, &list)
	if jsonErr != nil {
		return l.legacyIssueRef(userID, listID)
	}

	return list, originalJSONList, nil
}

func (l *listStore) saveList(userID, listID string, list []*ReminderRef, originalJSONList []byte) (bool, error) {
	newJSONList, jsonErr := json.Marshal(list)
	if jsonErr != nil {
		return false, jsonErr
	}

	ok, appErr := l.api.KVCompareAndSet(listKey(userID, listID), originalJSONList, newJSONList)
	if appErr != nil {
		return false, errors.New(appErr.Error())
	}

	return ok, nil
}

func (l *listStore) legacyIssueRef(userID, listID string) ([]*ReminderRef, []byte, error) {
	originalJSONList, err := l.api.KVGet(listKey(userID, listID))
	if err != nil {
		return nil, nil, err
	}

	if originalJSONList == nil {
		return []*ReminderRef{}, originalJSONList, nil
	}

	var list []string
	jsonErr := json.Unmarshal(originalJSONList, &list)
	if jsonErr != nil {
		return nil, nil, jsonErr
	}

	newList := []*ReminderRef{}
	for _, v := range list {
		newList = append(newList, &ReminderRef{ReminderID: v})
	}

	return newList, originalJSONList, nil
}
