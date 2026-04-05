package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func newContent(content string) Content {
	return Content{
		Type: "text",
		Text: content,
	}
}

type RoleContent struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

func userContent(content string) RoleContent {
	return RoleContent{
		Role:    "user",
		Content: newContent(content),
	}
}

// the most useless payload ever
type outputConfig struct {
	Format struct {
		Type string `json:"type"`

		Schema struct {
			Type string `json:"object"`

			Properties struct {
				Title struct {
					Type string `json:"type"`
				} `json:"title"`
			} `json:"schema"`
			Required             []string `json:"required"`
			AdditionalProperties bool     `json:"additionalProperties"`
		} `json:"schema"`
	} `json:"format"`
}

// i think its funny how impossibly legible this is
func defaultOutputConfig() outputConfig {
	return outputConfig{
		Format: struct {
			Type   string "json:\"type\""
			Schema struct {
				Type       string "json:\"object\""
				Properties struct {
					Title struct {
						Type string "json:\"type\""
					} "json:\"title\""
				} "json:\"schema\""
				Required             []string "json:\"required\""
				AdditionalProperties bool     "json:\"additionalProperties\""
			} "json:\"schema\""
		}{Type: "json_schema", Schema: struct {
			Type       string "json:\"object\""
			Properties struct {
				Title struct {
					Type string "json:\"type\""
				} "json:\"title\""
			} "json:\"schema\""
			Required             []string "json:\"required\""
			AdditionalProperties bool     "json:\"additionalProperties\""
		}{Type: "object", Properties: struct {
			Title struct {
				Type string "json:\"type\""
			} "json:\"title\""
		}{Title: struct {
			Type string "json:\"type\""
		}{Type: "string"}}, Required: []string{"title"}, AdditionalProperties: false}}}

}

type AIRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`

	Messages []RoleContent

	Output_config outputConfig `json:"output_config"`
	Stream        string       `json:"stream"`
}

func setReqHeaders(req *http.Request) {

	req.Header.Set("Authorization", "Bearer "+OAUTH_TOKEN)

	req.Header.Set("X-Claude-Code-Session-Id", SESSION_ID)

	uuid, _ := uuid.NewRandom()
	fmt.Println(uuid)
	req.Header.Set("x-client-request-id", uuid.String())

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set(
		"anthropic-beta",
		"claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,redact-thinking-2026-02-12,context-management-2025-06-27,prompt-caching-scope-2026-01-05,advanced-tool-use-2025-11-20,effort-2025-11-24",
	)
	req.Header.Set("anthropic-dangerous-direct-browser-access", "true")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", "api.anthropic.com")
	req.Header.Set("User-Agent", "claude-cli/2.1.91 (external, cli)")
	req.Header.Set("x-app", "cli")
	req.Header.Set("X-Stainless-Arch", "x64")
	req.Header.Set("X-Stainless-Lang", "js")
	req.Header.Set("X-Stainless-OS", "Linux")
	req.Header.Set("X-Stainless-Package-Version", "0.80.0")
	req.Header.Set("X-Stainless-Retry-Count", "0")
	req.Header.Set("X-Stainless-Runtime", "node")
	req.Header.Set("X-Stainless-Runtime-Version", "v24.3.0")
	req.Header.Set("X-Stainless-Timeout", "600")
}

func testDecode() {
	var decodedBody []byte
	switch enc := "gzip"; enc {
	case "gzip":
		reader, _ := gzip.NewReader(bytes.NewReader(resBody))
		defer reader.Close()
		n, err := reader.Read(decodedBody)

		fmt.Println(n)
		if err != nil {
			panic(err)
		}

	case "":
		decodedBody = resBody

	default:
		panic(fmt.Sprintf("unknown encoding type: %s", enc))
	}

	// resBody := res.Body

	fmt.Printf("client: response body: %s\n", decodedBody)

}

func main() {

	body := AIRequest{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 1024,
		Messages: []RoleContent{
			userContent("hello, claude. What are your capabilities")},
		Output_config: defaultOutputConfig(),
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	// requestURL := fmt.Sprintf("http://localhost:%d", serverPort)
	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	setReqHeaders(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}

	var decodedBody []byte
	switch enc := res.Header.Get("Content-Encoding"); enc {
	case "gzip":
		reader, _ := gzip.NewReader(bytes.NewReader(resBody))
		defer reader.Close()
		n, err := reader.Read(decodedBody)

		fmt.Println(n)
		if err != nil {
			panic(err)
		}

	case "":
		decodedBody = resBody

	default:
		panic(fmt.Sprintf("unknown encoding type: %s", enc))
	}

	// resBody := res.Body

	fmt.Printf("client: response body: %s\n", decodedBody)
}
