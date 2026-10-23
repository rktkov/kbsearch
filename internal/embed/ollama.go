// Package embed — клиент локальной модели эмбеддингов через Ollama.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client ходит в локальный Ollama. Никаких внешних API.
type Client struct {
	url   string
	model string
	dims  int
	http  *http.Client
}

// New создаёт клиент. Таймаут щедрый: первый запрос грузит модель в память.
func New(url, model string, dims int) *Client {
	return &Client{
		url:   url,
		model: model,
		dims:  dims,
		http:  &http.Client{Timeout: 180 * time.Second},
	}
}

// Model возвращает имя модели — оно пишется в метаданные индекса.
func (c *Client) Model() string { return c.model }

// Dims возвращает размерность вектора.
func (c *Client) Dims() int { return c.dims }

// BatchSize — сколько текстов отправляем за один запрос.
const BatchSize = 16

// EmbedPassages эмбеддит документы батчами.
func (c *Client) EmbedPassages(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, 0, len(texts))
	for start := 0; start < len(texts); start += BatchSize {
		end := start + BatchSize
		if end > len(texts) {
			end = len(texts)
		}
		vecs, err := c.embed(ctx, texts[start:end])
		if err != nil {
			return nil, fmt.Errorf("embed batch %d-%d: %w", start, end, err)
		}
		out = append(out, vecs...)
	}
	return out, nil
}

// EmbedQuery эмбеддит один запрос.
func (c *Client) EmbedQuery(ctx context.Context, q string) ([]float32, error) {
	vecs, err := c.embed(ctx, []string{q})
	if err != nil {
		return nil, err
	}
	if len(vecs) != 1 {
		return nil, fmt.Errorf("ожидали 1 вектор, получили %d", len(vecs))
	}
	return vecs[0], nil
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (c *Client) embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(embedRequest{Model: c.model, Input: texts})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama недоступна: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama ответила %s", resp.Status)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embeddings) != len(texts) {
		return nil, fmt.Errorf("векторов %d, текстов %d", len(out.Embeddings), len(texts))
	}
	for i, v := range out.Embeddings {
		if len(v) != c.dims {
			return nil, fmt.Errorf("вектор %d: размерность %d, ожидали %d", i, len(v), c.dims)
		}
	}
	return out.Embeddings, nil
}
