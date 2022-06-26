package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func getHelp() string {
	return `Available Commands:

add [message]
	Adds a Reminder.

	example: /reminder add Don't forget to be awesome

help
	Display usage.
`
}

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "reminder",
		DisplayName:      "Post Reminder",
		Description:      "Interact with your Reminders.",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: add, help",
		AutoCompleteHint: "[command]",
		AutocompleteData: getAutocompleteData(),
	}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

// ExecuteCommand executes a given command and returns a command response.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	spaceRegExp := regexp.MustCompile(`\s+`)
	trimmedArgs := spaceRegExp.ReplaceAllString(strings.TrimSpace(args.Command), " ")
	stringArgs := strings.Split(trimmedArgs, " ")
	lengthOfArgs := len(stringArgs)
	restOfArgs := []string{}

	var handler func([]string, *model.CommandArgs) (bool, error)
	if lengthOfArgs == 1 {
		p.postCommandResponse(args, getHelp())
	} else {
		command := stringArgs[1]
		if lengthOfArgs > 2 {
			restOfArgs = stringArgs[2:]
		}
		switch command {
		case "add":
			p.postCommandResponse(args, getHelp())
		default:
			p.postCommandResponse(args, getHelp())
			return &model.CommandResponse{}, nil
		}
	}
	isUserError, err := handler(restOfArgs, args)
	if err != nil {
		if isUserError {
			p.postCommandResponse(args, fmt.Sprintf("__Error: %s.__\n\nRun `/reminder help` for usage instructions.", err.Error()))
		} else {
			p.API.LogError(err.Error())
			p.postCommandResponse(args, "An unknown error occurred. Please talk to your system administrator for help.")
		}
	}

	return &model.CommandResponse{}, nil
}

func getAutocompleteData() *model.AutocompleteData {
	reminder := model.NewAutocompleteData("reminder", "[command]", "Available commands: add, help")

	add := model.NewAutocompleteData("add", "[message]", "Adds a Reminder")
	add.AddTextArgument("E.g. be awesome", "[message]", "")
	reminder.AddCommand(add)

	help := model.NewAutocompleteData("help", "", "Display usage")
	reminder.AddCommand(help)
	return reminder
}
