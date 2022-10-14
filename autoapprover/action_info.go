package autoapprover

import (
	"encoding/json"
	"time"
)

type Parser interface {
	Parse() string
}

type ActionInfo struct {
	ID         string `json:"id"`
	AgentID    string `json:"coreClientID"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
	ExpireTime int64  `json:"expireTime"`
}

func (a *ActionInfo) Parse() string {
	out, _ := json.Marshal(a)
	return string(out)
}

func (a *ActionInfo) IsNotExpired() bool {
	return a.ExpireTime > time.Now().Unix()
}
