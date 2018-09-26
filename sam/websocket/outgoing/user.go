package outgoing

import (
	"encoding/json"
)

type (
	User struct {
		// Channel to part (nil) for ALL channels
		ID          string `json:"ID"`
		Name        string `json:"name"`
		Username    string `json:"username"`
		Handle      string `json:"handle"`
		Connections uint   `json:"connections"`
	}

	UserSet []*User
)

func (p *User) EncodeMessage() ([]byte, error) {
	return json.Marshal(Payload{User: p})
}

func (p *UserSet) EncodeMessage() ([]byte, error) {
	return json.Marshal(Payload{UserSet: p})
}
