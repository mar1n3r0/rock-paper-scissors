package main

import (
	"encoding/json"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbRpsAccount = "rps_account"

// A component is a customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type auth struct {
	app.Compo
	notificationPermission app.NotificationPermission
	sh                     *shell.Shell
	myPeerID               string
	username               string
	registered             bool
	loggedIn               bool
	myAccount              Account
	allAccounts            []Account
	action                 string
}

type Account struct {
	ID       string `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`             // ID
	Username string `mapstructure:"username" json:"username" validate:"uuid_rfc4122"`   // Username
	LoggedIn bool   `mapstructure:"logged_in" json:"logged_in" validate:"uuid_rfc4122"` // LoggedIn
}

func (a *auth) OnMount(ctx app.Context) {
	a.notificationPermission = ctx.Notifications().Permission()
	switch a.notificationPermission {
	case app.NotificationDefault:
		a.notificationPermission = ctx.Notifications().RequestPermission()
	case app.NotificationDenied:
		app.Window().Call("alert", "In order to play Rock Paper Scissors notifications must be enabled")
		return
	}

	sh := shell.NewShell("localhost:5001")
	a.sh = sh

	myPeer, err := a.sh.ID()
	if err != nil {
		ctx.DelState("loggedIn")
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	a.myPeerID = myPeer.ID

	ctx.GetState("loggedIn", &a.loggedIn)

	if a.loggedIn {
		ctx.Navigate("/home")
	}

	ctx.GetState("action", &a.action)

	a.getAccounts(ctx)
}

func (a *auth) doLogout(ctx app.Context) {
	ctx.GetState("playerName", &a.username)

	for _, acc := range a.allAccounts {
		if acc.ID == a.myPeerID && acc.Username == a.username && acc.LoggedIn {
			acc.LoggedIn = false

			accountJSON, err := json.Marshal(acc)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Async(func() {
				err = a.sh.OrbitDocsPut(dbRpsAccount, accountJSON)
				if err != nil {
					ctx.Notifications().New(app.Notification{
						Title: "Error",
						Body:  err.Error(),
					})
					return
				}

				ctx.Dispatch(func(ctx app.Context) {
					ctx.DelState("playerName")

					ctx.Notifications().New(app.Notification{
						Title: "Success",
						Body:  "Logged out",
					})
					ctx.Update()
				})
			})
		}
	}
}

func (a *auth) getAccounts(ctx app.Context) {
	ctx.Async(func() {
		accountJSON, err := a.sh.OrbitDocsQuery(dbRpsAccount, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(accountJSON)) != "null" && len(accountJSON) > 0 {
			err = json.Unmarshal(accountJSON, &a.allAccounts)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				if a.action == "logout" {
					a.doLogout(ctx)
				}
			})
		}
	})
}

// The Render method is where the component appearance is defined.
func (a *auth) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			app.Div().Class("header-container").Body(
				app.Header().Class("landing").Body(
					app.H1().Class("owText").Text("Rock || Paper || Scissors"),
				),
			),
			app.Form().
				OnSubmit(a.OnSubmit).
				Body(
					app.Div().
						Class("form-group").
						Body(
							app.H2().Text("Authentication"),
							app.Input().
								ID("username").
								Name("username").
								Type("text").
								Placeholder("Enter username").
								Required(true).
								AutoComplete(true).
								OnInput(a.checkForDuplicates).
								OnChange(a.ValueTo(&a.username)),
							app.If(a.registered && !a.loggedIn, func() app.UI {
								return app.Button().
									ID("auth-btn").
									Type("submit").
									Text("Login").
									Value("login")
							}).Else(func() app.UI {
								return app.Button().
									ID("auth-btn").
									Type("submit").
									Text("Register").
									Value("register")
							}),
						),
				),
		)
}

func (a *auth) checkForDuplicates(ctx app.Context, e app.Event) {
	inputName := ctx.JSSrc().Get("value").String()

	var allAccounts []Account

	for i, acc := range a.allAccounts {
		if acc.ID == a.myPeerID && inputName == acc.Username {
			a.registered = true
			a.myAccount = a.allAccounts[i]
			ctx.Update()
		} else {
			allAccounts = append(a.allAccounts, acc)
		}
	}

	a.allAccounts = allAccounts
}

func (a *auth) OnSubmit(ctx app.Context, e app.Event) {
	e.PreventDefault()

	action := app.Window().GetElementByID("auth-btn").Get("value").String()

	if action == "register" {
		for _, acc := range a.allAccounts {
			if a.username == acc.Username {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  "Username is taken",
				})
				return
			}
		}
		a.registerAccount(ctx)
	} else {
		if a.myAccount.LoggedIn {
			a.loggedIn = true
			ctx.SetState("loggedIn", true).Persist()
			ctx.SetState("playerName", a.myAccount.Username).Persist()
			ctx.Navigate("/home")
		} else {
			a.loginAccount(ctx)
		}
	}
}

func (a *auth) registerAccount(ctx app.Context) {
	account := Account{
		ID:       a.myPeerID,
		Username: a.username,
		LoggedIn: true,
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = a.sh.OrbitDocsPut(dbRpsAccount, accountJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.Notifications().New(app.Notification{
				Title: "Success",
				Body:  "Registration completed",
			})

			ctx.SetState("loggedIn", true).Persist()
			ctx.SetState("playerName", account.Username).Persist()
			ctx.Navigate("/home")
		})
	})
}

func (a *auth) loginAccount(ctx app.Context) {
	a.myAccount.LoggedIn = true

	accountJSON, err := json.Marshal(a.myAccount)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = a.sh.OrbitDocsPut(dbRpsAccount, accountJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.Notifications().New(app.Notification{
				Title: "Success",
				Body:  "Logged in",
			})

			ctx.SetState("loggedIn", true).Persist()
			ctx.SetState("playerName", a.myAccount.Username).Persist()
			ctx.Navigate("/home")
		})
	})
}
