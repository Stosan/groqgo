package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/joho/godotenv"
	"github.com/Stosan/groqgo/types"
)


func Client(qp *types.ChatArgs) (string, error) {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %w", err)
	}

	jsonPayload, err := json.Marshal(qp)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("GROQ_API_KEY")))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var chatErr types.ChatError
	if err := json.Unmarshal(responseData, &chatErr); err != nil {
		return "", fmt.Errorf("error unmarshaling chat error: %w", err)
	}


	var response types.ChatCompletionResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling chat completion response: %w", err)
	}

	content := response.Choices[0].Message.Content
	return content, nil
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
			return "", fmt.Errorf("error, %w", err)
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
