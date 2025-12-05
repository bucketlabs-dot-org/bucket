package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	//"path/filepath"
	"strings"

	"github.com/gagehenrich/bucket/cli/internal/config"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.APIBase,
		apiKey:  cfg.APIKey,
		http:    &http.Client{},
	}
}

func (c *Client) attachAuth(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
}

type UploadInitResponse struct {
	FileID    string `json:"file_id"`
	UploadURL string `json:"upload_url"`
	TinyCode  string `json:"tiny_code"`
	Secret    string `json:"secret"`
	ExpiresAt string `json:"expires_at"`
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

type DownloadAuthResponse struct {
	DownloadURL string `json:"download_url"`
	Filename	string `json:"filename"`
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

type FileInfo struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	TinyCode  string `json:"tiny_code"`
	ExpiresAt string `json:"expires_at"`
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
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed: %s", b)
	}

	var files []FileInfo
	json.NewDecoder(resp.Body).Decode(&files)
	return files, nil
}

func ExtractTinyCode(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func (c *Client) LoginOrCreate(email, password string) (string, error) {
	payload := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)

	req, _ := http.NewRequest("POST", c.baseURL+"/v1/account/login", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("login/create failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login/create failed: %s", string(b))
	}

	req2, _ := http.NewRequest("POST", c.baseURL+"/v1/account/keys", bytes.NewBuffer([]byte(payload)))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := c.http.Do(req2)
	if err != nil {
		return "", fmt.Errorf("api key creation failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp2.Body)
		return "", fmt.Errorf("api key creation failed: %s", string(b))
	}

	var out struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode api key failed: %w", err)
	}
	if out.APIKey == "" {
		return "", fmt.Errorf("empty api key returned")
	}
	return out.APIKey, nil
}
