package controllers

import (
	"errors"
	"fmt"
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
	chatting.Chat = congo_chat.InitCongoChat(app.DB.Root,
		congo_chat.WithAuth(chatting.auth.CongoAuth),
		congo_chat.WithModel("deepseek-r1:1.5b"))

	http.Handle("GET /chatting/{user}", chatting.auth.ProtectFunc(chatting.handleMessages, "user"))
	http.Handle("POST /chatting/messages", chatting.auth.ProtectFunc(chatting.sendMessage, "user"))
	http.Handle("POST /chatting/new-agent", chatting.auth.ProtectFunc(chatting.newChatbot, "user"))

	for _, chatbot := range chatting.Agents() {
		log.Println("Starting agent", chatbot.Name)
		go func(chatbot *congo_chat.Chatbot) {
			err := chatting.Chat.Register(chatbot)
			log.Println("Failed to listen for messages", err)
		}(chatbot)
	}
}

func (chatting ChattingController) Handle(req *http.Request) congo.Controller {
	chatting.Request = req
	return &chatting
}

func (chatting *ChattingController) Mailbox() (*congo_chat.Mailbox, error) {
	user, _ := chatting.auth.Authenticate(chatting.Request, "user")
	return chatting.Chat.GetMailboxForOwner(user.ID)
}

func (chatting *ChattingController) Agents() []*congo_chat.Chatbot {
	agents, _ := chatting.Chat.AllChatbots()
	return agents
}

func (chatting *ChattingController) Mailboxes() (res []*congo_chat.Mailbox, err error) {
	mbs, err := chatting.Chat.AllMailboxes()
	if err != nil {
		return nil, err
	}
	i, _ := chatting.auth.Authenticate(chatting.Request, "user")
	for _, mb := range mbs {
		if mb.OwnerID != i.ID {
			res = append(res, mb)
		}
	}
	return res, err
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

	user, _ := chatting.auth.Authenticate(r, "user")
	userID := r.PathValue("user")
	if userID == "me" {
		userID = user.ID
	}

	mb, err := chatting.Chat.GetMailboxForOwner(userID)
	if err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	feed, close := chatting.Chat.Listen(mb.ID)
	defer close()

	for {
		select {
		case <-r.Context().Done():
			return
		case m := <-feed.Messages:
			log.Println("testing", userID, user.ID, m.FromID, m.ToID)
			// if userID == user.ID && (m.FromID == userID && m.ToID == userID) {
			// 	log.Println("Not sending message", userID, user.ID, m.FromID, m.ToID)
			// 	continue
			// }
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

	fromMailbox, err := chatting.Chat.GetMailbox(user.ID)
	if err != nil {
		fromMailbox, _ = chatting.Chat.NewMailboxWithID(user.ID, user.ID, user.Name)
	}

	toMailbox := r.FormValue("mailbox")
	if toMailbox == "me" {
		toMailbox = user.ID
	}

	if _, err := fromMailbox.Send(toMailbox, message); err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (chatting *ChattingController) newChatbot(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		chatting.Render(w, r, "error-message", errors.New("missing name"))
		return
	}

	prompt := r.FormValue("prompt")
	if prompt == "" {
		chatting.Render(w, r, "error-message", errors.New("missing prompt"))
		return
	}

	chatbot, err := chatting.Chat.NewChatbot(name, prompt)
	if err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	go func() {
		err := chatting.Chat.Register(chatbot)
		log.Println("Failed to listen for messages", err)
	}()

	mb, err := chatbot.Mailbox()
	if err != nil {
		chatting.Render(w, r, "error-message", err)
		return
	}

	chatting.Redirect(w, r, fmt.Sprintf("/%s", mb.ID))
}
