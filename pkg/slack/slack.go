package slack

import (
	"encoding/json"
	"net/http"
	"fmt"

	"go.uber.org/zap"
	"github.com/annaworks/surubot/pkg/api"
	Conf "github.com/annaworks/surubot/pkg/conf"
	"github.com/slack-go/slack"
)

type SlashService struct {
	Logger *zap.Logger
	Conf Conf.Conf
	Api *slack.Client
}

func NewSlashService(logger *zap.Logger, conf Conf.Conf) *SlashService {
	return &SlashService{
		Logger: logger,
		Conf: conf,
		Api: slack.New(conf.SLACK_TOKEN),
	}
}

func (s SlashService) GetSlashRoute() *api.Route{
	return &api.Route{
		Path: "/api/v1/slash",
		Handler: s.HandleSlashCommand,
		Method: http.MethodPost,
		Name: "slash",
	}
}

func (s SlashService) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	s.Logger.Info("Received a slash command")

	slash, err := slack.SlashCommandParse(r)
	if err != nil {
		s.Logger.Error("Error in parsing slash command", zap.Error(err))
		return
	}

	fmt.Printf("Command called: %v", slash.Command)

	switch slash.Command {
	case "/suru":
		m := NewQuestionMessage(slash.Text, slash.UserName)
		msg := m.BuildMessage()

		b, err := json.MarshalIndent(msg, "", "    ")
		if err != nil {
			fmt.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)

		s.Logger.Info("Handled slash command with message response")
		return

	default:
		s.Logger.Error("Command not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const (
	contextText    = "Asked by @%v"
	viewButtonText = "View"
	viewButtonId   = "viewButton"

	answerButtonText = "Answer"
	answerButtonId   = "answerButton"
)

type questionMessage struct {
	text 			string
	username 	string
}

func NewQuestionMessage(text, username string) *questionMessage {
	return &questionMessage{
		text: text,
		username: username,
	}
}

func (q questionMessage) BuildMessage() slack.Message {
	// Fields
	textField := slack.NewTextBlockObject("mrkdwn", q.text, false, false)
	fieldSlice := make([]*slack.TextBlockObject, 0)
	fieldSlice = append(fieldSlice, textField)
	fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

	// Context
	contextDisplay := fmt.Sprintf(contextText, q.username)
	contextField := slack.NewTextBlockObject(slack.MarkdownType, contextDisplay, false, false)
	contextBlock := slack.NewContextBlock("context", contextField)

	// Approve and Deny Buttons
	viewBtnTxt := slack.NewTextBlockObject("plain_text", viewButtonText, false, false)
	viewBtn := slack.NewButtonBlockElement("", viewButtonId, viewBtnTxt)

	answerBtnTxt := slack.NewTextBlockObject("plain_text", answerButtonText, false, false)
	answerBtn := slack.NewButtonBlockElement("", answerButtonId, answerBtnTxt)

	actionBlock := slack.NewActionBlock("", viewBtn, answerBtn)

	msg := slack.NewBlockMessage(
		fieldsSection,
		contextBlock,
		actionBlock,
	)
	msg.ResponseType = slack.ResponseTypeInChannel

	return msg
}
