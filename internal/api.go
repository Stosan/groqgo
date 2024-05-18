package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Stosan/groqgo/types"
	"github.com/joho/godotenv"
)

var mu sync.Mutex

func Client(qp *types.ChatArgs) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}

	jsonPayload, err := json.Marshal(qp)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))
	// Add custom headers if needed
	// req.Header.Set("Custom-Header", "value")

	client := &http.Client{}
	responseChan := make(chan *http.Response)
	errorChan := make(chan error)

	go func() {
		resp, err := retryRequest(client, req)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- resp
	}()

	var resp *http.Response
	select {
	case resp = <-responseChan:
	case err = <-errorChan:
		return "", fmt.Errorf("error making HTTP request: %w", err)
	case <-ctx.Done():
		return "", fmt.Errorf("request timed out")
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	b,_:= io.ReadAll(resp.Body)
	var clientErr types.ErrorResponse
	if err := decoder.Decode(&clientErr); err != nil {
		return "", fmt.Errorf("error unmarshaling chat error: %w", err)
	} else if bytes.HasPrefix(b, []byte(`{"error"`)) {
		return "", fmt.Errorf("API error: %v", clientErr)
	}

	var response types.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error unmarshaling chat completion response: %w", err)
	}

	content := response.Choices[0].Message.Content
	return content, nil
}

func retryRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 5; i++ { // Retry up to 5 times
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
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


func StreamClient(qp *types.ChatArgs) (string, error) {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	jsonPayload, err := json.Marshal(qp)
	if err != nil {
		println(err.Error())
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		println(err.Error())
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		println(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	body := resp.Body

	scanner := bufio.NewScanner(body)
	var errorPrefix = []byte(`data: {"error":`)
	var hasErrorPrefix bool
	
	for scanner.Scan() {
		var result string
		err := scanner.Err()
		if err != nil || hasErrorPrefix {
			panic(err)
		}

		if bytes.HasPrefix(scanner.Bytes(), errorPrefix) {
			hasErrorPrefix = true
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "data:") {
			noPrefixLine := strings.TrimPrefix(line, "data: ")
			if string(noPrefixLine) == "[DONE]" {
				return "", io.EOF
			}

			noPrefixLineBytes := []byte(noPrefixLine)
			var chunk types.ChatCompletionChunkResponse
			err = json.Unmarshal(noPrefixLineBytes, &chunk)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				continue
			}

			
			result += chunk.Choices[0].Delta.Content
		}
		return result, nil
	}

	return "", nil
}
