package groqgo

import (
	"github.com/Stosan/groqgo/internal"
	"github.com/Stosan/groqgo/types"
)

type GroqChatArgs struct {
	*types.ChatArgs
}

func ChatGroq(kwargs ...map[string]interface{}) *GroqChatArgs{
	var args types.ChatArgs

	for _, kwarg := range kwargs {
		if val, ok := kwarg["model"]; ok {
			args.Model = val.(string)
		}
		if val, ok := kwarg["messages"]; ok {
			args.Messages = val.([]map[string]string)
		}
		if val, ok := kwarg["temperature"]; ok {
			args.Temperature = val.(float64)
		}
		if val, ok := kwarg["top_p"]; ok {
			args.TopP= val.(float64)
		}
		if val, ok := kwarg["stream"]; ok {
			args.Stream = val.(bool)
		}
		if val, ok := kwarg["stop"]; ok {
			if stopVal, ok := val.([]string); ok {
				args.Stop = stopVal
			} else if val == nil {
				args.Stop = nil
			}
		}

		// ... other fields ...
	}
	return &GroqChatArgs{&args}
}




// ChatClient sends a prompt to the chat client and returns the response.
func (args *GroqChatArgs) ChatClient(prompt string) (string, ChatError) {
	if args.ChatArgs.Messages == nil {
		args.ChatArgs.Messages = make([]map[string]string, 0)
	}
	args.ChatArgs.Messages = append(args.ChatArgs.Messages, map[string]string{"role": "user", "content": prompt})
	args.Stream = false
	response, err := internal.Client(args.ChatArgs)
	if err != nil {
		return "", ChatError{err}
	}
	return response, ChatError{err}
}

// ChatError represents a chat-related error.
type ChatError struct {
	error
}


func (params *GroqChatArgs) StreamClient(prompt string) (string,error){
	params.ChatArgs.Messages = append(params.ChatArgs.Messages, map[string]string{"role": "user", "content": prompt})
	response,err:=internal.StreamClient(params.ChatArgs)
	if err != nil{
		return "",err
	}
	return response,nil
}
