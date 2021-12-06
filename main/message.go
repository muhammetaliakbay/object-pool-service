package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ClaimMessage struct {
	Type    string   `json:"type"`
	Objects []string `json:"objects"`
}

type LoadMessage struct {
	Type string `json:"type"`
	Size uint32 `json:"size"`
}

type MarkMessage struct {
	Size uint32 `json:"size"`
}

type QueueMessage struct {
	Objects []string `json:"objects"`
	Group   string   `json:"group"`
}

type RequeueMessage struct {
	Objects []string `json:"objects"`
}

type ReleaseMessage struct {
	Objects []string `json:"objects"`
}

type Message struct {
	Type string `json:"type"`
}

func UnmarshalMessage(data []byte) (message interface{}, err error) {
	var pre = &Message{}
	if err = json.Unmarshal(data, pre); err != nil {
		return
	}
	switch pre.Type {
	case "mark":
		message = &MarkMessage{}
	case "queue":
		message = &QueueMessage{}
	case "requeue":
		message = &RequeueMessage{}
	case "release":
		message = &ReleaseMessage{}
	default:
		err = errors.New(fmt.Sprintf("Unexpected message type: %s", pre.Type))
		return
	}
	err = json.Unmarshal(data, message)
	return
}

func MarshallClaimMessage(objects []string) (data []byte) {
	data, _ = json.Marshal(&ClaimMessage{
		Type:    "claim",
		Objects: objects,
	})
	return
}

func MarshallLoadMessage(size uint32) (data []byte) {
	data, _ = json.Marshal(&LoadMessage{
		Type: "load",
		Size: size,
	})
	return
}
