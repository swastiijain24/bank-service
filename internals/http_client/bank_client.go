package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type HttpClient interface {
	MakeRequest(ctx context.Context, transactionId string, body []byte, opType string) (string, string, error)
	CallDebit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64, mpinHash string) (string, string, error)
	CallCredit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string, string, error)
	CallRefund(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string, string, error)
	GetStatusFromBank(ctx context.Context, transactionId string, transactionType string) (string, string, error)
	DiscoverAccounts(ctx context.Context, phone string, ) ([]string, error)
	SetMpin(ctx context.Context, accountId string, mpinEn string) error
	ChangeMpin(ctx context.Context, accountId string,  oldMpinEn string, newMpinEn string) error
	GetBalance(ctx context.Context, accountId string,  mpinEn string) (int64, error)
}

type BankClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewBankClient(url string) HttpClient {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, //must set to false in production
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
	log.Print("made request")
	if err != nil {
		log.Print(err)
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("bank returned error :%d", resp.StatusCode)
	}

	var result BankTxnResponse
	json.NewDecoder(resp.Body).Decode(&result)
	log.Print("debit done 12")
	log.Print("credit done 22")
	log.Print(result)
	return result.BankReferenceID, result.Status, nil
}

func (c *BankClient) CallDebit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64, mpinHash string) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"from_account_id": payerId,
		"to_account_id":   payeeId,
		"amount":          amount,
		"mpin_hash":       mpinHash,
		"external_id":     transactionId,
	})
	log.Print("json req marshaling 9")
	return c.MakeRequest(ctx, transactionId, body, "DEBIT")
}

func (c *BankClient) CallCredit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"from_account_id": payerId,
		"to_account_id":   payeeId,
		"amount":          amount,
		"external_id":     transactionId,
	})
	log.Print("marshaled json credit 18")
	return c.MakeRequest(ctx, transactionId, body, "CREDIT")
}

func (c *BankClient) CallRefund(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"from_account_id": payerId,
		"to_account_id":   payeeId,
		"amount":          amount,
		"external_id":     transactionId,
	})
	return c.MakeRequest(ctx, transactionId, body, "REFUND")
}

func (c *BankClient) GetStatusFromBank(ctx context.Context, transactionId string, transactionType string) (string, string, error) {
	url := fmt.Sprintf("%s/transactions/status/%s", c.BaseURL, transactionId)

	body, _ := json.Marshal(map[string]interface{}{
		"external_id":      transactionId,
		"transaction_type": transactionType,
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
	return result.BankReferenceID, result.Status, nil
}

func (c *BankClient) DiscoverAccounts(ctx context.Context, phone string, ) ([]string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"phone":     phone,
	})

	url := fmt.Sprintf("%s/accounts/discover", c.BaseURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("bank returned error :%d", resp.StatusCode)
	}

	var result []string
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (c *BankClient) SetMpin(ctx context.Context, accountId string, mpinEn string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"mpin_encrypted":mpinEn,
	})

	url := fmt.Sprintf("%s/accounts/mpin/%s", c.BaseURL, accountId)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))

	_, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	return nil 
}

func (c *BankClient) ChangeMpin(ctx context.Context, accountId string, oldMpinEn string, newMpinEn string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"old_mpin_encrypted":oldMpinEn,
		"new_mpin_encrypted":newMpinEn,
	})

	url := fmt.Sprintf("%s/account/mpin/%s", c.BaseURL, accountId)
	req, _ := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))

	_, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	return nil 
}

func (c *BankClient) GetBalance(ctx context.Context, accountId string, mpinEn string) (int64, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"mpin_encrypted":mpinEn,
	})

	url := fmt.Sprintf("%s/account/balance/%s", c.BaseURL, accountId)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0,  err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bank returned error :%d", resp.StatusCode)
	}

	var result int64
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

type BankTxnResponse struct {
	BankReferenceID string    `json:"bank_reference_id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
