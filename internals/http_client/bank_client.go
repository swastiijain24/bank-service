package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
		BaseURL :url,
		HTTPClient : &http.Client{
			Timeout: 10 * time.Second,
			Transport: transport,
		},
	}
}

func (c *BankClient) CallDebit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64, mpinHash string) (string,string ,  error){

	url := fmt.Sprintf("%s/transactions/debit", c.BaseURL)

	body , _ := json.Marshal(map[string]interface{}{
		"external_id" : transactionId,
		"from_account_id": payerId,
		"to_account_identifier" : payeeId,
		"amount":amount,
		"mpin_hash":mpinHash,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "txn_"+transactionId)
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))


	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "","", err 
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return "","", fmt.Errorf("bank returned error :%d", resp.StatusCode) 
	}

	var result BankTxnResponse 
	json.NewDecoder(resp.Body).Decode(&result)
	return result.bankReferenceId, result.status,  nil 
}

func (c *BankClient) CallCredit(ctx context.Context, transactionId string, payerId string, payeeId string, amount int64) (string,string ,  error){

	url := fmt.Sprintf("%s/transactions/credit", c.BaseURL)

	body , _ := json.Marshal(map[string]interface{}{
		"external_id" : transactionId,
		"from_account_id": payerId,
		"to_account_identifier" : payeeId,
		"amount":amount,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "txn_"+transactionId)
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))


	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "","", err 
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return "","", fmt.Errorf("bank returned error :%d", resp.StatusCode) 
	}

	var result BankTxnResponse 
	json.NewDecoder(resp.Body).Decode(&result)
	return result.bankReferenceId, result.status,  nil 
}

func (c *BankClient)  GetStatusFromBank(ctx context.Context, transactionId string) (string, string, error){
	url := fmt.Sprintf("%s/transactions/status/%s", c.BaseURL, transactionId)

	body , _ := json.Marshal(map[string]interface{}{
		"external_id" : transactionId,
	})

	req, _ := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("NPCI_API_KEY"))


	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "","", err 
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return "","", fmt.Errorf("bank returned error :%d", resp.StatusCode) 
	}

	var result BankTxnResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return  result.bankReferenceId, result.status,  nil 
}


type BankTxnResponse struct{
	bankReferenceId string 
	status string 
	created_at  time.Time
}