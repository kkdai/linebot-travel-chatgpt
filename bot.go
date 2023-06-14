package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			// Handle only on text message
			case *linebot.TextMessage:
				// Directly to ChatGPT
				if strings.Contains(message.Text, ":gpt") {
					handleGPT(GPT_FunctionCall, event, message.Text)
				}
			}
		}
	}
}

func handleGPT(action GPT_ACTIONS, event *linebot.Event, message string) {
	switch action {
	case GPT_FunctionCall:
		keyword, reply := gptFuncCall(message)

		if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("關鍵字："+keyword), linebot.NewTextMessage(reply)).Do(); err != nil {
			log.Print(err)
		}
	}
}

func isGroupEvent(event *linebot.Event) bool {
	return event.Source.GroupID != "" || event.Source.RoomID != ""
}

func getGroupID(event *linebot.Event) string {
	if event.Source.GroupID != "" {
		return event.Source.GroupID
	} else if event.Source.RoomID != "" {
		return event.Source.RoomID
	}

	return ""
}

func sendCarouselMessage(event *linebot.Event, template *linebot.CarouselTemplate, altText string) {
	if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTemplateMessage(altText, template)).Do(); err != nil {
		log.Println(err)
	}
}
