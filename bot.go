package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

const (
	IMG_NOT_FOUND    string = "https://www.salonlfc.com/wp-content/uploads/2018/01/image-not-found-scaled-1150x647.png"
	ALT_TRAVEL_FLEX  string = "旅遊小幫手幫你推薦的景點"
	PROMPT_NOT_FOUND string = "你是一個想要去旅行社的導遊，你根據以下的對話來推薦台灣行程，如果對話內容跟旅遊無關，請給予旅遊相關的建議。 \n----\n"
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
				if isGroupEvent(event) {
					//in group must using :gpt
					if strings.Contains(message.Text, ":gpt") {
						handleGPT(GPT_FunctionCall, event, message.Text)
					}
				} else {
					//1:1 direct go gpt
					handleGPT(GPT_FunctionCall, event, message.Text)
				}
			}
		}
	}
}

func handleGPT(action GPT_ACTIONS, event *linebot.Event, message string) {
	switch action {
	case GPT_FunctionCall:
		message = strings.TrimPrefix(message, ":gpt")
		keyword, poiJsonRet := gptFuncCall(message)
		poi := handlePOIResponse([]byte(poiJsonRet))
		var gptMsg = ""
		var summary []byte
		var err error
		//找不到的時候，把原來問題帶回去問一次。
		if len(poi.Pois) == 0 {
			// gptMsg = gptCompleteContext(PROMPT_NOT_FOUND + message)
			// keyword, reply = gptFuncCall(gptMsg)
			// poi = handlePOIResponse([]byte(reply))
			gptMsg = "{}"
			if summary, err = OpenAIChatFuncCall(getSummaryString(PROMPT_NOT_FOUND+message, keyword, poiJsonRet)); err != nil {
				log.Println("OpenAIChatFuncCall getSummaryString fail:", err)
				return
			}
		} else {
			// 有答案
			if summary, err = OpenAIChatFuncCall(getSummaryString(message, keyword, poiJsonRet)); err != nil {
				log.Println("OpenAIChatFuncCall getSummaryString fail:", err)
				return
			}
		}

		log.Println("OpenAIChatFuncCall getSummaryString result:", string(summary))
		catResponse := handleFuncCallResponse(summary)
		log.Println("getSummaryString catResponse:", catResponse)
		sumMsg, _ := interfaceToString(catResponse.Choices[0].Message.Content)

		if gptMsg != "" {
			if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(sumMsg)).Do(); err != nil {
				log.Print(err)
			}
		} else {
			flexBuble := getPOIsFlexBubble(poi)
			log.Println("Prepre FlexMsg")

			flexContainerObj := &linebot.CarouselContainer{
				Type:     linebot.FlexContainerTypeCarousel,
				Contents: flexBuble,
			}
			flexMsg := linebot.NewFlexMessage(ALT_TRAVEL_FLEX, flexContainerObj)

			if _, err := bot.ReplyMessage(event.ReplyToken, flexMsg, linebot.NewTextMessage(sumMsg)).Do(); err != nil {
				log.Print(err)

				if out, err := json.Marshal(flexContainerObj); err != nil {
					log.Println("Marshal error:", err)
				} else {
					log.Println("---\nflex\n---\n", string(out))
				}
			}
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

func getPOIsFlexBubble(records ResponsePOI) []*linebot.BubbleContainer {
	log.Println("getPOIsFlexBubble")
	if len(records.Pois) == 0 {
		log.Println("err1")
		return nil
	}

	var columnList []*linebot.BubbleContainer
	for _, result := range records.Pois {
		log.Println("Add flex:", result.Name, result.CoverPhoto, result.PoiURL)
		name := linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   result.Name,
			Weight: linebot.FlexTextWeightTypeBold,
			Size:   linebot.FlexTextSizeTypeSm,
			Wrap:   true,
		}

		nickN := result.Name
		if len(result.Nickname) > 0 {
			nickN = result.Nickname[0]
		}
		nickName := linebot.TextComponent{
			Type:   linebot.FlexComponentTypeText,
			Text:   nickN,
			Weight: linebot.FlexTextWeightTypeBold,
			Size:   linebot.FlexTextSizeTypeSm,
			Wrap:   true,
		}

		btn := linebot.ButtonComponent{
			Type: linebot.FlexComponentTypeButton,
			Action: &linebot.URIAction{
				Label: "帶我去",
				URI:   result.PoiURL,
			},
		}

		var boxBody []linebot.FlexComponent
		boxBody = append(boxBody, &name, &nickName, &btn)

		// Title's hard limit by Line
		coverPhoto := IMG_NOT_FOUND
		if result.CoverPhoto != "" {
			coverPhoto = result.CoverPhoto
		}
		tmpColumn := linebot.BubbleContainer{
			Type: linebot.FlexContainerTypeBubble,
			Size: linebot.FlexBubbleSizeTypeMicro,
			Hero: &linebot.ImageComponent{
				Type:        linebot.FlexComponentTypeImage,
				URL:         coverPhoto,
				Size:        linebot.FlexImageSizeTypeFull,
				AspectRatio: linebot.FlexImageAspectRatioType1to1,
				AspectMode:  linebot.FlexImageAspectModeTypeCover,
			},
			Body: &linebot.BoxComponent{
				Type:     linebot.FlexComponentTypeBox,
				Layout:   linebot.FlexBoxLayoutTypeVertical,
				Contents: boxBody,
			},
		}

		columnList = append(columnList, &tmpColumn)
	}
	return columnList
}

func getPOIsCarouseTemplate(records ResponsePOI) (template *linebot.CarouselTemplate) {
	if len(records.Pois) == 0 {
		log.Println("err1")
		return nil
	}

	columnList := []*linebot.CarouselColumn{}
	for _, result := range records.Pois {
		// Title's hard limit by Line
		tmpColumn := linebot.NewCarouselColumn(
			result.CoverPhoto,
			result.Name,
			result.Nickname[0],
			linebot.NewURIAction("👉 點我打開", result.PoiURL),
		)
		columnList = append(columnList, tmpColumn)
	}
	template = linebot.NewCarouselTemplate(columnList...)
	return template
}
