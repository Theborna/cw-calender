package api

import (
	"cw-cal/model"
	"fmt"
	"log"
	"strings"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

const (
	UNREGISTERED = `user @%s (%d) is not registered on this service
	please register using /register`
	REGISTERED = `user @%s (%d) is already registered on this service
	use /deadline to check deadlines
	or /updateUrl to update your calendar url`
	INFO = `user @%s (%d)
	calendar link: %s`
)

var (
	// Universal markup builders.
	menu = &tele.ReplyMarkup{ResizeKeyboard: true}
	// Reply buttons.
	btnHelp     = menu.Text("ℹ Help")
	btnSettings = menu.Text("⚙ Settings")
)

func errHandler(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

type CwBot struct {
	*tele.Bot

	users  []model.User // TODO: use database
	logged map[int64]*model.User
}

func NewBot(pref tele.Settings) *CwBot {
	b, err := tele.NewBot(pref)
	errHandler(err)
	Bot := &CwBot{
		Bot:    b,
		users:  []model.User{},
		logged: make(map[int64]*model.User),
	}
	Bot.handlers()
	return Bot
}

func (b *CwBot) handlers() {
	b.Handle("/deadline", b.deadlineHandler(), middleware.IgnoreVia(), b.registeredMW())
	b.Handle("/register", b.registerHandler(), middleware.IgnoreVia(), b.unregisteredMW())
	b.Handle("/info", b.infoHandler(), middleware.IgnoreVia(), b.registeredMW())
	b.Handle("/addCal", b.addCalHandler(), middleware.IgnoreVia(), b.registeredMW())
	b.Handle("/login", b.loginHandler(), middleware.IgnoreVia(), b.unregisteredMW())
	// TODO: implement buttons
	// menu.Reply(
	// 	menu.Row(btnHelp, btnSettings),
	// )
	// b.Handle("/start", func(c tele.Context) error {
	// 	return c.Send("Hello!", menu)
	// })
	// // On reply button pressed (message)
	// b.Handle(&btnHelp, func(c tele.Context) error {
	// 	return c.Edit("Here is some help: ...")
	// })
}

func (b *CwBot) deadlineHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		// user := b.registeredUser(ctx.Sender())
		for _, cal := range b.logged[ctx.Sender().ID].Schedule {
			ctx.Send(cal.Data)
		}
		return ctx.Send("available deadlines listed above")
	}
}

func (b *CwBot) infoHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		// user := b.registeredUser(ctx.Sender())
		links := []string{}
		for _, cal := range b.logged[ctx.Sender().ID].Schedule {
			links = append(links, cal.Link())
		}
		return ctx.Send(fmt.Sprintf(INFO, b.logged[ctx.Sender().ID].Username, b.logged[ctx.Sender().ID].ID, strings.Join(links, "\n")))
	}
}

func (b *CwBot) registerHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		user := model.User{
			User:     ctx.Sender(),
			Schedule: []model.Calendar{},
		}
		// TODO: use database
		b.users = append(b.users, user)
		b.logged[ctx.Sender().ID] = &user
		return ctx.Send("register successful")
	}
}

func (b *CwBot) loginHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		return ctx.Send("na")
	}
}

func (b *CwBot) addCalHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		cal, err := model.NewCalendar(ctx.Message().Payload)
		if err != nil {
			return ctx.Send(err.Error())
		}
		// user := b.registeredUser(ctx.Sender())
		b.logged[ctx.Sender().ID].Schedule = append(b.logged[ctx.Sender().ID].Schedule, *cal)
		return ctx.Send("added successfully")
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

func (b *CwBot) isRegistered(sender *tele.User) bool {
	return b.logged[sender.ID] != nil
}
