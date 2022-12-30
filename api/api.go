package api

import (
	"cw-cal/database"
	"cw-cal/model"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

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
	calendar link: 
	%s`
	WELCOME_TEXT = `Hello!
	I can help you see your deadlines from telegram!
	use /register to register to the service
	use /addCal to add your deadlines calendar link
	use /deadline to see your current deadlines
	`
	REPORT_TEXT = `Hi...
	Sorry to hear that you are experiencing issues with our service
	please report your problem using:
	/report <problem>
	and state your problem after /report
	example:
	'/report the bot does not update my deadlines!'
	`
)

var (
	// Universal markup builders.
	menu = &tele.ReplyMarkup{ResizeKeyboard: true}
	// Reply buttons.
	btnDeadline = menu.Text("⚠️ Deadlines")
	btnInfo     = menu.Text("ℹ Info")
	btnHelp     = menu.Text("ℹ Help")
	btnReport   = menu.Text("💢 Report")
)

func errHandler(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

type CwBot struct {
	*tele.Bot

	users    []model.User // TODO: use database
	database *database.CwDatabase
	logged   map[int64]*model.User
}

func NewBot(pref tele.Settings) *CwBot {
	b, err := tele.NewBot(pref)
	errHandler(err)
	db, _ := database.NewCwDatabase()
	// errHandler(err)
	Bot := &CwBot{
		Bot:      b,
		users:    []model.User{},
		database: db,
		logged:   make(map[int64]*model.User),
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
	b.Handle("/report", b.reportHandler(), middleware.IgnoreVia(), b.registeredMW())

	b.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("hi")
	})
	menu.Reply(
		menu.Row(btnDeadline, btnInfo, btnHelp, btnReport),
	)
	b.Handle("/start", func(c tele.Context) error {
		return c.Send(WELCOME_TEXT, menu)
	})
	b.Handle(&btnDeadline, b.deadlineHandler(), b.registeredMW())
	b.Handle(&btnInfo, b.infoHandler(), b.registeredMW())
	b.Handle(&btnHelp, func(c tele.Context) error {
		return c.Send(WELCOME_TEXT, menu)
	})
	b.Handle(&btnReport, func(c tele.Context) error {
		return c.Send(REPORT_TEXT)
	})
}

func (b *CwBot) deadlineHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		// user := b.registeredUser(ctx.Sender())
		for _, cal := range b.logged[ctx.Sender().ID].Schedule {
			cal.UpdateEvents()
			text := []string{}
			for _, e := range cal.Events {
				text = append(text, e.ToString())
			}
			ctx.Send(strings.Join(text, `====================================`))
		}
		return ctx.Send("available deadlines listed above")
	}
}

func (b *CwBot) infoHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		links := []string{}
		for _, cal := range b.logged[ctx.Sender().ID].Schedule {
			links = append(links, cal.Link)
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
		if err := b.database.AddUser(user); err != nil {
			return ctx.Send(err.Error())
		}
		b.logged[ctx.Sender().ID] = &user
		return ctx.Send("register successful")
	}
}

func (b *CwBot) loginHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		user, err := b.database.GetUser(ctx.Sender())
		if user != nil {
			b.logged[ctx.Sender().ID] = user
			return ctx.Send("logged successfully")
		}
		return ctx.Send(err.Error())
	}
}

func (b *CwBot) addCalHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		cal, err := model.NewCalendar(ctx.Message().Payload)
		if err != nil {
			return ctx.Send(err.Error())
		}
		if b.logged[ctx.Sender().ID].AddCall(cal) {
			if err := b.database.AddCal(cal, ctx.Sender().ID); err != nil {
				return ctx.Send(err.Error())
			}
			return ctx.Send("added successfully")
		}
		return ctx.Send("calendar already exists")
	}
}

func (b *CwBot) reportHandler() tele.HandlerFunc {
	return func(ctx tele.Context) error {
		sender, message := ctx.Sender().Username, ctx.Message().Payload
		text := fmt.Sprintf("sender: t.me/%s, message: %s", sender, message)
		if len(message) > 0 {
			err := ioutil.WriteFile(fmt.Sprintf("./reports/report_%v.txt", time.Now().Format("01-02-2006")), []byte(text), 0)
			if err != nil {
				log.Fatal(err)
				return ctx.Send("failed to write report")
			}
		}
		return ctx.Send("report sent successfully")
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
