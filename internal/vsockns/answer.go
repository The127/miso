package vsockns

import (
	"errors"
	"fmt"
	"strings"
)

// the mark every answer of the helper starts with, so success never has to
// be guessed from what an answer says
const (
	okMark     = "+"
	failedMark = "-"
)

// answerOK tells miso a question worked, with the files it hands over.
func answerOK(conn int, text string, files ...int) error {
	return say(conn, okMark+text, files...)
}

// answerFailed tells miso why a question did not work.
func answerFailed(conn int, failed error) error {
	return say(conn, failedMark+failed.Error())
}

// answer is the helper's next answer: what it said and handed over, or why
// the question failed.
func (n *Namespace) answer() (string, []int, error) {
	said, files, err := hear(n.conn)
	if err != nil {
		return "", nil, err
	}

	if reason, failed := strings.CutPrefix(said, failedMark); failed {
		return "", nil, errors.New(reason)
	}

	text, ok := strings.CutPrefix(said, okMark)
	if !ok {
		return "", nil, fmt.Errorf("the helper answered without a mark: %q", said)
	}

	return text, files, nil
}

// ask puts a question to the helper, with the files it hands over, and
// takes its answer.
func (n *Namespace) ask(question string, files ...int) (string, []int, error) {
	if err := say(n.conn, question, files...); err != nil {
		return "", nil, err
	}

	return n.answer()
}
