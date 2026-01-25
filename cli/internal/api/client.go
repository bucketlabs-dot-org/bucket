package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bucketlabs-dot-org/bucket/cli/internal/config"
)

const (
	keyPrefix = "bk-"
	keySuffix = "-0205"
)

// API structures
type AccountInfoResponse struct {
	Tier      string `json:"tier"`
	UsedBytes int64  `json:"used_bytes"`
	Quota     int64  `json:"quota"`
}

type Client struct {
	baseURL    string
	apiKey     string
	deviceID   string
	deviceName string
	http       *http.Client
}

type DeleteResponse struct {
	Response string `json:"delete_response"`
}

type DownloadAuthResponse struct {
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
}

type FileInfo struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	TinyCode  string `json:"tiny_code"`
	ExpiresAt string `json:"expires_at"`
	SecretKey string `json:"download_secret_hash"`
}

type UploadInitResponse struct {
	FileID    string `json:"file_id"`
	UploadURL string `json:"upload_url"`
	TinyCode  string `json:"tiny_code"`
	Secret    string `json:"secret"`
	ExpiresAt string `json:"expires_at"`
}

type TwoFARequiredError struct{}

// API functions
func New(cfg *config.Config) *Client {
	return &Client{
		baseURL:    cfg.APIBase,
		apiKey:     cfg.APIKey,
		deviceID:   cfg.DeviceID,
		deviceName: cfg.DeviceName,
		http: &http.Client{
			Timeout: 0,
			Transport: &http.Transport{
				ExpectContinueTimeout: 10 * time.Minute,
				ResponseHeaderTimeout: 10 * time.Minute,
			},
		},
	}
}

func (e *TwoFARequiredError) Error() string {
	return "2fa_required"
}

// formatAPIKey wraps the raw API key with prefix and suffix
// Config stores: 8db56714-1229-41be-a938-2f536b75de94
// Wire format:   bk-8db56714-1229-41be-a938-2f536b75de94-0205
func formatAPIKey(rawKey string) string {
	if rawKey == "" {
		return ""
	}
	// Avoid double-formatting if already formatted
	if strings.HasPrefix(rawKey, keyPrefix) && strings.HasSuffix(rawKey, keySuffix) {
		return rawKey
	}
	return keyPrefix + rawKey + keySuffix
}

// attachAuth adds authentication headers to requests
// SECURITY: Sends formatted API key + device ID for verification
func (c *Client) attachAuth(req *http.Request) {
	if c.apiKey != "" {
		formattedKey := formatAPIKey(c.apiKey)
		req.Header.Set("Authorization", "Bearer "+formattedKey)
	}
	// CRITICAL: Send device_id with every authenticated request
	if c.deviceID != "" {
		req.Header.Set("X-Device-ID", c.deviceID)
	}
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) RequestUpload(filename string, size int64) (*UploadInitResponse, error) {
	payload := fmt.Sprintf(`{"filename":"%s","size_bytes":%d}`, filename, size)

	req, _ := http.NewRequest("POST", c.baseURL+"/v1/upload/request", bytes.NewBuffer([]byte(payload)))
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload request failed: %s", b)
	}

	var out UploadInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UploadFile(url, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()

	req, _ := http.NewRequest("PUT", url, f)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", b)
	}

	return nil
}

func (c *Client) VerifyUpload(fileID string) error {
	payload := fmt.Sprintf(`{"file_id":"%s"}`, fileID)

	req, _ := http.NewRequest("POST", c.baseURL+"/v1/upload/verify", bytes.NewBuffer([]byte(payload)))
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload verification failed: %s", b)
	}

	return nil
}

func (c *Client) CleanupFailedUpload(fileID string) error {
	payload := fmt.Sprintf(`{"file_id":"%s"}`, fileID)

	req, _ := http.NewRequest("POST", c.baseURL+"/v1/upload/cleanup", bytes.NewBuffer([]byte(payload)))
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cleanup failed: %s", b)
	}

	return nil
}

func (c *Client) AuthDownload(tiny, secret string) (string, string, error) {
	body := fmt.Sprintf(`{"tiny":"%s","secret":"%s"}`, tiny, secret)

	req, _ := http.NewRequest("POST", c.baseURL+"/v1/download/auth", bytes.NewBuffer([]byte(body)))
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("auth failed: %s", b)
	}

	var out struct {
		DownloadURL string `json:"download_url"`
		Filename    string `json:"filename"`
	}
	json.NewDecoder(resp.Body).Decode(&out)
	return out.DownloadURL, out.Filename, nil
}

func (c *Client) DownloadFile(url string, suggestedFilename string) (string, error) {
	resp, err := c.http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("download failed: %s", b)
	}

	// Use suggested filename from server
	filename := suggestedFilename
	if filename == "" {
		filename = "downloaded.file"
	}

	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return filename, err
}

func (c *Client) DeleteFile(tiny string) error {
	payload := fmt.Sprintf(`{"tiny":"%s"}`, tiny)

	req, _ := http.NewRequest("POST", c.baseURL+"/v1/delete", bytes.NewBuffer([]byte(payload)))
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("object delete failed: %s", b)
	}

	return nil
}

func (c *Client) ListFiles() ([]FileInfo, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/v1/files", nil)
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			return nil, fmt.Errorf("Invalid subscription type. Visit https://bucketlabs.org/auth to upgrade.")
		} else {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API response: %s", b)
		}
	}

	var files []FileInfo
	json.NewDecoder(resp.Body).Decode(&files)
	return files, nil
}

func ExtractTinyCode(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func (c *Client) Login(email, password, otpCode string) (string, error) {
	// -------------------------
	// LOGIN (password check)
	// -------------------------
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}

	body1, _ := json.Marshal(loginPayload)

	req1, _ := http.NewRequest(
		"POST",
		c.baseURL+"/v1/account/login",
		bytes.NewBuffer(body1),
	)
	req1.Header.Set("Content-Type", "application/json")

	resp1, err := c.http.Do(req1)
	if err != nil {
		return "", err
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp1.Body)
		return "", fmt.Errorf("login failed: %s", b)
	}

	// -------------------------
	// API KEY CREATION
	// -------------------------
	keyPayload := map[string]string{
		"email":     email,
		"password":  password,
		"device_id": c.deviceID,   // REQUIRED
		"name":      c.deviceName, // REQUIRED
	}

	if otpCode != "" {
		keyPayload["code"] = otpCode
	}

	body2, _ := json.Marshal(keyPayload)

	req2, _ := http.NewRequest(
		"POST",
		c.baseURL+"/v1/account/keys",
		bytes.NewBuffer(body2),
	)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := c.http.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusPaymentRequired {
		return "", &TwoFARequiredError{}
	}

	if resp2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp2.Body)
		return "", fmt.Errorf("api key creation failed: %s", b)
	}

	var out struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&out); err != nil {
		return "", err
	}

	return out.APIKey, nil
}

func (c *Client) FetchAccountInfo() (*AccountInfoResponse, error) {
	req, _ := http.NewRequest("GET", c.baseURL+"/v1/account/info", nil)
	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account info failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("account info failed: %s", b)
	}

	var out AccountInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &out, nil
}

func (c *Client) Logout() error {
	req, err := http.NewRequest(
		"POST",
		c.baseURL+"/v1/account/logout",
		nil,
	)
	if err != nil {
		return err
	}

	c.attachAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Treat non-200 as warning, not fatal
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout failed: %s", b)
	}

	return nil
}