package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	gpt3 "github.com/sashabaranov/go-openai"
)

type Arguments struct {
	Keyword string `json:"keyword"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role         string      `json:"role"`
			Content      interface{} `json:"content"`
			FunctionCall struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function_call"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// gptCompleteContextHandle the response from the function call.
func gptCompleteContext(ori string) (ret string) {
	// Get the context.
	ctx := context.Background()

	// For more details about the API of Open AI Chat Completion: https://platform.openai.com/docs/guides/chat
	req := gpt3.ChatCompletionRequest{
		// Model: The GPT4o turbo model is the most powerful model available.
		Model: gpt3.GPT3Dot5Turbo,
		// The message to complete.
		Messages: []gpt3.ChatCompletionMessage{{
			Role:    gpt3.ChatMessageRoleUser,
			Content: ori,
		}},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		ret = fmt.Sprintf("Err: %v", err)
	} else {
		// The response contains a list of choices, each with a score.
		// The score is a float between 0 and 1, with 1 being the most likely.
		// The choices are sorted by score, with the first choice being the most likely.
		// So we just take the first choice.
		ret = resp.Choices[0].Message.Content
	}

	return ret
}

func gptFuncCall(msg string) (keyword string, ret string) {
	var result []byte
	var err error
	log.Println("getQueryString:", getQueryString(msg))
	if result, err = OpenAIChatFuncCall(getQueryString(msg)); err != nil {
		log.Println("OpenAIChatFuncCall fail:", err)
		return "", ""
	}
	log.Println("OpenAIChatFuncCall result:", string(result))
	catResponse := handleFuncCallResponse(result)
	log.Println("catResponse:", catResponse)

	// Call 3rd party API
	if len(catResponse.Choices) == 0 {
		return "無 keyword", "資料有誤，請重新查詢"
	}
	log.Println("Arguments:", catResponse.Choices[0].Message.FunctionCall.Arguments)
	arg := handleArgument([]byte(catResponse.Choices[0].Message.FunctionCall.Arguments))
	poiResult, _ := SearchPOI(arg.Keyword)
	return arg.Keyword, string(poiResult)
}

func OpenAIChatFuncCall(requestBody map[string]interface{}) ([]byte, error) {
	url := "https://api.openai.com/v1/chat/completions"
	apiKey := apiKey

	jsonValue, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal JSON: %s", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("Failed to create HTTP request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("", apiKey)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %s", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %s", err)
	}

	return body, nil
}

func getQueryString(msg string) map[string]interface{} {
	return map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": msg,
			},
		},
		"functions": []map[string]interface{}{
			{
				"name":        "search_poi",
				"description": "Get the keyword about travel information",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"keyword"},
				},
			},
		},
	}
}

func getSummaryString(msg, arg, result string) map[string]interface{} {
	return map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": msg,
			},
			{
				"role":    "assistant",
				"content": nil,
				"function_call": map[string]interface{}{
					"name":      "search_poi",
					"arguments": arg,
				},
			},
			{
				"role":    "function",
				"name":    "search_poi",
				"content": result,
			},
		},
		"functions": []map[string]interface{}{
			{
				"name":        "search_poi",
				"description": "Get the keyword about travel information",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"keyword"},
				},
			},
		},
	}
}
