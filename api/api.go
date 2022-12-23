package api

import (
	"log"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func errHandler(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

type CwBot struct {
	*tele.Bot

	name string
}

func NewBot(pref tele.Settings) *CwBot {
	b, err := tele.NewBot(pref)
	errHandler(err)
	Bot := &CwBot{
		Bot:  b,
		name: "hiii",
	}
	Bot.handlers()
	return Bot
}

func (b *CwBot) handlers() {
	b.Handle("/deadline", b.deadlineHandler(), middleware.IgnoreVia(), b.registeredMW())
}

func (b *CwBot) deadlineHandler() tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("lol")
	}
}

func (b *CwBot) registeredMW() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Callback() != nil {
				defer c.Respond()
			}
			return next(c)
		}
	}
}
