package controllers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_chat"
)

type ChattingController struct {
	congo.BaseController
	Chat *congo_chat.CongoChat
	auth *congo_auth.Controller
}

func (chatting *ChattingController) Setup(app *congo.Application) {
	chatting.BaseController.Setup(app)
	chatting.auth = app.Use("auth").(*congo_auth.Controller)
	chatting.Chat = congo_chat.InitCongoChat(app.DB.Root, chatting.auth.CongoAuth)
	app.HandleFunc("GET /chatting/{user}", chatting.handleMessages)
	app.Handle("POST /chatting/messages", chatting.auth.ProtectFunc(chatting.sendMessage))
}

func (chatting ChattingController) Handle(req *http.Request) congo.Controller {
	chatting.Request = req
	return &chatting
}

func (chatting *ChattingController) Mailbox() (*congo_chat.Mailbox, error) {
	user, _ := chatting.auth.Authenticate("user", chatting.Request)
	mb, err := chatting.Chat.GetMailboxForOwner(user.ID)
	return mb, err
}

func (chatting *ChattingController) Contacts() ([]*congo_auth.Identity, error) {
	mb, err := chatting.Mailbox()
	log.Println("mailbox", mb, err)
	if err != nil {
		return nil, err
	}

	contacts, err := mb.Contacts()
	if err != nil {
		return nil, err
	}

	var identities []*congo_auth.Identity
	for _, contact := range contacts {
		i, err := chatting.auth.Lookup(contact.OwnerID)
		if err != nil {
			return nil, err
		}
		identities = append(identities, i)
	}

	return identities, nil
}

func (chatting *ChattingController) Messages() ([]*congo_chat.Message, error) {
	mb, err := chatting.Mailbox()
	if err != nil {
		return nil, err
	}
	senderID := chatting.PathValue("user")
	if senderID == "me" {
		user, _ := chatting.auth.Authenticate("user", chatting.Request)
		senderID = user.ID
	}
	return mb.Messages(senderID)
}

func (chatting ChattingController) handleMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		chatting.Render(w, r, "error-message", errors.New("streaming unsupported"))
		return
	}

	// user, _ := chatting.auth.Authenticate("user", r)
	userID := r.PathValue("user")
	if userID == "me" {
		user, _ := chatting.auth.Authenticate("user", r)
		userID = user.ID
	}

	feed, close := chatting.Chat.Listen(userID)
	defer close()

	log.Println("Listening for messages...", userID)
	for m := range feed {
		var buf bytes.Buffer
		chatting.Render(&buf, r, "chat-message", m)
		content := strings.ReplaceAll(buf.String(), "\n", "")
		_, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", content)
		if err != nil {
			log.Println("Failed to write message: ", err)
			return
		}
		flusher.Flush()
	}
}

func (chatting ChattingController) sendMessage(w http.ResponseWriter, r *http.Request) {
	message := r.FormValue("message")
	if message == "" {
		chatting.Render(w, r, "error-message", errors.New("missing message"))
		return
	}

	user, _ := chatting.auth.Authenticate("user", r)
	if user == nil {
		chatting.Render(w, r, "error-message", errors.New("unauthorized"))
		return
	}

	mb, err := chatting.Chat.GetMailbox(user.ID)
	if err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	mailbox := r.FormValue("mailbox")
	if mailbox == "me" {
		mailbox = user.ID
	}
	log.Println("Sending message", r.FormValue("mailbox"), user.ID, message)
	if _, err := mb.Send(mailbox, message); err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
