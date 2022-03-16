package transco

import "testing"

const (
    TRANSCOORDITOR_URL = "http://localhost:8000"
)

func printErr(t *testing.T, err error) {
	t.Errorf("Error occurred [%v]", err)
}

func assertError(t *testing.T, err error) {
	if err != nil {
		printErr(t, err)
	}
}

func session() (*Session, error) {
	return New(TRANSCOORDITOR_URL).StartSession()
}

func TestSimpleCommitSession(t *testing.T) {
	session, err := session()

	if err != nil {
		printErr(t, err)
		return
	}

	testCases := []struct {
		PartJoinBody      *ParticipantJoinBody
		ParticipantCommit *ParticipantCommit
	}{
		{
			PartJoinBody: &ParticipantJoinBody{
				ClientId:  "c-1",
				RequestId: "1",
			},
			ParticipantCommit: &ParticipantCommit{},
		},
		{
			PartJoinBody: &ParticipantJoinBody{
				ClientId:  "c-2",
				RequestId: "2",
			},
			ParticipantCommit: &ParticipantCommit{},
		},
		{
			PartJoinBody: &ParticipantJoinBody{
				ClientId:  "c-3",
				RequestId: "3",
			},
			ParticipantCommit: &ParticipantCommit{},
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
		PartJoinBody      *ParticipantJoinBody
		ParticipantCommit *ParticipantCommit
	}{
		{
			PartJoinBody: &ParticipantJoinBody{
				ClientId:  "c-1",
				RequestId: "1",
			},
			ParticipantCommit: &ParticipantCommit{},
		},
		{
			PartJoinBody: &ParticipantJoinBody{
				ClientId:  "c-2",
				RequestId: "2",
			},
			ParticipantCommit: &ParticipantCommit{},
		},
		{
			PartJoinBody: &ParticipantJoinBody{
				ClientId:  "c-3",
				RequestId: "3",
			},
			ParticipantCommit: &ParticipantCommit{},
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
