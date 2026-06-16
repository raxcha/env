package apiclient

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"env/filesystem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	addr               string
	token              string
	accessClientID     string
	accessClientSecret string
	http               *http.Client
	mu                 sync.RWMutex
	baseHashes         map[string]string
	conflictHashes     map[string]string
}

func New(addr, token string) *Client {
	addr = strings.TrimRight(strings.TrimSpace(addr), "/")
	return &Client{
		addr:               addr,
		token:              token,
		accessClientID:     firstEnv("ENV_CF_ACCESS_CLIENT_ID", "CF_ACCESS_CLIENT_ID"),
		accessClientSecret: firstEnv("ENV_CF_ACCESS_CLIENT_SECRET", "CF_ACCESS_CLIENT_SECRET"),
		http:               &http.Client{},
		baseHashes:         map[string]string{},
		conflictHashes:     map[string]string{},
	}
}

type FileEntry struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	BaseHash string `json:"base_hash,omitempty"`
}

type WriteRequest struct {
	Branch string      `json:"branch"`
	Filter string      `json:"filter,omitempty"`
	Files  []FileEntry `json:"files"`
}

type ConflictInfo struct {
	Path            string `json:"path"`
	ConflictContent string `json:"conflict_content"`
}

type WriteResponse struct {
	Written   []string       `json:"written"`
	Conflicts []ConflictInfo `json:"conflicts"`
}

type ReadRequest struct {
	Branch string `json:"branch"`
	Filter string `json:"filter,omitempty"`
}

type FileInfo struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type ReadResponse struct {
	Files     []FileInfo `json:"files"`
	Conflicts []string   `json:"conflicts"`
}

func HashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h)
}

// wirePage mirrors api.Page for JSON encoding/decoding without importing the api module.
type wirePage struct {
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	Type     string         `json:"type"`
	Stage    string         `json:"stage"`
	Sorting  string         `json:"sorting"`
	Content  []string       `json:"content"`
	Metadata map[string]any `json:"metadata"`
	Children []*wirePage    `json:"children,omitempty"`
}

func (c *Client) Fetch(path string, depth int) (*filesystem.Page, error) {

	u := c.addr + "/page?path=" + url.QueryEscape(path) + "&depth=" + strconv.Itoa(depth)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	wp := &wirePage{}
	if err := json.Unmarshal(body, wp); err != nil {
		return nil, err
	}
	return fromWire(wp), nil
}

func (c *Client) Push(p *filesystem.Page) error {

	body, err := json.Marshal(toWire(p))
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, c.addr+"/page", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("push: %s", resp.Status)
	}
	return nil
}

func (c *Client) Delete(path string) error {

	u := c.addr + "/page?path=" + url.QueryEscape(path)
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete: %s", resp.Status)
	}
	return nil
}

func (c *Client) URL() string {
	return c.addr
}

func (c *Client) setAuth(req *http.Request) {
	if c.accessClientID != "" && c.accessClientSecret != "" {
		req.Header.Set("CF-Access-Client-Id", c.accessClientID)
		req.Header.Set("CF-Access-Client-Secret", c.accessClientSecret)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func toWire(p *filesystem.Page) *wirePage {
	wp := &wirePage{
		Name:     p.Name,
		Path:     p.Path,
		Type:     p.Type,
		Stage:    p.Stage,
		Sorting:  p.Sorting,
		Content:  p.Content,
		Metadata: p.Metadata,
	}
	for _, child := range p.Children {
		wp.Children = append(wp.Children, toWire(child))
	}
	return wp
}

func (c *Client) BaseHash(path string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baseHashes[path]
}

func (c *Client) SetBaseHash(path, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseHashes[path] = hash
}

func (c *Client) ConflictHash(path string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conflictHashes[path]
}

func (c *Client) SetConflictHash(path, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conflictHashes[path] = hash
}

func (c *Client) ClearConflictHash(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.conflictHashes, path)
}

func (c *Client) Write(req WriteRequest) (*WriteResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(http.MethodPost, c.addr+"/sync", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	c.setAuth(r)

	resp, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("write: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wr WriteResponse
	if err := json.Unmarshal(b, &wr); err != nil {
		return nil, err
	}
	return &wr, nil
}

func (c *Client) Read(req ReadRequest) (*ReadResponse, error) {
	u := c.addr + "/sync?branch=" + url.QueryEscape(req.Branch)
	if req.Filter != "" {
		u += "&filter=" + url.QueryEscape(req.Filter)
	}

	r, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(r)

	resp, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("read: %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rr ReadResponse
	if err := json.Unmarshal(b, &rr); err != nil {
		return nil, err
	}
	return &rr, nil
}

func fromWire(wp *wirePage) *filesystem.Page {
	p := &filesystem.Page{
		Name:     wp.Name,
		Path:     wp.Path,
		Type:     wp.Type,
		Stage:    wp.Stage,
		Sorting:  wp.Sorting,
		Content:  wp.Content,
		Metadata: wp.Metadata,
		Options:  map[string]any{},
		Children: []*filesystem.Page{},
		Og:       &filesystem.Page{},
		Diff:     []string{},
	}
	if p.Content == nil {
		p.Content = []string{}
	}
	if p.Metadata == nil {
		p.Metadata = map[string]any{}
	}

	p.Metadata = filesystem.ParseMetadataFromContent(p.Content, p.Name)

	for _, child := range wp.Children {
		p.Children = append(p.Children, fromWire(child))
	}
	return p
}
