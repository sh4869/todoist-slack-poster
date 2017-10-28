package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/nlopes/slack"
	"github.com/sachaos/todoist/lib"
)

func main() {
	setting := getKeys()
	api := slack.New(setting.SlackToken)
	config := todoist.Config{AccessToken: setting.TodoistToken, DebugMode: false}
	c := todoist.NewClient(&config)
	c.Sync(context.Background())
	s := c.Store
	var attchments []slack.Attachment
	var items todoist.Items
	deadtask := getDeadTask(s.Items)
	todaytask := getTodayTask(s.Items)
	tommorowtask := getTommorwTask(s.Items)
	items = append(items, deadtask...)
	items = append(items, todaytask...)
	items = append(items, tommorowtask...)
	text := "<@sh4869>\n【締め切りを過ぎたタスク】" + strconv.Itoa(len(deadtask)) + "件\n【今日のタスク】" + strconv.Itoa(len(todaytask)) + "件\n【明日のタスク】" + strconv.Itoa(len(tommorowtask)) + "件"
	for _, item := range items {
		var color string
		switch item.Priority {
		case 4:
			color = "#ff0000"
		case 3:
			color = "#ff8100"
		case 2:
			color = "#ffc800"
		}
		var simekiri string
		if item.DueDateTime().Unix() <= time.Now().Unix() {
			simekiri = item.DueDateTime().Format("2006/01/02") + " *(〆切り過ぎてる!!!)* "
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
		IconEmoji:   ":kirika:",
	})
}

func getTommorwTask(items todoist.Items) todoist.Items {
	var result todoist.Items
	for _, item := range items {
		if item.DueDateTime().Format("2006-01-02") == time.Now().AddDate(0, 0, 1).Format("2006-01-02") {
			result = append(result, item)
		}
	}
	return result
}

func getTodayTask(items todoist.Items) todoist.Items {
	var result todoist.Items
	for _, item := range items {
		if item.DueDateTime().Format("2006-01-02") == time.Now().Format("2006-01-02") {
			result = append(result, item)
		}
	}
	return result
}

func getDeadTask(items todoist.Items) todoist.Items {
	var result todoist.Items
	today := time.Now()
	for _, item := range items {
		if item.DueDateTime().Unix() > time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local).Unix() && item.DueDateTime().Unix() < time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local).Unix() {
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
