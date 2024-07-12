package services

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
)

// EdgeConfigOperation contians the the available edge config operations
type EdgeConfigOperation string

const (
	// EdgeConfigCreate is used to create a new key in the edge config
	EdgeConfigCreate EdgeConfigOperation = "create"
	// EdgeConfigUpdate is used to update a given key in the edge config
	EdgeConfigUpdate EdgeConfigOperation = "update"
)

// EdgeConfig is a struct that contains the configuration for the edge config
type EdgeConfig struct {
	ID         string
	ReadToken  string
	AdminToken string
}

// NewEdgeConfig is a function that is used to generate a edge config
func NewEdgeConfig(id, readToken, adminToken string) *EdgeConfig {
	return &EdgeConfig{
		ID:         id,
		ReadToken:  readToken,
		AdminToken: adminToken,
	}
}

// EdgeConfigItem contains an item in the edge config
type EdgeConfigItem struct {
	Operation EdgeConfigOperation `json:"operation"`
	Key       string              `json:"key"`
	Value     string              `json:"value"`
}

// Update is a function that is used to update or create an item in the edge config
func (e *EdgeConfig) Update(items []EdgeConfigItem) error {
	data, err := sonic.Marshal(struct {
		Item []EdgeConfigItem `json:"items"`
	}{
		Item: items,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://api.vercel.com/v1/edge-config/%s/items", e.ID), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+e.AdminToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post the update to the edge config : %d", resp.StatusCode)
	}

	return nil
}

// Read is a function that is used to read the items in the edge config with the key
func (e *EdgeConfig) Read(key string) (payload any, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://edge-config.vercel.com/%s/item/%s?token=%s", e.ID, key, e.ReadToken), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retreive the item : %d", resp.StatusCode)
	}

	err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
