package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/ccutch/congo/pkg/congo_chat"
)

type ChattingController struct {
	congo.BaseController
	Chat *congo_chat.CongoChat
	auth *congo_auth.AuthController
}

func (chatting *ChattingController) Setup(app *congo.Application) {
	chatting.BaseController.Setup(app)
	chatting.auth = app.Use("auth").(*congo_auth.AuthController)
	chatting.Chat = congo_chat.InitCongoChat(app.DB.Root, chatting.auth.CongoAuth)
	app.Handle("GET /chatting/{user}", chatting.auth.ProtectFunc(chatting.handleMessages))
	app.Handle("POST /chatting/messages", chatting.auth.ProtectFunc(chatting.sendMessage))
	app.Handle("GET /chatting/invite", chatting.auth.Protect(app.Serve("url-copied-toast")))
}

func (chatting ChattingController) Handle(req *http.Request) congo.Controller {
	chatting.Request = req
	return &chatting
}

func (chatting *ChattingController) Mailbox() (*congo_chat.Mailbox, error) {
	user, _ := chatting.auth.Authenticate(chatting.Request, "user")
	return chatting.Chat.GetMailboxForOwner(user.ID)
}

func (chatting *ChattingController) Contacts() (ids []*congo_auth.Identity, err error) {
	users, err := chatting.auth.Search("")
	if err != nil {
		return nil, err
	}

	i, _ := chatting.auth.Authenticate(chatting.Request, "user")
	for _, user := range users["user"] {
		if user.ID == i.ID {
			continue
		}
		ids = append(ids, user)
	}

	return ids, nil
}

func (chatting *ChattingController) Messages() ([]*congo_chat.Message, error) {
	mb, err := chatting.Mailbox()
	if err != nil {
		return nil, err
	}

	senderID := chatting.PathValue("user")
	if senderID == "me" {
		user, _ := chatting.auth.Authenticate(chatting.Request, "user")
		senderID = user.ID
	}

	return mb.Messages(senderID)
}

func (chatting ChattingController) handleMessages(w http.ResponseWriter, r *http.Request) {
	flush, err := chatting.EventStream(w, r)
	if err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	userID := r.PathValue("user")
	if userID == "me" {
		user, _ := chatting.auth.Authenticate(r, "user")
		userID = user.ID
	}

	feed, close := chatting.Chat.Listen(userID)
	defer close()

	for {
		select {
		case <-r.Context().Done():
			return
		case m := <-feed.Messages:
			flush("chat-message", m)
		}
	}
}

func (chatting ChattingController) sendMessage(w http.ResponseWriter, r *http.Request) {
	message := r.FormValue("message")
	if message == "" {
		chatting.Render(w, r, "error-message", errors.New("missing message"))
		return
	}

	user, _ := chatting.auth.Authenticate(r, "user")
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
