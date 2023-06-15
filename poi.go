package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const (
	POIAPIUrl = "https://nextjs-chatgpt-plugin-starter.vercel.app/api/get-poi"
)

var (
	httpClient = &http.Client{}
	ErrMarshal = errors.New("error while marshalling data")
	ErrRequest = errors.New("error while making the request")
	ErrRead    = errors.New("failed to read response body")
)

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

func SearchPOI(keyword string) ([]byte, error) {
	data := &SearchPOIRequest{Keyword: keyword}
	reqBody, err := json.Marshal(data)
	if err != nil {
		return nil, ErrMarshal
	}

	resp, err := httpClient.Post(POIAPIUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, ErrRequest
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
