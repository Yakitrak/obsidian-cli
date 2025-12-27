package cmd

import (
	"bufio"
	"fmt"
	"strings"
)

type promptAction int

const (
	actionNone promptAction = iota
	actionBack
	actionSkip
	actionHelp
)

func promptLine(in *bufio.Reader, prompt string) (string, promptAction, error) {
	fmt.Print(prompt)
	line, err := in.ReadString('\n')
	if err != nil {
		return "", actionNone, err
	}
	line = strings.TrimSpace(line)
	switch strings.ToLower(line) {
	case "back":
		return "", actionBack, nil
	case "skip":
		return "", actionSkip, nil
	case "?":
		return "", actionHelp, nil
	default:
		return line, actionNone, nil
	}
}

func isYes(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}
