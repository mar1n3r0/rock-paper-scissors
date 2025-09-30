package main

import (
	"encoding/json"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

type challenge struct {
	app.Compo
	sh          *shell.Shell
	myPeerID    string
	playerName  string
	challenges  []Match
	inChallenge bool
}

func (c *challenge) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	c.sh = sh

	myPeer, err := c.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	c.myPeerID = myPeer.ID

	ctx.GetState("playerName", &c.playerName)

	c.inChallenge = true

	c.getChallenges(ctx)
}

func (c *challenge) OnNav(ctx app.Context) {
	url := ctx.Page().URL().Path
	path := strings.ReplaceAll(url, "/", "")
	linkElName := "link-" + path

	if !app.Window().GetElementByID(linkElName).IsNull() && !app.Window().GetElementByID(linkElName).IsNaN() && !app.Window().GetElementByID(linkElName).IsUndefined() {
		app.Window().GetElementByID(linkElName).Get("classList").Call("toggle", "active")
	}
}

func (c *challenge) getChallenges(ctx app.Context) {
	ctx.Async(func() {
		accountJSON, err := c.sh.OrbitDocsQuery(dbRpsChallenge, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(accountJSON)) != "null" && len(accountJSON) > 0 {
			var challenges []Match

			err = json.Unmarshal(accountJSON, &challenges)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			for _, cc := range challenges {
				if cc.Host.Username == c.playerName && cc.Status != "" && !cc.HostNotified {
					if cc.Status != StatusPending {
						switch cc.Status {
						case StatusCompleted:
							if cc.Winner == c.playerName {
								c.notifyPlayer(ctx, "winner", cc.Opponent.Username)
							} else {
								c.notifyPlayer(ctx, "loser", cc.Opponent.Username)
							}
						case StatusDraw:
							c.notifyPlayer(ctx, "draw", cc.Opponent.Username)
						}

						cc.HostNotified = true
						c.saveChallenge(ctx, cc)
					}
				}
			}

			ctx.Dispatch(func(ctx app.Context) {
				c.challenges = challenges
			})
		}
	})
}

func (c *challenge) notifyPlayer(ctx app.Context, outcome string, opponent string) {
	switch outcome {
	case "winner":
		ctx.Notifications().New(app.Notification{
			Title: "Congrats",
			Body:  "You won your recent match with " + opponent,
		})
	case "loser":
		ctx.Notifications().New(app.Notification{
			Title: "Try again",
			Body:  "You lost your recent match with " + opponent,
		})
	case "draw":
		ctx.Notifications().New(app.Notification{
			Title: "A tie",
			Body:  "Your recent match with " + opponent + " ended in  a draw. Bets refunded.",
		})
	}
}

// The Render method is where the component appearance is defined.
func (c *challenge) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Div().ID("main").Body(
				app.Table().Body(
					app.TBody().Body(
						app.Tr().Body(
							app.Td().ID("table-header").Text("Challenges").ColSpan(3),
						),

						app.Range(c.challenges).Slice(func(i int) app.UI {
							return app.If(c.challenges[i].Status == StatusPending && c.challenges[i].Opponent.Username == c.playerName, func() app.UI {
								c.inChallenge = false
								return app.Tr().Body(
									app.Td().Text(c.challenges[i].Host.Username),
									app.Td().Body(
										app.Button().
											Class("challenge-btn").
											Text("Decline").
											Value(c.challenges[i].ID).
											OnClick(c.declineChallenge),
									),
									app.Td().Body(
										app.Button().
											Class("challenge-btn").
											Text("Accept").
											Value(c.challenges[i].ID).
											OnClick(c.acceptChallenge),
									),
								)
							})
						}),
					),
				),
			),
		)
}

func (c *challenge) acceptChallenge(ctx app.Context, e app.Event) {
	challengeID := ctx.JSSrc().Get("value").String()

	ctx.Dispatch(func(ctx app.Context) {
		ctx.Navigate("/match/" + challengeID)
	})
}

func (c *challenge) declineChallenge(ctx app.Context, e app.Event) {
	challengeID := ctx.JSSrc().Get("value").String()

	var challenge Match
	for _, c := range c.challenges {
		if c.ID == challengeID {
			challenge = c
		}
	}

	challenge.Status = StatusDeclined

	challengeJSON, err := json.Marshal(challenge)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = c.sh.OrbitDocsPut(dbRpsChallenge, challengeJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}
	})
}

func (c *challenge) saveChallenge(ctx app.Context, challenge Match) {
	challengeJSON, err := json.Marshal(challenge)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = c.sh.OrbitDocsPut(dbRpsChallenge, challengeJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}
	})
}
