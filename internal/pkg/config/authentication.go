package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type User struct {
	Name         string `json:"name"          yaml:"name"`
	PasswordHash string `json:"password hash" yaml:"password hash"`
}

type Authentication struct {
	Users   []User          `json:"users" yaml:"users"`
	userMap map[string]User `json:"-"     yaml:"-"`
}

func NewAuthentication() *Authentication {
	auth := new(Authentication)
	auth.userMap = make(map[string]User)
	return auth
}

func (auth *Authentication) ParseFromRaw(data []byte) error {
	err := yaml.Unmarshal(data, auth)
	if err != nil {
		return err
	}
	if len(auth.Users) == 0 {
		return fmt.Errorf("no record users")
	}
	for _, user := range auth.Users {
		if _, ok := auth.userMap[user.Name]; ok {
			return fmt.Errorf("duplicated user names")
		}
		auth.userMap[user.Name] = user
	}
	return nil
}

func (auth *Authentication) Read(confFileName string) error {
	if data, err := os.ReadFile(confFileName); err != nil {
		return err
	} else {
		if err := auth.ParseFromRaw(data); err != nil {
			return err
		}
	}
	return nil
}

var (
	UnauthorizedError = errors.New("Unauthorized")
)

func (auth *Authentication) Authenticate(name, pass string) (*User, error) {
	if user, ok := auth.userMap[name]; !ok {
		return nil, UnauthorizedError
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pass)); err != nil {
			wwlog.Warn("%v\n", err)
			return nil, UnauthorizedError
		}
		return &user, nil
	}
}
