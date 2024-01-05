package server

import (
	"fmt"
	train "github.com/playbody/train-ticket-service/proto"
	"net/mail"
	"strings"
)

func isValidUser(user *train.User) (bool, error) {
	if strings.TrimSpace(user.FirstName) == "" {
		return true, fmt.Errorf("first name must not be empty")
	}
	if strings.TrimSpace(user.LastName) == "" {
		return true, fmt.Errorf("last name must not be empty")
	}
	_, err := mail.ParseAddress(user.Email)
	return err == nil, err
}
