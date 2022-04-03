package transco_test

import (
	"testing"

	"github.com/barrydevp/transco"
)

const (
	TRANSCOORDITOR_URI = "http://localhost:8000"
)

func printErr(t *testing.T, err error) {
	t.Errorf("Error occurred [%v]", err)
}

func assertError(t *testing.T, err error) {
	if err != nil {
		printErr(t, err)
	}
}

func session() (*transco.Session, error) {
	c, err := transco.New(TRANSCOORDITOR_URI)
	if err != nil {
		return nil, err
	}
	return c.StartSession()
}

func TestSimpleCommitSession(t *testing.T) {
	session, err := session()

	if err != nil {
		printErr(t, err)
		return
	}

	testCases := []struct {
		PartJoinBody      *transco.ParticipantJoinBody
		ParticipantCommit *transco.ParticipantCommit
	}{
		{
			PartJoinBody: &transco.ParticipantJoinBody{
				ClientId:  "c-1",
				RequestId: "1",
			},
			ParticipantCommit: &transco.ParticipantCommit{},
		},
		{
			PartJoinBody: &transco.ParticipantJoinBody{
				ClientId:  "c-2",
				RequestId: "2",
			},
			ParticipantCommit: &transco.ParticipantCommit{},
		},
		{
			PartJoinBody: &transco.ParticipantJoinBody{
				ClientId:  "c-3",
				RequestId: "3",
			},
			ParticipantCommit: &transco.ParticipantCommit{},
		},
	}

	for _, test := range testCases {
		part, err := session.JoinSession(test.PartJoinBody)

		if err != nil {
			session.AbortSession()
			printErr(t, err)
			return
		}

		err = part.PartialCommit(test.ParticipantCommit.Compensate, test.ParticipantCommit.Complete)

		if err != nil {
			session.AbortSession()
			printErr(t, err)
			return
		}
	}

	err = session.CommitSession()

	assertError(t, err)
}

func TestSimpleAbortSession(t *testing.T) {
	session, err := session()

	if err != nil {
		printErr(t, err)
		return
	}

	testCases := []struct {
		PartJoinBody      *transco.ParticipantJoinBody
		ParticipantCommit *transco.ParticipantCommit
	}{
		{
			PartJoinBody: &transco.ParticipantJoinBody{
				ClientId:  "c-1",
				RequestId: "1",
			},
			ParticipantCommit: &transco.ParticipantCommit{},
		},
		{
			PartJoinBody: &transco.ParticipantJoinBody{
				ClientId:  "c-2",
				RequestId: "2",
			},
			ParticipantCommit: &transco.ParticipantCommit{},
		},
		{
			PartJoinBody: &transco.ParticipantJoinBody{
				ClientId:  "c-3",
				RequestId: "3",
			},
			ParticipantCommit: &transco.ParticipantCommit{},
		},
	}

	for _, test := range testCases {
		part, err := session.JoinSession(test.PartJoinBody)

		if err != nil {
			session.AbortSession()
			printErr(t, err)
			return
		}

		err = part.PartialCommit(test.ParticipantCommit.Compensate, test.ParticipantCommit.Complete)

		if err != nil {
			session.AbortSession()
			printErr(t, err)
			return
		}
	}

	err = session.AbortSession()

	assertError(t, err)
}
