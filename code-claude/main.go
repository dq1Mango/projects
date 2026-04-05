package main

import (
	"bytes"
	// "compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

var (
	// BaseURL = ""
	MessageURL = "https://api.anthropic.com/v1/messages?beta=true"
	TokenURL   = "https://platform.claude.com/v1/oauth/token"
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
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

func userContent(content string) RoleContent {
	return RoleContent{
		Role:    "user",
		Content: []Content{newContent(content)},
	}
}

type RefreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	ClientId     string `json:"client_id"`
	Scope        string `json:"scope"`
}

func newRefreshRequest(refresh_token string, client_id uuid.UUID) RefreshRequest {
	return RefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: refresh_token,
		ClientId:     client_id.String(),

		Scope: "user:profile user:inference user:sessions:claude_code user:mcp_servers user:file_upload",
	}
}

// the most useless payload ever
type outputConfig struct {
	Format struct {
		Type string `json:"type"`

		Schema struct {
			Type string `json:"type"`

			Properties struct {
				Title struct {
					Type string `json:"type"`
				} `json:"title"`
			} `json:"properties"`
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
				Type       string "json:\"type\""
				Properties struct {
					Title struct {
						Type string "json:\"type\""
					} "json:\"title\""
				} "json:\"properties\""
				Required             []string "json:\"required\""
				AdditionalProperties bool     "json:\"additionalProperties\""
			} "json:\"schema\""
		}{Type: "json_schema", Schema: struct {
			Type       string "json:\"type\""
			Properties struct {
				Title struct {
					Type string "json:\"type\""
				} "json:\"title\""
			} "json:\"properties\""
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

type EffortConfig struct {
	Effort string `json:"effort"`
}

func effortConfig(effort string) EffortConfig {
	return EffortConfig{
		Effort: effort,
	}
}

type MetaData struct {
	UserID string `json:"user_id"`
}

// if you are thinking to yourself:
// "hmmm, that looks an awful lot like a json object of strings inside a string inside a json object",
// you would be correct!
// someone at anthropic was feeling very lazy that day...
func metatdata() MetaData {
	return MetaData{
		UserID: fmt.Sprintf("{\"device_id\":\"%s\",\"account_uuid\":\"%s\",\"session_id\":\"%s\"}",
			DeviceID, AccountUUID, SESSION_ID),
	}

}

type AIRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`

	Messages []RoleContent `json:"messages"`
	System   []Content     `json:"system"`

	Metadata MetaData `json:"metadata"`

	Output_config any  `json:"output_config"`
	Stream        bool `json:"stream"`
}

func setReqHeaders(req *http.Request) {

	req.Header.Set("Authorization", "Bearer "+OAUTH_TOKEN)

	req.Header.Set("X-Claude-Code-Session-Id", SESSION_ID)

	uuid, _ := uuid.NewRandom()
	req.Header.Set("x-client-request-id", uuid.String())

	req.Header.Set("Accept", "application/json")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
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

// func testDecode() {
// 	resBody :=
//
// 	var decodedBody []byte
// 	switch enc := "gzip"; enc {
// 	case "gzip":
// 		reader, _ := gzip.NewReader(bytes.NewReader(resBody))
// 		defer reader.Close()
// 		n, err := reader.Read(decodedBody)
//
// 		fmt.Println(n)
// 		if err != nil {
// 			panic(err)
// 		}
//
// 	case "":
// 		decodedBody = resBody
//
// 	default:
// 		panic(fmt.Sprintf("unknown encoding type: %s", enc))
// 	}
//
// 	// resBody := res.Body
//
// 	fmt.Printf("client: response body: %s\n", decodedBody)
//
// }

func main() {

	body := AIRequest{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 1024,

		Messages: []RoleContent{
			userContent("hello, claude. What are your capabilities")},
		System: []Content{
			newContent(
				"x-anthropic-billing-header: cc_version=2.1.92.957; cc_entrypoint=cli; cch=e73d9;",
			),
		},

		// Output_config: defaultOutputConfig(),
		Metadata:      metatdata(),
		Output_config: effortConfig("medium"),
		Stream:        true,
	}

	b, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
	fmt.Println()
	// return

	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	// jsonBody, err = os.ReadFile("sillybody.json")
	// if err != nil {
	// 	panic(err)
	// }

	// requestURL := fmt.Sprintf("http://localhost:%d", serverPort)
	req, err := http.NewRequest(http.MethodPost, MessageURL, bytes.NewBuffer(jsonBody))
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
		os.Exit(1)
	}

	if res.Header.Get("Content-Encoding") == "gzip" {
		fmt.Println("response used gzip")
	} else {
		fmt.Println("ru hroh")
	}

	// var decodedBody []byte
	// switch enc := res.Header.Get("Content-Encoding"); enc {
	// case "gzip":
	// 	reader, _ := gzip.NewReader(bytes.NewReader(resBody))
	// 	defer reader.Close()
	// 	n, err := reader.Read(decodedBody)
	//
	// 	fmt.Println(n)
	// 	if err != nil {
	// 		fmt.Printf("client: could not read response body: %s\n", err)
	// 		panic(err)
	// 	}
	//
	// case "":
	// 	decodedBody = resBody
	//
	// default:
	// 	panic(fmt.Sprintf("unknown encoding type: %s", enc))
	// }

	// resBody := res.Body

	fmt.Printf("client: response body: %s\n", resBody)
}
