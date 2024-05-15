package main

import (
	sr "groqgo/src"
)

type GroqChatArgs struct {
	*sr.ChatArgs
}

func ChatGroq(kwargs ...map[string]interface{}) *sr.ChatArgs{
	var args sr.ChatArgs

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
	return &args
}

func (ca *GroqChatArgs) Call(prompt string) string{
	ca.Messages = append(ca.Messages, map[string]string{"role": "user", "content": prompt})
	ca.Stream = false
	resp, err := sr.Client(ca.ChatArgs)
	if err != nil{
		return err.Error()
	}
	return resp
}

func (ca *GroqChatArgs) Stream_(prompt string) string{
	ca.Messages = append(ca.Messages, map[string]string{"role": "user", "content": prompt})
	strm_resp,err:=sr.StreamClient(ca.ChatArgs)
	if err != nil{
		return err.Error()
	}
	return strm_resp
}

