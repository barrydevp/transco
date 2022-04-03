package transco

import (
	"time"
)

type ParticipantAction struct {
	Data         interface{}               `json:"data"`
	Uri          *string                   `json:"uri"`
	Status       string                    `json:"status"`
	Results      []*map[string]interface{} `json:"results"`
	InvokedCount int                       `json:"invokedCount"`
}

type Participant struct {
	session *Session `json:"-"`

	Id        int64  `json:"id"`
	SessionId string `json:"sessionId"`

	ClientId         string             `json:"clientId"`
	RequestId        string             `json:"requestId"`
	State            string             `json:"state"`
	CompensateAction *ParticipantAction `json:"compensateAction,omitempty"`
	CompleteAction   *ParticipantAction `json:"completeAction,omitempty"`
	UpdatedAt        *time.Time         `json:"updatedAt,omitempty"`
	CreatedAt        *time.Time         `json:"createdAt"`
}

type Session struct {
	client *Client `json:"-"`

	Id string `json:"id"`

	State           string     `json:"state"`
	Timeout         int        `json:"timeout"`
	UpdatedAt       *time.Time `json:"updatedAt,omitempty"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	CreatedAt       *time.Time `json:"createdAt,omitempty"`
	Errors          []string   `json:"errors,omitempty"`
	Retries         int        `json:"retries"`
	TerminateReason string     `json:"terminateReason,omitempty"`
}

type ParticipantJoinBody struct {
	ClientId  string `json:"clientId"`
	RequestId string `json:"requestId"`
}

type ParticipantCommit struct {
	Id         int64              `json:"participantId"`
	Compensate *ParticipantAction `json:"compensate"`
	Complete   *ParticipantAction `json:"complete"`
}
