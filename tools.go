package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

func handlePOIResponse(responseJSON []byte) ResponsePOI {
	var response ResponsePOI
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON: %s\n", err)
		return ResponsePOI{}
	}
	return response
}

func handleArgument(responseJSON []byte) Arguments {
	var response Arguments
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON: %s\n", err)
		return Arguments{}
	}
	return response
}

func handleFuncCallResponse(responseJSON []byte) ChatCompletionResponse {
	var response ChatCompletionResponse
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		fmt.Printf("Failed to unmarshal JSON: %s\n", err)
		return ChatCompletionResponse{}
	}
	return response
}

func interfaceToString(val interface{}) (string, error) {
	str, ok := val.(string)
	if !ok {
		return "", errors.New("the interface value is not a string")
	}
	return str, nil
}
