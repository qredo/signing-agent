package lib

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/custody-engine/automated-approver/config"
	"gitlab.qredo.com/custody-engine/automated-approver/util"
	"testing"
)

// mock the serve function
func mocServe(fakeError error) ServeCB {
	return func(url string, handler Handler, errHandler ErrHandler) (doneCH, stopCH chan struct{}, err error) {
		data := []byte(`{
		"id":"9b1deb4d-3b7d",
		"coreClientID":"cc229a7e3dd3",
		"type":"Action",
		"status":"OK",
		"timestamp":1665492366,
		"expireTime":1675492366
	}`)
		doneCH = make(chan struct{})
		stopCH = make(chan struct{})
		go func() {
			<-stopCH
			close(doneCH)
		}()
		handler(data)
		if fakeError != nil {
			errHandler(fakeError)
		}
		return doneCH, stopCH, nil
	}
}

func TestFeedRead(t *testing.T) {
	cfg := config.Base{QredoAPIDomain: "mock_domain", QredoAPIBasePath: "mock_path"}
	agent, err := New(&cfg, util.NewFileStore("mock.db"))
	assert.NoError(t, err)
	fakeError := errors.New("fake error")
	doneCH, stopCH, err := NewFeed("mock_url", agent, mocServe(fakeError)).ActionEvent(
		func(e *WsActionInfoEvent) {
			assert.Equal(t, "9b1deb4d-3b7d", e.ID)
			assert.Equal(t, "cc229a7e3dd3", e.AgentID)
			assert.Equal(t, "Action", e.Type)
			assert.Equal(t, "OK", e.Status)
			assert.Equal(t, int64(1665492366), e.Timestamp)
			assert.Equal(t, int64(1675492366), e.ExpireTime)
		},
		func(err error) {
			assert.Equal(t, err, fakeError)
		},
	)
	assert.NoError(t, err)
	stopCH <- struct{}{}
	<-doneCH
}
