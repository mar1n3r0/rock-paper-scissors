package main

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

type player struct {
	app.Compo
	sh         *shell.Shell
	myPeerID   string
	playerName string
	players    []Account
}

func (p *player) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	p.sh = sh

	myPeer, err := p.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	p.myPeerID = myPeer.ID

	ctx.GetState("playerName", &p.playerName)

	p.getPlayers(ctx)
}

func (p *player) OnNav(ctx app.Context) {
	url := ctx.Page().URL().Path
	path := strings.ReplaceAll(url, "/", "")
	linkElName := "link-" + path

	if !app.Window().GetElementByID(linkElName).IsNull() && !app.Window().GetElementByID(linkElName).IsNaN() && !app.Window().GetElementByID(linkElName).IsUndefined() {
		app.Window().GetElementByID(linkElName).Get("classList").Call("toggle", "active")
	}
}

func (p *player) getPlayers(ctx app.Context) {
	ctx.Async(func() {
		accountJSON, err := p.sh.OrbitDocsQuery(dbRpsAccount, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(accountJSON)) != "null" && len(accountJSON) > 0 {
			var players []Account

			err = json.Unmarshal(accountJSON, &players)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				p.players = players
			})
		}
	})
}

// The Render method is where the component appearance is defined.
func (p *player) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Div().ID("main").Body(
				app.Table().Body(
					app.TBody().Body(
						app.Tr().Body(
							app.Td().ID("table-header").Text("Players").ColSpan(2),
						),
						app.Range(p.players).Slice(func(i int) app.UI {
							return app.If(p.players[i].Username != p.playerName, func() app.UI {
								return app.Tr().Body(
									app.Td().Text(p.players[i].Username),
									app.Td().Body(
										app.Button().
											Class("challenge-btn").
											Text("Challenge").
											Value(p.players[i].Username).
											OnClick(p.challengePlayer),
									),
								)
							})
						}),
					),
				),
			),
		)
}

func (p *player) challengePlayer(ctx app.Context, e app.Event) {
	opponentUsername := ctx.JSSrc().Get("value").String()

	challenge := Match{
		ID:     uuid.NewString(),
		Status: StatusPending,
		Host: Selection{
			Username: p.playerName,
		},
		Opponent: Selection{
			Username: opponentUsername,
		},
	}

	p.createChallenge(ctx, challenge)
}

func (p *player) createChallenge(ctx app.Context, challenge Match) {
	challengeJSON, err := json.Marshal(challenge)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = p.sh.OrbitDocsPut(dbRpsChallenge, challengeJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			ctx.Navigate("/match/" + challenge.ID)
		})
	})
}
