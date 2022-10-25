package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gitlab.qredo.com/computational-custodian/signing-agent/defs"
)

type mockActionManager struct {
	ApproveCalled bool
	RejectCalled  bool
	LastActionId  string
	NextError     error
}

func (m *mockActionManager) Approve(actionID string) error {
	m.ApproveCalled = true
	m.LastActionId = actionID
	return m.NextError
}
func (m *mockActionManager) Reject(actionID string) error {
	m.RejectCalled = true
	m.LastActionId = actionID
	return m.NextError
}

func TestActionHandler_ActionApprove_empty_actionId(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("PUT", "/client/action/ ", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var err error
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		_, err = NewActionHandler(actionManagerMock).ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, actionManagerMock.ApproveCalled)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "empty actionID", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestActionHandler_ActionApprove(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("PUT", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var err error
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		_, err = NewActionHandler(actionManagerMock).ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, actionManagerMock.ApproveCalled)
	assert.Nil(t, err)
}

func TestActionHandler_ActionReject_empty_actionId(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("DELETE", "/client/action/ ", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var err error
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		_, err = NewActionHandler(actionManagerMock).ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.False(t, actionManagerMock.RejectCalled)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "empty actionID", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestActionHandler_ActionReject(t *testing.T) {
	//Arrange
	actionManagerMock := &mockActionManager{}
	req, _ := http.NewRequest("DELETE", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var err error
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		_, err = NewActionHandler(actionManagerMock).ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, actionManagerMock.RejectCalled)
	assert.Nil(t, err)
}
