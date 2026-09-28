package jev

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const EVAL_ENDPOINT = "https://api.typesafe.ai/v1/systemone"

const DEFAULT_MODEL = "jev-latest"

type QuestionType string

const (
	Noul   QuestionType = "noul"
	Choice QuestionType = "choice"
	Score  QuestionType = "score"
)

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type NoulAnswer struct {
	Noul float32 `json:"noul"`
}

type ChoiceAnswer struct {
	Choice       string             `json:"choice"`
	Probabilites map[string]float32 `json:"probabilities"`
	Confidence   float32            `json:"confidence"`
}

type ScoreAnswer struct {
	Score        float32            `json:"score"`
	Legend       map[string]string  `json:"legend"`
	Probabilites map[string]float32 `json:"probabilities"`
	Confidence   float32            `json:"confidence"`
}

type RawAnswer struct {
	Type QuestionType `json:"type"`

	// noul properties
	Noul float32 `json:"noul"`

	// choice properties
	Choice       string             `json:"choice"`
	Probabilites map[string]float32 `json:"probabilities"`
	Confidence   float32            `json:"confidence"`

	// score properties
	Score  float32           `json:"score"`
	Legend map[string]string `json:"legend"`
	// Probabilites map[string]float32 `json:"probabilities"`
	// Confidence float32 `json:"confidence"`
}

type Response struct {
	Model   string               `json:"model"`
	Usage   Usage                `json:"usage"`
	Answers map[string]RawAnswer `json:"answers"`
}

type Jev struct {
	token string
}

func NewJev(token string) *Jev {
	return &Jev{token}
}

func (j *Jev) post(r io.Reader) (*http.Response, error) {
	var headers http.Header = map[string][]string{}

	headers.Add("Authorization", fmt.Sprintf("Bearer %s", j.token))
	headers.Add("Content-Type", "application/json")

	url, err := url.Parse(EVAL_ENDPOINT)
	if err != nil {
		panic(err)
	}

	rc, ok := r.(io.ReadCloser)
	if !ok {
		rc = io.NopCloser(r)
	}

	req := http.Request{
		Method: http.MethodPost,
		URL:    url,
		Header: headers,
		Body:   rc,
	}

	return http.DefaultClient.Do(&req)
}

type UsefullResponse struct {
	Usage Usage

	Nouls   map[string]NoulAnswer
	Choices map[string]ChoiceAnswer
	Scores  map[string]ScoreAnswer
}

func parseResponse(response Response) UsefullResponse {
	usefull := UsefullResponse{
		Usage:   response.Usage,
		Nouls:   map[string]NoulAnswer{},
		Choices: map[string]ChoiceAnswer{},
		Scores:  map[string]ScoreAnswer{},
	}

	for key, answer := range response.Answers {
		switch answer.Type {
		case Noul:
			noul := NoulAnswer{Noul: answer.Noul}

			usefull.Nouls[key] = noul
		case Choice:
			choice := ChoiceAnswer{
				Choice:       answer.Choice,
				Probabilites: answer.Probabilites,
				Confidence:   answer.Confidence,
			}

			usefull.Choices[key] = choice
		case Score:
			score := ScoreAnswer{
				Score:        answer.Score,
				Probabilites: answer.Probabilites,
				Confidence:   answer.Confidence,
				Legend:       answer.Legend,
			}

			usefull.Scores[key] = score
		}
	}

	return usefull
}

func (j *Jev) SystemOne(request []byte) (*UsefullResponse, error) {

	resp, err := j.post(bytes.NewReader(request))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
		fmt.Println(string(body))
		return nil, errors.New("bad status: " + resp.Status)
	}

	var response Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	usefull := parseResponse(response)

	return &usefull, nil

}
