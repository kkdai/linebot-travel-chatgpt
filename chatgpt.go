package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	gpt3 "github.com/sashabaranov/go-openai"
)

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

type SearchPOIRequest struct {
	Keyword string `json:"keyword"`
}

type ResponsePOI struct {
	Pois []struct {
		PoiURL     string   `json:"poiURL"`
		CoverPhoto string   `json:"coverPhoto"`
		Name       string   `json:"name"`
		Nickname   []string `json:"nickname"`
	} `json:"pois"`
}

func SearchPOI(keyword string) string {
	url := "https://nextjs-chatgpt-plugin-starter.vercel.app/api/get-poi"

	data := &SearchPOIRequest{Keyword: keyword}
	reqBody, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error while marshalling data: %v", err)
		return ""
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error while making the request: %v", err)
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %s\n", err)
		return ""
	}

	return string(bodyBytes)
}

func gptCompleteContext(ori string) (ret string) {
	// Get the context.
	ctx := context.Background()

	// For more details about the API of Open AI Chat Completion: https://platform.openai.com/docs/guides/chat
	req := gpt3.ChatCompletionRequest{
		// Model: The GPT-3.5 turbo model is the most powerful model available.
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

func handleFuncCall(responseJSON []byte) ChatCompletionResponse {
	var response ChatCompletionResponse
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON: %s\n", err)
		return ChatCompletionResponse{}
	}
	return response
}

func OpenAIChatFuncCall(requestBody map[string]interface{}) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	apiKey := os.Getenv("ChatGptToken")

	jsonValue, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("Failed to marshal JSON: %s", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", fmt.Errorf("Failed to create HTTP request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(":", apiKey)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %s", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read response body: %s", err)
	}

	return string(body), nil
}

func getQueryString(msg string) map[string]interface{} {
	return map[string]interface{}{
		"model": "gpt-3.5-turbo-0613",
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
						"unit": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"keyword"},
				},
			},
		},
	}
}
