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


func (params *GroqChatArgs) ChatClient(prompt string)(string,error){
	params.ChatArgs.Messages = append(params.ChatArgs.Messages, map[string]string{"role": "user", "content": prompt})
	params.Stream = false
	response, err := internal.Client(params.ChatArgs)
	if err != nil{
		return "",err
	}
	return response,nil
}



func (params *GroqChatArgs) StreamClient(prompt string) (string,error){
	params.ChatArgs.Messages = append(params.ChatArgs.Messages, map[string]string{"role": "user", "content": prompt})
	response,err:=internal.StreamClient(params.ChatArgs)
	if err != nil{
		return "",err
	}
	return response,nil
}
