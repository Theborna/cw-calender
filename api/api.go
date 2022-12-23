package api

import (
	"cw-cal/model"
	"fmt"
	"log"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const (
	UNREGISTERED = `user @%s __%d__ is not registered on this service
	please register using /register`
	REGISTERED = `user @%s __%d__ is already registered on this service
	use /deadline to check deadlines
	or /updateUrl to update your calendar url`
)

func errHandler(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

type CwBot struct {
	*tele.Bot

	users []model.User // TODO: use database
}

func NewBot(pref tele.Settings) *CwBot {
	b, err := tele.NewBot(pref)
	errHandler(err)
	Bot := &CwBot{
		Bot:   b,
		users: []model.User{},
	}
	Bot.handlers()
	return Bot
}

func (b *CwBot) handlers() {
	b.Handle("/deadline", b.deadlineHandler(), middleware.IgnoreVia(), b.registeredMW())
	b.Handle("/register", b.registerHandler(), middleware.IgnoreVia(), b.unregisteredMW())
}

func (b *CwBot) deadlineHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		return ctx.Send("lol")
	}
}

func (b *CwBot) registerHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		user := model.User{
			User: ctx.Sender(),
		}
		// TODO: use database
		b.users = append(b.users, user)
		return ctx.Send("register successful")
	}
}

func (b *CwBot) registeredMW() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(ctx tele.Context) error {
			if b.isRegistered(ctx.Sender()) {
				return next(ctx)
			}
			return ctx.Send(fmt.Sprintf(UNREGISTERED, ctx.Sender().Username, ctx.Sender().ID))
		}
	}
}

func (b *CwBot) unregisteredMW() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(ctx tele.Context) error {
			if b.isRegistered(ctx.Sender()) {
				return ctx.Send(fmt.Sprintf(REGISTERED, ctx.Sender().Username, ctx.Sender().ID))
			}
			return next(ctx)
		}
	}
}

func (b *CwBot) isRegistered(sender *tele.User) bool { // TODO: use database
	for _, user := range b.users {
		if user.ID == sender.ID {
			return true
		}
	}
	return false
}
