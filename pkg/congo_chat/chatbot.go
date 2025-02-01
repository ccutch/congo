package congo_chat

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Chatbot struct {
	chat *CongoChat
	congo.Model
	Name   string
	Prompt string
}

type ChatbotMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Images  []any  `json:"images"`
}

func (chat *CongoChat) NewChatbot(name, prompt string) (*Chatbot, error) {
	c := Chatbot{chat, chat.db.NewModel(uuid.NewString()), name, prompt}
	return &c, chat.db.Query(`

		INSERT INTO chatbots (id, name, prompt)
		VALUES (?, ?, ?)
		RETURNING created_at, updated_at

	`, c.ID, c.Name, c.Prompt).Scan(&c.CreatedAt, &c.UpdatedAt)
}

func (chat *CongoChat) GetChatbot(id string) (*Chatbot, error) {
	c := Chatbot{Model: chat.db.Model(), chat: chat}
	return &c, chat.db.Query(`

		SELECT id, name, prompt, created_at, updated_at
		FROM chatbots
		WHERE id = ?

	`, id).Scan(&c.ID, &c.Name, &c.Prompt, &c.CreatedAt, &c.UpdatedAt)
}

func (chat *CongoChat) AllChatbots() ([]*Chatbot, error) {
	var chatbots []*Chatbot
	return chatbots, chat.db.Query(`

		SELECT id, name, prompt, created_at, updated_at
		FROM chatbots

	`).All(func(scan congo.Scanner) error {
		c := Chatbot{Model: chat.db.Model(), chat: chat}
		chatbots = append(chatbots, &c)
		return scan(&c.ID, &c.Name, &c.Prompt, &c.CreatedAt, &c.UpdatedAt)
	})
}

func (chatbot *Chatbot) Generate(prompt string) (string, error) {
	var body bytes.Buffer
	json.NewEncoder(&body).Encode(map[string]any{
		"model":  chatbot.chat.Model,
		"prompt": prompt,
		"stream": false,
	})

	log.Println("calling /api/generate")
	resp, err := http.Post(chatbot.chat.host+"/api/generate", "application/json", &body)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to send message: %s", resp.Status)
	}

	var response struct {
		Model   string `json:"model"`
		Created string `json:"created_at"`
		Res     string `json:"response"`
		Done    bool   `json:"done"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil || response.Res == "" {
		err = cmp.Or(err, fmt.Errorf("no message found: %s", resp.Status))
		return "", fmt.Errorf("failed to decode response: %s", err)
	}

	return response.Res, nil
}

func (chatbot *Chatbot) Chat(content string, history []ChatbotMessage) (*Message, error) {
	var body bytes.Buffer
	json.NewEncoder(&body).Encode(map[string]any{
		"model":    chatbot.chat.Model,
		"messages": history,
		"stream":   false,
	})

	log.Println("calling /api/chat")
	resp, err := http.Post(chatbot.chat.host+"/api/chat", "application/json", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to send message: %s/api/chat %s %s", chatbot.chat.host, chatbot.chat.Model, resp.Status)
	}

	var response struct {
		Model   string   `json:"model"`
		Created string   `json:"created_at"`
		Message *Message `json:"message"`
		Done    bool     `json:"done"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil || response.Message == nil {
		err = cmp.Or(err, fmt.Errorf("no message found: %s", resp.Status))
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	i := strings.LastIndex(response.Message.Content, "</think>")
	response.Message.Content = response.Message.Content[i+len("</think>"):]
	response.Message.Content = strings.TrimSpace(response.Message.Content)
	return response.Message, nil
}

func (chatbot *Chatbot) Mailbox() (*Mailbox, error) {
	log.Println("Loading old mailbox")
	mailbox, err := chatbot.chat.GetMailboxForOwner(chatbot.ID)
	if err != nil {
		log.Println("Creating new mailbox")
		mailbox, err = chatbot.chat.NewMailboxWithID(chatbot.ID, chatbot.ID, chatbot.Name, 10000)
	}
	return mailbox, err
}

func (chat *CongoChat) Register(chatbot *Chatbot) error {
	mailbox, err := chatbot.Mailbox()
	if err != nil {
		return errors.Wrap(err, "failed to create or load mailbox")
	}

	l, close := chat.Listen(chatbot.ID)
	defer close()

	log.Println("Listening for messages")
	for m := range l.Messages {
		log.Println("Received message", m.FromID, m.ToID)
		if m.FromID == mailbox.ID {
			continue
		}

		history, err := mailbox.Messages(m.FromID)
		if err != nil {
			log.Println("mailbox not found")
			return err
		}

		log.Println("Sending message to agent")
		res, err := chatbot.Chat(m.Content, chatbot.formatHistory(history))
		if err != nil {
			log.Println("Failed to chat with agent")
			return err
		}

		log.Println("Sending message back to user")
		if _, err := mailbox.Send(m.FromID, res.Content); err != nil {
			return err
		}
	}

	return nil
}

func (chatbot *Chatbot) formatHistory(history []*Message) []ChatbotMessage {
	mb, err := chatbot.Mailbox()
	if err != nil {
		return nil
	}
	var formatted []ChatbotMessage
	for _, m := range history {
		role := "user"
		if mb.ID == m.FromID {
			role = "assistant"
		}
		formatted = append(formatted, ChatbotMessage{
			Role:    role,
			Content: m.Content,
		})
	}
	slices.Reverse(formatted)
	return formatted
}
