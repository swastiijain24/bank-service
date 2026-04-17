package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type BankClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewBankClient(url string) *BankClient {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &BankClient{
		BaseURL: url,
		HTTPClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

func (c *BankClient) MakeRequest(ctx context.Context, transactionId string, body []byte, opType string) (string, string, error) {
	url := fmt.Sprintf("%s/transactions/%s", c.BaseURL, strings.ToLower(opType))
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "txn_"+transactionId+"_"+opType)
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("bank returned error :%d", resp.StatusCode)
	}

	var result BankTxnResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.bankReferenceId, result.status, nil
}

func (c *BankClient) CallDebit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64, mpinHash string) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"external_id":           transactionId,
		"from_account_id":       payerId,
		"to_account_identifier": payeeId,
		"amount":                amount,
		"mpin_hash":             mpinHash,
	})
	return c.MakeRequest(ctx, transactionId, body, "DEBIT")

}

func (c *BankClient) CallCredit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"external_id":           transactionId,
		"from_account_id":       payerId,
		"to_account_identifier": payeeId,
		"amount":                amount,
	})
	return c.MakeRequest(ctx, transactionId, body, "CREDIT")
}

func (c *BankClient) CallRefund(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"external_id":           transactionId,
		"from_account_id":       payerId,
		"to_account_identifier": payeeId,
		"amount":                amount,
	})
	return c.MakeRequest(ctx, transactionId, body, "REFUND")
}

func (c *BankClient) GetStatusFromBank(ctx context.Context, transactionId string) (string, string, error) {
	url := fmt.Sprintf("%s/transactions/status/%s", c.BaseURL, transactionId)

	body, _ := json.Marshal(map[string]interface{}{
		"external_id": transactionId,
	})

	req, _ := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("bank returned error :%d", resp.StatusCode)
	}

	var result BankTxnResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.bankReferenceId, result.status, nil
}

type BankTxnResponse struct {
	bankReferenceId string
	status          string
	created_at      time.Time
}
