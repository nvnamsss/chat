package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nvnamsss/chat/src/configs"
	"github.com/nvnamsss/chat/src/dtos"
	"github.com/nvnamsss/chat/src/errors"
	"github.com/nvnamsss/chat/src/logger"
)

// LLMAdapter defines the interface for LLM service communication
type LLMAdapter interface {
	GenerateResponse(ctx context.Context, request *dtos.LLMRequest) (*dtos.LLMResponse, error)
}

// llmAdapter implements the LLMAdapter interface
type llmAdapter struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
}

// NewLLMAdapter creates a new LLMAdapter
func NewLLMAdapter(config configs.LLM) LLMAdapter {
	return &llmAdapter{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		model:   config.Model,
	}
}

// GenerateResponse sends a request to the LLM vendor service and returns the response
func (a *llmAdapter) GenerateResponse(ctx context.Context, request *dtos.LLMRequest) (*dtos.LLMResponse, error) {
	log := logger.Context(ctx)
	startTime := time.Now()

	// Set default model if not provided
	if request.Model == "" {
		request.Model = a.model
	}

	// Prepare request body
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternal, "Failed to marshal LLM request")
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/generate", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternal, "Failed to create LLM request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))

	log.Debugf("Sending request to LLM service: %s", url)

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrLLMService, "Failed to connect to LLM service")
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(errors.ErrLLMService, fmt.Sprintf("LLM service returned error: %d", resp.StatusCode))
	}

	// Parse response
	var llmResponse dtos.LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResponse); err != nil {
		return nil, errors.Wrap(err, errors.ErrInternal, "Failed to parse LLM response")
	}

	elapsed := time.Since(startTime)
	log.Infof("LLM request completed in %v with %d tokens used", elapsed, llmResponse.Usage.TotalTokens)

	return &llmResponse, nil
}

type nothingLLMAdapter struct {
}

func NewNothingLLMAdapter() LLMAdapter {
	return &nothingLLMAdapter{}
}
func (a *nothingLLMAdapter) GenerateResponse(ctx context.Context, request *dtos.LLMRequest) (*dtos.LLMResponse, error) {

	return &dtos.LLMResponse{
		Message: dtos.LLMMessage{
			Role:    "assistant",
			Content: "This is a mock response from the LLM service.",
		},
	}, nil
}
