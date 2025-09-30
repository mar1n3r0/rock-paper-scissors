package main

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type nav struct {
	app.Compo
}

func newNav() *nav {
	return &nav{}
}

func (n *nav) OnMount(ctx app.Context) {
	n.setupEventListener()
}

func (n *nav) setupEventListener() {
	app.Window().GetElementByID("menuBtn").Call("addEventListener", "click", app.FuncOf(func(this app.Value, args []app.Value) interface{} {
		app.Window().GetElementByID("sidebar").Get("classList").Call("toggle", "active")
		return nil
	}),
	)
}

func (n *nav) Render() app.UI {
	return app.Div().Class("header-container").Body(
		app.Header().Body(
			app.Button().ID("menuBtn").Aria("label", "Open Menu").Text("☰"),
			app.H1().Class("owText").Text("Rock || Paper || Scissors"),
		),
		app.Div().ID("sidebar").Body(
			app.Nav().Body(
				app.A().ID("link-home").Href("/home").Text("Home"),
				app.A().ID("link-wallet").Href("/wallet").Text("Wallet"),
				app.A().ID("link-players").Href("/players").Text("Challenge Players"),
				app.A().ID("link-challenges").Href("/challenges").Text("Pending Challenges"),
				app.A().ID("link-transactions").Href("/transactions").Text("Transactions"),
				app.A().ID("link-stats").Href("/stats").Text("Stats"),
				app.A().Href("#").Text("Logout").OnClick(n.doLogout),
			),
		),
	)
}

func (n *nav) doLogout(ctx app.Context, e app.Event) {
	e.PreventDefault()
	ctx.DelState("loggedIn")
	ctx.SetState("action", "logout")
	ctx.Navigate("/")
}
