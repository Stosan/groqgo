package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/Stosan/groqgo/types"
	"github.com/joho/godotenv"
)

// Configure the HTTP transport for connection reuse
var transport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     90 * time.Second,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 10 * time.Second,
}

// Global HTTP client to reuse across requests
var httpClient = &http.Client{
	Transport: transport,
	Timeout:   0, // No timeout for streaming; use context for control
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


func Client(req types.ChatArgs) (string, error) {
	// Marshal the payload to JSON
	reqJsonPayload, err := json.Marshal(req)

	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer([]byte(reqJsonPayload)))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))

	// Make the request with retry logic
	resp, err := retryRequest(httpClient, request)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		// Read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading response body: %w", err)
		}

		// Convert the response body to a string
		bodyString := string(bodyBytes)

		// Print the response body
		fmt.Println(bodyString)
	}

	// Check if the response status indicates an error
	if resp.StatusCode >= 400 {
		var clientErr *types.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&clientErr); err != nil {
			return "", fmt.Errorf("error unmarshaling error response: %w", err)
		}
		return "", fmt.Errorf("API error: %v", clientErr)
	}

	// Unmarshal the successful response
	var response types.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error unmarshaling chat completion response: %w", err)
	}
	
	// Extract the content from the response
	content := response.Choices[0].Message.Content

	return content, nil
}




func StreamCompleteClient(req types.ChatArgs) (string, error) {

	// Marshal the payload to JSON
	reqJsonPayload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(reqJsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))

	// Make the request
	resp, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Use a scanner to read the streaming response
	scanner := bufio.NewScanner(resp.Body)
	result := strings.Builder{}

	for scanner.Scan() {
		line := scanner.Text()
		// Check for data prefix
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


func StreamClient(req types.ChatArgs, chunkchan chan string)  error{

	// Marshal the payload to JSON
	reqJsonPayload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create a new HTTP request
	request, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(reqJsonPayload))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set request headers
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))

	// Make the request
	resp, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return  fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Use a scanner to read the streaming response
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		// Check for data prefix
		if strings.HasPrefix(line, "data:") {
			noPrefixLine := strings.TrimPrefix(line, "data: ")
			if noPrefixLine == "[DONE]" {
				break
			}

			var chunk types.ChatCompletionChunkResponse
			if err := json.Unmarshal([]byte(noPrefixLine), &chunk); err != nil {
				return  fmt.Errorf("error unmarshaling chunk response: %w", err)
			}

			chunkchan <- chunk.Choices[0].Delta.Content
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Close the channel after sending all words
	defer close(chunkchan)
	return nil
}
