package mpesa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	sandboxBaseURL    = "https://sandbox.safaricom.co.ke"
	productionBaseURL = "https://api.safaricom.co.ke"

	oauthEndpoint    = "/oauth/v1/generate?grant_type=client_credentials"
	stkPushEndpoint  = "/mpesa/stkpush/v1/processrequest"
	stkQueryEndpoint = "/mpesa/stkpushquery/v1/query"

	// TransactionType for Lipa Na M-Pesa Online
	TransactionTypeCustomerPayBillOnline  = "CustomerPayBillOnline"
	TransactionTypeCustomerBuyGoodsOnline = "CustomerBuyGoodsOnline"
)

// Client is a stateless Daraja API client.
// Credentials are passed per call to support multi-tenancy.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Daraja API client for the given environment.
// environment should be "sandbox" or "production".
func NewClient(environment string) *Client {
	base := sandboxBaseURL
	if environment == "production" {
		base = productionBaseURL
	}
	return &Client{
		baseURL: base,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// --- OAuth ---

// AccessTokenResponse is the response from the Daraja OAuth endpoint.
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// GetAccessToken fetches an OAuth access token using the consumer key and secret.
func (c *Client) GetAccessToken(consumerKey, consumerSecret string) (string, error) {
	url := c.baseURL + oauthEndpoint

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("mpesa: failed to create auth request: %w", err)
	}

	// Basic Auth: base64(consumerKey:consumerSecret)
	credentials := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("mpesa: auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("mpesa: failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mpesa: auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp AccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("mpesa: failed to parse auth response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// --- STK Push ---

// STKPushRequest is the request body for initiating an STK Push.
type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            int    `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

// STKPushResponse is the response from the STK Push endpoint.
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

// STKPush initiates an STK Push (Lipa Na M-Pesa Online) request.
func (c *Client) STKPush(accessToken string, req STKPushRequest) (*STKPushResponse, error) {
	url := c.baseURL + stkPushEndpoint

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("mpesa: failed to marshal STK push request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("mpesa: failed to create STK push request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("mpesa: STK push request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mpesa: failed to read STK push response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mpesa: STK push failed with status %d: %s", resp.StatusCode, string(body))
	}

	var stkResp STKPushResponse
	if err := json.Unmarshal(body, &stkResp); err != nil {
		return nil, fmt.Errorf("mpesa: failed to parse STK push response: %w", err)
	}

	return &stkResp, nil
}

// --- STK Push Query ---

// STKQueryRequest is the request body for querying an STK Push transaction.
type STKQueryRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	CheckoutRequestID string `json:"CheckoutRequestID"`
}

// STKQueryResponse is the response from the STK Push query endpoint.
type STKQueryResponse struct {
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResultCode          string `json:"ResultCode"`
	ResultDesc          string `json:"ResultDesc"`
}

// STKPushQuery queries the status of an STK Push transaction.
func (c *Client) STKPushQuery(accessToken string, req STKQueryRequest) (*STKQueryResponse, error) {
	url := c.baseURL + stkQueryEndpoint

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("mpesa: failed to marshal query request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("mpesa: failed to create query request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("mpesa: query request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mpesa: failed to read query response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mpesa: query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var queryResp STKQueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("mpesa: failed to parse query response: %w", err)
	}

	return &queryResp, nil
}

// --- Callback ---

// STKCallbackBody is the top-level structure Safaricom sends to the callback URL.
type STKCallbackBody struct {
	Body struct {
		StkCallback STKCallback `json:"stkCallback"`
	} `json:"Body"`
}

// STKCallback contains the callback details from Safaricom.
type STKCallback struct {
	MerchantRequestID string            `json:"MerchantRequestID"`
	CheckoutRequestID string            `json:"CheckoutRequestID"`
	ResultCode        int               `json:"ResultCode"`
	ResultDesc        string            `json:"ResultDesc"`
	CallbackMetadata  *CallbackMetadata `json:"CallbackMetadata,omitempty"`
}

// CallbackMetadata holds the metadata items in the callback.
type CallbackMetadata struct {
	Item []CallbackMetadataItem `json:"Item"`
}

// CallbackMetadataItem is a key-value pair in callback metadata.
type CallbackMetadataItem struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value"`
}

// GetMpesaReceiptNumber extracts the M-Pesa receipt number from callback metadata.
func (cb *STKCallback) GetMpesaReceiptNumber() string {
	if cb.CallbackMetadata == nil {
		return ""
	}
	for _, item := range cb.CallbackMetadata.Item {
		if item.Name == "MpesaReceiptNumber" {
			if v, ok := item.Value.(string); ok {
				return v
			}
		}
	}
	return ""
}

// GetTransactionDate extracts the transaction date from callback metadata.
func (cb *STKCallback) GetTransactionDate() string {
	if cb.CallbackMetadata == nil {
		return ""
	}
	for _, item := range cb.CallbackMetadata.Item {
		if item.Name == "TransactionDate" {
			// Safaricom sends this as a float64 number (e.g. 20191219102115)
			if v, ok := item.Value.(float64); ok {
				return fmt.Sprintf("%.0f", v)
			}
		}
	}
	return ""
}

// --- Helpers ---

// GeneratePassword generates the password for STK Push requests.
// Password = base64(shortcode + passkey + timestamp)
func GeneratePassword(shortCode, passKey, timestamp string) string {
	return base64.StdEncoding.EncodeToString([]byte(shortCode + passKey + timestamp))
}

// GenerateTimestamp returns the current timestamp in the format required by Daraja (YYYYMMDDHHmmss).
func GenerateTimestamp() string {
	loc, _ := time.LoadLocation("Africa/Nairobi")
	return time.Now().In(loc).Format("20060102150405")
}
