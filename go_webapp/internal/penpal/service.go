package penpal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"

	"zachbot/internal/ollama"
)

type Service struct {
	cfg        Config
	client     ollama.Client
	logger     *logrus.Logger
	httpClient *http.Client
	seen       map[string]struct{}
}

func NewService(cfg Config, client ollama.Client, logger *logrus.Logger) *Service {
	return &Service{
		cfg:        cfg,
		client:     client,
		logger:     logger,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		seen:       make(map[string]struct{}),
	}
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	s.logger.Infof("Starting penpal poller against %s every %s", s.cfg.MailhogAPI, s.cfg.PollInterval)

	if err := s.pollOnce(ctx); err != nil {
		s.logger.WithError(err).Warn("initial poll failed")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.pollOnce(ctx); err != nil {
				s.logger.WithError(err).Warn("poll failed")
			}
		}
	}
}

func (s *Service) pollOnce(ctx context.Context) error {
	msgs, err := s.fetchMessages(ctx)
	if err != nil {
		return err
	}

	for _, m := range msgs {
		if _, processed := s.seen[m.ID]; processed {
			continue
		}
		if strings.EqualFold(m.From.Address, s.cfg.FromAddress) {
			s.seen[m.ID] = struct{}{}
			continue
		}
		if err := s.handleMessage(ctx, m); err != nil {
			s.logger.WithError(err).Warnf("failed to handle message %s", m.ID)
			continue
		}
		s.seen[m.ID] = struct{}{}
		_ = s.deleteMessage(ctx, m.ID)
	}
	return nil
}

func (s *Service) handleMessage(ctx context.Context, m MHMessage) error {
	body := m.Body
	if body == "" {
		body = "(empty message)"
	}

	prompt := fmt.Sprintf("You are an email pen pal. Reply concisely and warmly.\n\nSubject: %s\nFrom: %s\n\nMessage:\n%s", m.Subject, m.From, body)
	req := ollama.ChatRequest{
		Prompt:  prompt,
		Model:   s.cfg.DefaultModel,
		Persona: s.cfg.DefaultPersona,
	}

	resp, err := s.client.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("llm generate: %w", err)
	}

	if err := s.sendReply(m, resp.Response); err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"from":    m.From.Address,
		"subject": m.Subject,
	}).Info("replied to email")
	return nil
}

func (s *Service) sendReply(msg MHMessage, reply string) error {
	e := email.NewEmail()
	e.From = s.cfg.FromAddress
	e.To = []string{msg.From.Address}
	e.Subject = "Re: " + msg.Subject
	e.Text = []byte(strings.TrimSpace(reply))

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	if err := e.Send(addr, nil); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func (s *Service) fetchMessages(ctx context.Context) ([]MHMessage, error) {
	url := fmt.Sprintf("%s/api/v2/messages?limit=50", strings.TrimRight(s.cfg.MailhogAPI, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mailhog api status %d", resp.StatusCode)
	}

	var parsed MHResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	out := make([]MHMessage, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		m := item.Normalize()
		if m != nil {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (s *Service) deleteMessage(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/v1/messages/%s", strings.TrimRight(s.cfg.MailhogAPI, "/"), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Mailhog API types
type MHResponse struct {
	Total int         `json:"total"`
	Count int         `json:"count"`
	Items []MHRawItem `json:"items"`
}

type MHRawItem struct {
	ID      string      `json:"ID"`
	Content MHContent   `json:"Content"`
	MIME    *MHMIMEPart `json:"MIME"`
}

type MHContent struct {
	Headers map[string][]string `json:"Headers"`
	Body    string              `json:"Body"`
}

type MHMIMEPart struct {
	Parts   []MHMIMEPart        `json:"Parts"`
	Body    string              `json:"Body"`
	Headers map[string][]string `json:"Headers"`
}

type MHMessage struct {
	ID      string
	From    *mail.Address
	Subject string
	Body    string
}

func (r MHRawItem) Normalize() *MHMessage {
	from := firstHeader(r.Content.Headers, "From")
	toParse := strings.TrimSpace(from)
	addr, _ := mail.ParseAddress(toParse)

	subject := firstHeader(r.Content.Headers, "Subject")
	body := pickBody(r)

	if addr == nil {
		return nil
	}
	return &MHMessage{
		ID:      r.ID,
		From:    addr,
		Subject: subject,
		Body:    body,
	}
}

func pickBody(r MHRawItem) string {
	if r.MIME != nil {
		if body := findTextPlain(*r.MIME); body != "" {
			return body
		}
	}
	if strings.TrimSpace(r.Content.Body) != "" {
		return r.Content.Body
	}
	return ""
}

func findTextPlain(part MHMIMEPart) string {
	ct := strings.ToLower(firstHeader(part.Headers, "Content-Type"))
	if strings.Contains(ct, "text/plain") {
		return part.Body
	}
	for _, child := range part.Parts {
		if b := findTextPlain(child); b != "" {
			return b
		}
	}
	return ""
}

func firstHeader(headers map[string][]string, key string) string {
	for k, vals := range headers {
		if strings.EqualFold(k, key) && len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}
