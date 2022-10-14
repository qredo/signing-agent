package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"

	"gitlab.qredo.com/custody-engine/automated-approver/api"
	"gitlab.qredo.com/custody-engine/automated-approver/rest"
)

const (
	QredoBE  = "https://play-api.qredo.network/api/v1/p"
	AgentID2 = "43CQtjpeSH1DeMSWQTPc3de2GWLsTAMcvGePYS6wKibn" // additional member added to policy

	// payloads for Qredo BE provided via templates
	TestCreateCompany = "../../testdata/e2e/createcompany.templ"
	TestCreateFund    = "../../testdata/e2e/createfund.templ"
	TestTrustedParty  = "../../testdata/e2e/trustedparty.templ"
	TestUpdatePolicy  = "../../testdata/e2e/updatePolicy.templ"
)

type (
	Wallets struct {
		WalletID      string `json:"wallet_id"`
		Asset         string `json:"asset"`
		Address       string `json:"address"`
		AddressType   string `json:"address_type"`
		ShortCode     string `json:"short_code"`
		Balance       int    `json:"balance"`
		PolicyWithraw Policy `json:"policy_withdraw"`
		PolicyTX      Policy `json:"policy_tx"`
	}

	Policy struct {
		ID      string `json:"id"`
		Members []struct {
			Entity struct {
				ID string `json:"id"`
			} `json:"entity"`
		} `json:"members"`
	}
)

// TestActionAutoApprove is used to confirm that a signing agent is used to automatically approve a policy update request.
// A new agent is registered and assigned to a newly created fund.  An additional entity is added to the fund and
// checked to confirm it has two members.  Chain update race conditions are avoided with short delays.
func TestActionAutoApprove(t *testing.T) {

	cfg := createTestConfig()
	cfg.AutoApprove.Enabled = true
	server := getTestServer(cfg)

	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()

	e := httpexpect.New(t, server.URL)

	// register agent
	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// register a new agent to use for the approval
	registrationResponse := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()
	agentID := registrationResponse.Object().Value("agentId").Raw().(string)
	assert.NotEqual(t, "", agentID)

	time.Sleep(2 * time.Second)

	// create a company
	companyID, err := createCompany()
	assert.NoError(t, err)
	assert.NotEqual(t, "", companyID, "companyID should not be empty string")

	// add agent as a trusted member
	err = trustedParty(agentID, companyID)
	assert.NoError(t, err)

	// create a fund with wallets
	fundID, err := createFund(agentID, companyID)
	assert.NoError(t, err)
	assert.NotEqual(t, "", fundID, "fundID should not be empty string")

	// get wallets
	wallets, err := getWallets(companyID, fundID)
	assert.NoError(t, err)
	assert.NotEmpty(t, wallets, "no wallets returned")

	// confirm just one wallet
	assert.Equal(t, 1, len(wallets))

	// confirm TX policy has just one member
	assert.Equal(t, 1, len(wallets[0].PolicyTX.Members))

	// add AgentID2 as a second member to the wallet
	err = trustedParty(AgentID2, companyID)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	members := fmt.Sprintf("\"%s\", \"%s\"", agentID, AgentID2)
	err = updatePolicy(wallets[0].PolicyTX.ID, members, companyID, fundID, wallets[0].WalletID)
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	// get the wallets again and confirm TX policy has 2 members
	wallets, err = getWallets(companyID, fundID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(wallets[0].PolicyTX.Members))
}

func createCompany() (string, error) {
	tmpl, _ := template.ParseFiles(TestCreateCompany)
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, nil); err != nil {
		return "", err
	}

	type CompanyResponse struct {
		CompanyID string `json:"company_id"`
		Ref       string `json:"ref"`
	}
	var c CompanyResponse

	path := "/company"
	_ = doBackEndCall(http.MethodPost, path, &tpl, &c)

	return c.CompanyID, nil
}

func createFund(agentID string, companyID string) (string, error) {
	tmpl, _ := template.ParseFiles(TestCreateFund)
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, agentID); err != nil {
		return "", err
	}

	var f struct {
		FundID               string `json:"fund_id"`
		CustodygroupWithdraw string `json:"custodygroup_withdraw"`
		CustodygroupTx       string `json:"custodygroup_tx"`
	}

	path := fmt.Sprintf("/company/%s/fund", companyID)
	err := doBackEndCall(http.MethodPost, path, &tpl, &f)
	if err != nil {
		return "", fmt.Errorf("cannot create fund: %v", err)
	}
	log.Println("fund: ", f)

	return f.FundID, nil
}

func getWallets(companyID string, fundID string) ([]Wallets, error) {

	type Fund struct {
		ID      string    `json:"ID"`
		Wallets []Wallets `json:"wallets"`
	}

	var resp Fund
	path := fmt.Sprintf("/company/%s/fund/%s", companyID, fundID)
	err := doBackEndCall(http.MethodGet, path, nil, &resp)
	if err != nil {
		return []Wallets{}, err
	}

	return resp.Wallets, nil
}

func trustedParty(agentID string, companyID string) error {
	tmpl, _ := template.ParseFiles(TestTrustedParty)
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, agentID); err != nil {
		return err
	}

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	path := fmt.Sprintf("/company/%s/trustedparty", companyID)
	_ = doBackEndCall(http.MethodPost, path, &tpl, &resp)
	log.Println("code:", resp.Code)

	if resp.Code != 200 {
		return fmt.Errorf("cannot assign trust %s", resp.Msg)
	}

	return nil
}

func updatePolicy(policyID string, members string, companyID string, fundID string, walletID string) error {
	type Policy struct {
		ID     string
		Member string
	}
	pol := Policy{policyID, members}

	tmpl, _ := template.ParseFiles(TestUpdatePolicy)
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, pol); err != nil {
		return err
	}

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	path := fmt.Sprintf("/company/%s/fund/%s/wallet/%s/policy", companyID, fundID, walletID)
	_ = doBackEndCall(http.MethodPut, path, &tpl, &resp)
	log.Println("code:", resp.Code)

	if resp.Code != 200 {
		return fmt.Errorf("cannot update policy: %s", resp.Msg)
	}

	return nil
}

// doBackEndCall calls the Qredo backend at path.  Headers, including the APIKEY, are set.
func doBackEndCall(method string, path string, data *bytes.Buffer, respData interface{}) error {
	addr := QredoBE + path
	log.Println("doBackEndCall: ", method, addr)

	var req *http.Request
	var err error
	if method == http.MethodGet {
		req, err = http.NewRequest(method, addr, nil)
	} else {
		req, err = http.NewRequest(method, addr, data)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", APIKey)
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	switch respData := respData.(type) {
	case nil:
		return nil
	default:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response body: %v", err)
		}
		if err := json.Unmarshal(b, respData); err != nil {
			return fmt.Errorf("decode response as JSON: %v", err)
		}
	}

	return nil
}
