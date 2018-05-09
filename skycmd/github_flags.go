package skycmd

import (
	"encoding/json"
	"errors"

	"github.com/coreos/dex/connector/github"
	"github.com/hashicorp/go-multierror"
)

func init() {
	RegisterConnector(&Connector{
		id:         "github",
		config:     &GithubFlags{},
		teamConfig: &GithubTeamFlags{},
	})
}

type GithubFlags struct {
	ClientID     string `long:"client-id" description:"Client id"`
	ClientSecret string `long:"client-secret" description:"Client secret"`
}

func (self *GithubFlags) Name() string {
	return "GitHub"
}

func (self *GithubFlags) Validate() error {
	var errs *multierror.Error

	if self.ClientID == "" {
		errs = multierror.Append(errs, errors.New("Missing client-id"))
	}

	if self.ClientSecret == "" {
		errs = multierror.Append(errs, errors.New("Missing client-secret"))
	}

	return errs.ErrorOrNil()
}

func (self *GithubFlags) Serialize(redirectURI string) ([]byte, error) {
	if err := self.Validate(); err != nil {
		return nil, err
	}

	return json.Marshal(github.Config{
		ClientID:     self.ClientID,
		ClientSecret: self.ClientSecret,
		RedirectURI:  redirectURI,
	})
}

type GithubTeamFlags struct {
	Users  []string `json:"users" long:"user" description:"List of whitelisted GitHub users" value-name:"USERNAME"`
	Groups []string `json:"groups" long:"group" description:"List of whitelisted GitHub groups (e.g. my-org or my-org:my-team)" value-name:"ORG_NAME:TEAM_NAME"`
}

func (self *GithubTeamFlags) IsValid() bool {
	return len(self.Users) > 0 || len(self.Groups) > 0
}

func (self *GithubTeamFlags) GetUsers() []string {
	return self.Users
}

func (self *GithubTeamFlags) GetGroups() []string {
	return self.Groups
}
