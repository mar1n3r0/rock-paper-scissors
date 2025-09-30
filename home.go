package main

import (
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

// A component is a customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type home struct {
	app.Compo
	sh         *shell.Shell
	myPeerID   string
	playerName string
}

func (h *home) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	h.sh = sh

	myPeer, err := h.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	h.myPeerID = myPeer.ID

	ctx.GetState("playerName", &h.playerName)
}

func (h *home) OnNav(ctx app.Context) {
	url := ctx.Page().URL().Path
	path := strings.ReplaceAll(url, "/", "")
	linkElName := "link-" + path

	if !app.Window().GetElementByID(linkElName).IsNull() && !app.Window().GetElementByID(linkElName).IsNaN() && !app.Window().GetElementByID(linkElName).IsUndefined() {
		app.Window().GetElementByID(linkElName).Get("classList").Call("toggle", "active")
	}
}

// The Render method is where the component appearance is defined.
func (h *home) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Div().ID("main").Body(
				app.H2().Class("welcome").Text("Welcome "+h.playerName+"!"),
				app.P().Class("intro").Text("Open menu to get started."),
			),
		)
}
