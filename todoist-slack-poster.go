package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/sachaos/todoist/lib"

	"github.com/nlopes/slack"
)

func main() {
	setting := getKeys()
	api := slack.New(setting.SlackToken)
	s, _ := todoist.SyncAll(setting.TodoistToken)
	var text string
	text = "本日のタスクです"
	var attchments []slack.Attachment
	for _, item := range getTodayTask(s.Items) {
		var color string
		switch item.Priority {
		case 1:
			color = "#ff0000"
		case 2:
			color = "#ff8100"
		case 3:
			color = "#ffc800"
		}
		var simekiri string
		if item.DueDateTime().Unix() <= time.Now().Unix() {
			simekiri = "*" + item.DueDateTime().Format("2006/01/02") + " *"
		} else {
			simekiri = item.DueDateTime().Format("2006/01/02")
		}
		at := slack.Attachment{
			Color: color,
			Title: item.GetContent(),
			Fields: []slack.AttachmentField{
				{
					Title: "Project",
					Value: item.GetProjectName(s.Projects),
					Short: true,
				},
				{
					Title: "Due Date",
					Value: simekiri,
					Short: true,
				},
			},
			MarkdownIn: []string{"fields"},
		}
		attchments = append(attchments, at)
	}
	api.PostMessage("#tasks", text, slack.PostMessageParameters{
		Username:    "TodoistTaskToaster",
		Markdown:    true,
		Attachments: attchments,
		IconURL:     setting.IconURL,
	})
}

func getTodayTask(items todoist.Items) todoist.Items {
	var result todoist.Items
	for _, item := range items {
		if item.DueDateTime().Unix() > time.Date(2010, 1, 1, 1, 1, 1, 11, time.UTC).Unix() && item.DueDateTime().Unix() <= time.Now().Unix() {
			result = append(result, item)
		}
	}
	return result
}

func getKeys() settingJSON {
	raw, err := ioutil.ReadFile(path.Dir(os.Args[0]) + "/setting.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var x settingJSON
	json.Unmarshal(raw, &x)
	return x
}

type settingJSON struct {
	SlackToken   string `json:"slack_token"`
	TodoistToken string `json:"todoist_token"`
	IconURL      string `json:"icon_url"`
}
