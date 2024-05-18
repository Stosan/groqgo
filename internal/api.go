package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Stosan/groqgo/types"
	"github.com/joho/godotenv"
)

// // Configure the HTTP transport for connection reuse
// var transport = &http.Transport{
// 	MaxIdleConns:        100,
// 	MaxIdleConnsPerHost: 100,
// 	IdleConnTimeout:     90 * time.Second,
// 	DialContext: (&net.Dialer{
// 		Timeout:   30 * time.Second,
// 		KeepAlive: 30 * time.Second,
// 	}).DialContext,
// 	TLSHandshakeTimeout: 10 * time.Second,
// }

// // Global HTTP client to reuse across requests
// var httpClient = &http.Client{
// 	Transport: transport,
// 	Timeout:   0, // No timeout for streaming; use context for control
// }


// Global HTTP client to reuse across requests
var httpClient = &http.Client{
	Timeout: 10 * time.Second, // Set a global timeout
}

func init() {
	// Load environment variables once during initialization
	if err := godotenv.Load(".env"); err != nil {
		panic(fmt.Sprintf("error loading .env file: %v", err))
	}
}


func retryRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 5; i++ { // Retry up to 5 times
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		timeout := calculateRetryTimeout(i)
		time.Sleep(timeout)
	}
	return resp, err
}

func calculateRetryTimeout(retryCount int) time.Duration {
	// Exponential backoff with jitter
	sleepSeconds := math.Min(float64(int(1<<retryCount)), 60) // Cap at 60 seconds
	jitter := sleepSeconds * (1 + 0.25*(rand.Float64()-0.5))
	return time.Duration(jitter) * time.Second
}


func Client(qp *types.ChatArgs) (string, error) {
	// Marshal the payload to JSON
	jsonPayload, err := json.Marshal(qp)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))

	// Make the request with retry logic
	resp, err := retryRequest(httpClient, req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Check if the response status indicates an error
	if resp.StatusCode >= 400 {
		var clientErr types.ErrorResponse
		if err := json.Unmarshal(body, &clientErr); err != nil {
			return "", fmt.Errorf("error unmarshaling error response: %w", err)
		}
		return "", fmt.Errorf("API error: %v", clientErr)
	}

	// Unmarshal the successful response
	var response types.ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling chat completion response: %w", err)
	}

	// Extract the content from the response
	content := response.Choices[0].Message.Content
	return content, nil
}





func StreamClient(qp *types.ChatArgs) (string, error) {
	// Marshal the payload to JSON
	jsonPayload, err := json.Marshal(qp)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))

	// Make the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Use a scanner to read the streaming response
	scanner := bufio.NewScanner(resp.Body)
	var result strings.Builder
	var errorPrefix = []byte(`data: {"error":`)
	var hasErrorPrefix bool

	for scanner.Scan() {
		line := scanner.Text()

		if bytes.HasPrefix(scanner.Bytes(), errorPrefix) {
			hasErrorPrefix = true
		}

		if hasErrorPrefix {
			var clientErr types.ErrorResponse
			if err := json.Unmarshal(scanner.Bytes(), &clientErr); err != nil {
				return "", fmt.Errorf("error unmarshaling error response: %w", err)
			}
			return "", fmt.Errorf("API error: %v", clientErr)
		}

		if strings.HasPrefix(line, "data:") {
			noPrefixLine := strings.TrimPrefix(line, "data: ")
			if noPrefixLine == "[DONE]" {
				break
			}

			var chunk types.ChatCompletionChunkResponse
			if err := json.Unmarshal([]byte(noPrefixLine), &chunk); err != nil {
				return "", fmt.Errorf("error unmarshaling chunk response: %w", err)
			}

			result.WriteString(chunk.Choices[0].Delta.Content)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return result.String(), nil
}

