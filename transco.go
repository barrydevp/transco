package transco

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	BASE_TRANSCOORDITOR_URL = "http://transcoorditor:8080"
	V1_PREFIX_PATH          = "api/v1"
)

type Client struct {
	rest *resty.Client
}

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

	State     string     `json:"state"`
	Timeout   int        `json:"timeout"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Errors    []string   `json:"errors,omitempty"`
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

func New(url string) *Client {
	if url == "" {
		url = BASE_TRANSCOORDITOR_URL
	}

	rest := resty.New()
	rest.SetBaseURL(url)

	return &Client{
		rest: rest,
	}
}

type ErrorResponse struct {
	Err string `json:"err"`
	Msg string `json:"msg"`
}

type SessionResponse struct {
	Data *Session `json:"data"`
}

type ParticipantResponse struct {
	Data *Participant `json:"data"`
}

func (c *Client) v1SessionPath() string {
	return V1_PREFIX_PATH + "/sessions"
}

func (c *Client) newRequest() *resty.Request {
	return c.rest.R().
		SetHeader("Content-Type", "application/json").
		SetError(&ErrorResponse{}).
		SetBody(`{}`) // default value for POST
}

func (c *Client) checkResponse(resp *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		errBody, ok := resp.Error().(*ErrorResponse)
		if ok {
			return fmt.Errorf("non 200 status: %v, msg: %v, err: %v", resp.StatusCode(), errBody.Msg, errBody.Err)
		}
		return fmt.Errorf("non 200 status: %v", resp.StatusCode())
	}

	return nil
}

func (c *Client) newSession() *Session {
	return &Session{
		client: c,
	}
}

func (c *Client) StartSession() (*Session, error) {
	session := c.newSession()
	respBody := &SessionResponse{Data: session}

	resp, err := c.newRequest().
		SetResult(respBody).
		Post(c.v1SessionPath())

	if err := c.checkResponse(resp, err); err != nil {
		return nil, err
	}

	return session, nil
}

func (c *Client) joinSession(sessionId string, body *ParticipantJoinBody, session *Session) (*Participant, error) {
	participant := &Participant{session: session}
	respBody := &ParticipantResponse{Data: participant}

	resp, err := c.newRequest().
		SetBody(body).
		SetResult(respBody).
		Post(c.v1SessionPath() + "/" + sessionId + "/join")

	if err := c.checkResponse(resp, err); err != nil {
		return nil, err
	}

	return participant, nil
}

func (c *Client) JoinSession(sessionId string, body *ParticipantJoinBody) (*Participant, error) {
	session := c.newSession()
	session.Id = sessionId

	return c.joinSession(sessionId, body, session)
}

func (c *Client) partialCommit(sessionId string, body *ParticipantCommit, participant *Participant) (*Participant, error) {
	respBody := &ParticipantResponse{Data: participant}

	resp, err := c.newRequest().
		SetBody(body).
		SetResult(respBody).
		Post(c.v1SessionPath() + "/" + sessionId + "/partial-commit")

	if err := c.checkResponse(resp, err); err != nil {
		return nil, err
	}

	return participant, err
}

func (c *Client) PartialCommit(sessionId string, body *ParticipantCommit) (*Participant, error) {
	session := c.newSession()
	session.Id = sessionId
	participant := &Participant{session: session}

	return c.partialCommit(sessionId, body, participant)
}

func (c *Client) commitSession(sessionId string, session *Session) (*Session, error) {
	respBody := &SessionResponse{Data: session}

	resp, err := c.rest.R().
		SetResult(respBody).
		Post(c.v1SessionPath() + "/" + sessionId + "/commit")

	if err := c.checkResponse(resp, err); err != nil {
		return nil, err
	}

	return session, nil
}

func (c *Client) CommitSession(sessionId string) (*Session, error) {
	session := c.newSession()
	// session.Id = sessionId

	return c.commitSession(sessionId, session)
}

func (c *Client) abortSession(sessionId string, session *Session) (*Session, error) {
	respBody := &SessionResponse{Data: session}

	resp, err := c.rest.R().
		SetResult(respBody).
		Post(c.v1SessionPath() + "/" + sessionId + "/abort")

	if err := c.checkResponse(resp, err); err != nil {
		return nil, err
	}

	return session, nil
}

func (c *Client) AbortSession(sessionId string) (*Session, error) {
	session := c.newSession()
	// session.Id = sessionId

	return c.abortSession(sessionId, session)
}

func (s *Session) JoinSession(body *ParticipantJoinBody) (*Participant, error) {
	participant, err := s.client.joinSession(s.Id, body, s)

	if err != nil {
		return nil, err
	}

	return participant, nil
}

func (p *Participant) PartialCommit(compensate *ParticipantAction, complete *ParticipantAction) error {
	session := p.session

	_, err := session.client.partialCommit(session.Id, &ParticipantCommit{
		Id:         p.Id,
		Compensate: compensate,
		Complete:   complete,
	}, p)

	if err != nil {
		return err
	}

	return nil
}

func (s *Session) CommitSession() error {
	_, err := s.client.commitSession(s.Id, s)

	if err != nil {
		return err
	}

	return nil
}

func (s *Session) AbortSession() error {
	_, err := s.client.abortSession(s.Id, s)

	if err != nil {
		return nil
	}

	return nil
}
