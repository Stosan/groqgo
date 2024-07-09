package groqgo


import (
	"github.com/Stosan/groqgo/internal"
	"github.com/Stosan/groqgo/types"
)

// ChatError represents a chat-related error.
type ClientChatError struct {
	error
}

type GroqChatArgs struct {
	types.ChatArgs
}

func ChatGroq(kwargs ...map[string]interface{}) GroqChatArgs{
	var args types.ChatArgs

	for _, kwarg := range kwargs {
		if val, ok := kwarg["model"]; ok {
			args.Model = val.(string)
		}
		if val, ok := kwarg["messages"]; ok {
			args.Messages = val.([]types.Message)
		}
		if val, ok := kwarg["temperature"]; ok {
			args.Temperature = val.(float64)
		}
		if val, ok := kwarg["top_p"]; ok {
			args.TopP= val.(float64)
		}
		if val, ok := kwarg["seed"]; ok {
			args.Seed= val.(int)
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
	return GroqChatArgs{args}
}




// ChatClient sends a prompt to the chat client and returns the response.
func (args GroqChatArgs) Chat(prompt string, system string) (string,  error) {
	if args.ChatArgs.Messages == nil {
		args.ChatArgs.Messages = make([]types.Message, 0)
	}
	if system == ""{
		args.ChatArgs.Messages = append(args.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	}else{
		args.ChatArgs.Messages = append(args.ChatArgs.Messages,types.Message{Role: "user", Content: prompt},types.Message{Role: "system", Content: system})
	}

	args.Stream = false
	response, err := internal.Client(args.ChatArgs)
	if err != nil {
		return "",  err
	}
	return response,  err
}



func (params GroqChatArgs) StreamCompleteChat(prompt string, system string) (string,  error) {
	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	if system == ""{
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	}else{
		params.ChatArgs.Messages = append(params.ChatArgs.Messages,types.Message{Role: "user", Content: prompt},types.Message{Role: "system", Content: system})
	}

	params.Stream = true

	response,err:= internal.StreamCompleteClient(params.ChatArgs)

	if err != nil {
		return "",  err
	}
	return response,  err
}


func (params GroqChatArgs) StreamChat(prompt string, system string)  <-chan string {
	if params.ChatArgs.Messages == nil {
		params.ChatArgs.Messages = make([]types.Message, 0)
	}

	if system == ""{
		params.ChatArgs.Messages = append(params.ChatArgs.Messages, types.Message{Role: "user", Content: prompt})
	}else{
		params.ChatArgs.Messages = append(params.ChatArgs.Messages,types.Message{Role: "user", Content: prompt},types.Message{Role: "system", Content: system})
	}

	params.Stream = true
	chunkchan := make(chan string)

    go func() {
		defer close(chunkchan)
        err := internal.StreamClient(params.ChatArgs, chunkchan)
        if err != nil {
           chunkchan <- err.Error()
        }
    }()

    return chunkchan
}
