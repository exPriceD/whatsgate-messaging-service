package domain

import "context"

type Client interface {
	SendTextMessage(ctx context.Context, phoneNumber, text string, async bool) (*SendMessageResponse, error)
	SendMediaMessage(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (*SendMessageResponse, error)
}

type SendMessageRequest struct {
	WhatsappID string    `json:"WhatsappID"`
	Async      bool      `json:"async"`
	Recipient  Recipient `json:"recipient"`
	Message    Message   `json:"message"`
}

type Recipient struct {
	ID     string `json:"id,omitempty"`
	Type   string `json:"type,omitempty"`
	Number string `json:"number,omitempty"`
}

type Message struct {
	Type  string `json:"type,omitempty"`
	Body  string `json:"body,omitempty"`
	Quote string `json:"quote,omitempty"`
	Media *Media `json:"media,omitempty"`
}

type Media struct {
	MimeType string `json:"mimetype"`
	Data     string `json:"data"`
	Filename string `json:"filename"`
}

type SendMessageResponse struct {
	ID      string `json:"id,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}
