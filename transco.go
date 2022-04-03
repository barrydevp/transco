package transco

import (
	"github.com/go-resty/resty/v2"
)

const (
	DefaultUri      = "http://transcoorditor:8000"
	V1PrefixApiPath = "api/v1"
)

type Client struct {
	conn *Connection
}

func New(uri string) (*Client, error) {
	if uri == "" {
		uri = DefaultUri
	}

	conn, err := NewConn(uri)
	if err != nil {
		return nil, err
	}

	if err = conn.Connect(); err != nil {
		return nil, err
	}

	return &Client{
		conn: conn,
	}, nil
}

func (c *Client) v1SessionPath() string {
	return V1PrefixApiPath + "/sessions"
}

func (c *Client) newSession() *Session {
	return &Session{
		client: c,
	}
}

func (c *Client) StartSession() (*Session, error) {
	session := c.newSession()

	_, err := c.conn.request(func(req *RestRequest) (*resty.Response, error) {
		return req.
			SetResult(session).
			Post(c.v1SessionPath())
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (c *Client) SessionFromId(sessionId string) (*Session, error) {
	session := c.newSession()
	_, err := c.conn.request(func(req *RestRequest) (*resty.Response, error) {
		return req.
			SetResult(session).
			Get(c.v1SessionPath() + "/" + sessionId)
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (c *Client) joinSession(sessionId string, body *ParticipantJoinBody, session *Session) (*Participant, error) {
	participant := &Participant{session: session}

	_, err := c.conn.request(func(req *RestRequest) (*resty.Response, error) {
		return req.
			SetResult(participant).
			SetBody(body).
			Post(c.v1SessionPath() + "/" + sessionId + "/join")
	})
	if err != nil {
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
	_, err := c.conn.request(func(req *RestRequest) (*resty.Response, error) {
		return req.
			SetResult(participant).
			SetBody(body).
			Post(c.v1SessionPath() + "/" + sessionId + "/partial-commit")
	})
	if err != nil {
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
	_, err := c.conn.request(func(req *RestRequest) (*resty.Response, error) {
		return req.
			SetResult(session).
			Post(c.v1SessionPath() + "/" + sessionId + "/commit")
	})
	if err != nil {
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
	_, err := c.conn.request(func(req *RestRequest) (*resty.Response, error) {
		return req.
			SetResult(session).
			Post(c.v1SessionPath() + "/" + sessionId + "/abort")
	})
	if err != nil {
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
		return err
	}

	return nil
}
