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
	// StoreListKey is the key used to store lists in the plugin KV store.
	StoreListKey = "reminders"
	// StoreIssueKey is the key used to store issues in the plugin KV store.
	StoreIssueKey = "item"
)

type ReminderRef struct {
	ReminderID   string `json:"issue_id"`
	ReminderDate int64  `json:"reminder_date"`
}

func listKey() string {
	return StoreListKey
}

func issueKey(issueID string) string {
	return fmt.Sprintf("%s_%s", StoreIssueKey, issueID)
}

type listStore struct {
	api plugin.API
}

func (l *listStore) AddReminder(issue *Reminder) error {
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

func (l *listStore) GetReminder(issueID string) (*Reminder, error) {
	originalJSONIssue, appErr := l.api.KVGet(issueKey(issueID))
	if appErr != nil {
		return nil, errors.New(appErr.Error())
	}

	if originalJSONIssue == nil {
		return nil, errors.New("cannot find issue")
	}

	var issue *Reminder
	err := json.Unmarshal(originalJSONIssue, &issue)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

func (l *listStore) RemoveReminder(issueID string) error {
	appErr := l.api.KVDelete(issueKey(issueID))
	if appErr != nil {
		return errors.New(appErr.Error())
	}

	return nil
}

func (l *listStore) GetAndRemoveReminder(issueID string) (*Reminder, error) {
	issue, err := l.GetReminder(issueID)
	if err != nil {
		return nil, err
	}

	err = l.RemoveReminder(issueID)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

func (l listStore) AddReference(remindDate int64, issueID string) error {
	for i := 0; i < StoreRetries; i++ {
		list, originalJSONList, err := l.getList()
		if err != nil {
			return err
		}

		for _, ir := range list {
			if ir.ReminderID == issueID {
				return errors.New("issue id already exists in list")
			}
		}

		list = append(list, &ReminderRef{
			ReminderID:   issueID,
			ReminderDate: remindDate,
		})

		ok, err := l.saveList(list, originalJSONList)
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

func (l *listStore) RemoveReference(issueID string) error {
	for i := 0; i < StoreRetries; i++ {
		list, originalJSONList, err := l.getList()
		if err != nil {
			return err
		}

		found := false
		for i, ir := range list {
			if ir.ReminderID == issueID {
				list = append(list[:i], list[i+1:]...)
				found = true
			}
		}

		if !found {
			return errors.New("cannot find issue")
		}

		ok, err := l.saveList(list, originalJSONList)
		if err != nil {
			return err
		}

		// If err is nil but ok is false, then something else updated between the get and set above
		// so we need to try again, otherwise we can return
		if ok {
			return nil
		}
	}

	return errors.New("unable to store list")
}

// NewListStore creates a new listStore
func NewListStore(api plugin.API) ListStore {
	return &listStore{
		api: api,
	}
}

func (l *listStore) GetList() ([]*ReminderRef, error) {
	irs, _, err := l.getList()
	return irs, err
}

func (l *listStore) getList() ([]*ReminderRef, []byte, error) {
	originalJSONList, err := l.api.KVGet(listKey())
	if err != nil {
		return nil, nil, err
	}

	if originalJSONList == nil {
		return []*ReminderRef{}, originalJSONList, nil
	}

	var list []*ReminderRef
	jsonErr := json.Unmarshal(originalJSONList, &list)
	if jsonErr != nil {
		return l.legacyIssueRef()
	}

	return list, originalJSONList, nil
}

func (l *listStore) saveList(list []*ReminderRef, originalJSONList []byte) (bool, error) {
	newJSONList, jsonErr := json.Marshal(list)
	if jsonErr != nil {
		return false, jsonErr
	}

	ok, appErr := l.api.KVCompareAndSet(listKey(), originalJSONList, newJSONList)
	if appErr != nil {
		return false, errors.New(appErr.Error())
	}

	return ok, nil
}

func (l *listStore) legacyIssueRef() ([]*ReminderRef, []byte, error) {
	originalJSONList, err := l.api.KVGet(listKey())
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
