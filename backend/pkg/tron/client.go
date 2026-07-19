package tron

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// Client wraps TronGrid HTTP API calls.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// TRC20Transfer is one item from the TronGrid TRC-20 transfer list.
type TRC20Transfer struct {
	TransactionID  string `json:"transaction_id"`
	TokenInfo      struct {
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
		Address  string `json:"address"`
	} `json:"token_info"`
	From           string `json:"from"`
	To             string `json:"to"`
	Type           string `json:"type"`
	Value          string `json:"value"` // smallest unit (e.g. 6 decimals for USDT)
	BlockTimestamp int64  `json:"block_timestamp"`
}

type trc20ListResp struct {
	Data []TRC20Transfer `json:"data"`
	Meta struct {
		PageSize int `json:"page_size"`
	} `json:"meta"`
}

// GetIncomingTRC20Transfers returns all incoming TRC-20 transfers to address for the given contract
// since minTimestampMs (milliseconds).
func (c *Client) GetIncomingTRC20Transfers(ctx context.Context, address, contractAddress string, minTimestampMs int64) ([]TRC20Transfer, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20?contract_address=%s&only_to=true&min_timestamp=%d&limit=200",
		c.baseURL, address, contractAddress, minTimestampMs)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result trc20ListResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

type txDetail struct {
	Data []struct {
		TxID        string `json:"txID"`
		BlockNumber int64  `json:"blockNumber"`
		RawData     struct {
			Data string `json:"data"` // hex-encoded memo bytes
		} `json:"raw_data"`
		Ret []struct {
			ContractRet string `json:"contractRet"`
		} `json:"ret"`
	} `json:"data"`
}

// GetTransactionDetail returns block number and decoded memo string for a txHash.
// Returns blockNumber=0 and memo="" when not confirmed or no memo.
func (c *Client) GetTransactionDetail(ctx context.Context, txHash string) (blockNumber int64, memo string, err error) {
	url := fmt.Sprintf("%s/v1/transactions/%s", c.baseURL, txHash)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", err
	}
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	var detail txDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return 0, "", err
	}
	if len(detail.Data) == 0 {
		return 0, "", nil
	}
	tx := detail.Data[0]
	blockNumber = tx.BlockNumber
	if tx.RawData.Data != "" {
		b, err := hex.DecodeString(tx.RawData.Data)
		if err == nil {
			memo = string(b)
		}
	}
	return blockNumber, memo, nil
}

type nowBlockResp struct {
	BlockHeader struct {
		RawData struct {
			Number int64 `json:"number"`
		} `json:"raw_data"`
	} `json:"block_header"`
}

// GetCurrentBlockNumber returns the latest confirmed block number.
func (c *Client) GetCurrentBlockNumber(ctx context.Context) (int64, error) {
	url := fmt.Sprintf("%s/wallet/getnowblock", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result nowBlockResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.BlockHeader.RawData.Number, nil
}

// TriggerResult is the response from /wallet/triggersmartcontract.
type TriggerResult struct {
	Transaction json.RawMessage `json:"transaction"`
	RawDataHex  string          `json:"-"`
}

// TriggerTRC20Transfer creates an unsigned TRC-20 transfer transaction via TronGrid.
// Returns the transaction JSON (to be signed and broadcast) and the raw_data_hex for signing.
func (c *Client) TriggerTRC20Transfer(ctx context.Context, from, to, contractAddr string, amount *big.Int) (*TriggerResult, string, error) {
	// ABI-encode transfer(address,uint256) parameters
	param, err := encodeTRC20TransferParam(to, amount)
	if err != nil {
		return nil, "", err
	}

	body := map[string]interface{}{
		"owner_address":     from,
		"contract_address":  contractAddr,
		"function_selector": "transfer(address,uint256)",
		"parameter":         param,
		"fee_limit":         10_000_000,
		"call_value":        0,
	}
	bodyBytes, _ := json.Marshal(body)

	url := fmt.Sprintf("%s/wallet/triggersmartcontract", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var triggerResp struct {
		Transaction struct {
			RawDataHex string          `json:"raw_data_hex"`
			Full       json.RawMessage `json:"-"`
		} `json:"transaction"`
	}
	// Decode the full response to get both raw_data_hex and the full transaction
	var rawResp map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, "", err
	}

	txBytes, ok := rawResp["transaction"]
	if !ok {
		return nil, "", fmt.Errorf("no transaction field in trigger response")
	}

	if err := json.Unmarshal(txBytes, &triggerResp.Transaction); err != nil {
		return nil, "", err
	}

	return &TriggerResult{Transaction: txBytes}, triggerResp.Transaction.RawDataHex, nil
}

// BroadcastTransaction broadcasts a signed transaction.
// txJSON is the full transaction JSON; signature is the hex-encoded 65-byte signature.
func (c *Client) BroadcastTransaction(ctx context.Context, txJSON json.RawMessage, signature string) error {
	// Merge signature into transaction JSON
	var tx map[string]json.RawMessage
	if err := json.Unmarshal(txJSON, &tx); err != nil {
		return err
	}
	sigBytes, _ := json.Marshal([]string{signature})
	tx["signature"] = sigBytes

	body, _ := json.Marshal(tx)

	url := fmt.Sprintf("%s/wallet/broadcasttransaction", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Result  bool   `json:"result"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.Result {
		return fmt.Errorf("broadcast failed: %s", result.Message)
	}
	return nil
}

// USDTAmountFromSun converts TRC-20 USDT value (6-decimal sun units) to a display string.
func USDTAmountFromSun(sun string) (*big.Float, bool) {
	sunInt, ok := new(big.Int).SetString(sun, 10)
	if !ok {
		return nil, false
	}
	// USDT on Tron uses 6 decimal places
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
	result := new(big.Float).Quo(new(big.Float).SetInt(sunInt), divisor)
	return result, true
}

// USDTToSun converts a USDT amount string to sun (6 decimal places).
func USDTToSun(amount string) (*big.Int, error) {
	f, _, err := new(big.Float).Parse(amount, 10)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %s", amount)
	}
	// multiply by 10^6
	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
	sunFloat := new(big.Float).Mul(f, multiplier)
	sunInt, _ := sunFloat.Int(nil)
	return sunInt, nil
}

// HashRawData returns SHA256(rawDataBytes) for Tron transaction signing.
func HashRawData(rawDataHex string) ([]byte, error) {
	b, err := hex.DecodeString(rawDataHex)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(b)
	return h[:], nil
}

// encodeTRC20TransferParam ABI-encodes the arguments for transfer(address,uint256).
// Returns hex string (without function selector).
func encodeTRC20TransferParam(toAddr string, amount *big.Int) (string, error) {
	addrBytes, err := TronBase58ToBytes(toAddr)
	if err != nil {
		return "", fmt.Errorf("invalid to address: %w", err)
	}
	// Use 20-byte address (drop 0x41 prefix byte)
	addr20 := addrBytes[1:] // 20 bytes

	// Encode: 32-byte address param (left-pad with 12 zero bytes) + 32-byte uint256 (big-endian)
	param := make([]byte, 64)
	copy(param[12:32], addr20) // 12 zero padding + 20-byte address
	amountBytes := amount.Bytes()
	if len(amountBytes) > 32 {
		return "", fmt.Errorf("amount too large")
	}
	copy(param[32+(32-len(amountBytes)):], amountBytes) // right-align in 32 bytes
	return hex.EncodeToString(param), nil
}
