package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

type stats struct {
	app.Compo
	sh              *shell.Shell
	myPeerID        string
	playerName      string
	wins            int
	draws           int
	losses          int
	totalAmountWon  int
	totalAmountLost int
}

func (s *stats) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	s.sh = sh

	myPeer, err := s.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	s.myPeerID = myPeer.ID

	ctx.GetState("playerName", &s.playerName)

	s.getChallengesWins(ctx)
	s.getChallengesDraws(ctx)
	s.getChallengesLosses(ctx)
}

func (s *stats) OnNav(ctx app.Context) {
	url := ctx.Page().URL().Path
	path := strings.ReplaceAll(url, "/", "")
	linkElName := "link-" + path

	if !app.Window().GetElementByID(linkElName).IsNull() && !app.Window().GetElementByID(linkElName).IsNaN() && !app.Window().GetElementByID(linkElName).IsUndefined() {
		app.Window().GetElementByID(linkElName).Get("classList").Call("toggle", "active")
	}
}

func (s *stats) getChallengesWins(ctx app.Context) {
	challengesJSON, err := s.sh.OrbitDocsQuery(dbRpsChallenge, "winner", s.playerName)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if strings.TrimSpace(string(challengesJSON)) != "null" && len(challengesJSON) > 0 {
		var challenges []Match

		err = json.Unmarshal(challengesJSON, &challenges)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		s.wins = len(challenges)

		for _, cc := range challenges {
			if cc.Host.Username == s.playerName {
				s.totalAmountWon += cc.BetAmount - cc.Host.Bet
			} else {
				s.totalAmountWon += cc.BetAmount - cc.Opponent.Bet
			}
		}
	}
}

func (s *stats) getChallengesLosses(ctx app.Context) {
	challengesJSON, err := s.sh.OrbitDocsQuery(dbRpsChallenge, "loser", s.playerName)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if strings.TrimSpace(string(challengesJSON)) != "null" && len(challengesJSON) > 0 {
		var challenges []Match

		err = json.Unmarshal(challengesJSON, &challenges)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		s.losses = len(challenges)

		for _, cc := range challenges {
			if cc.Host.Username == s.playerName {
				s.totalAmountLost += cc.Host.Bet
			} else {
				s.totalAmountLost += cc.Opponent.Bet
			}
		}
	}
}

func (s *stats) getChallengesDraws(ctx app.Context) {
	challengesJSON, err := s.sh.OrbitDocsQuery(dbRpsChallenge, "status", string(StatusDraw))
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if strings.TrimSpace(string(challengesJSON)) != "null" && len(challengesJSON) > 0 {
		var challenges []Match

		err = json.Unmarshal(challengesJSON, &challenges)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		for _, cc := range challenges {
			if cc.Host.Username == s.playerName || cc.Opponent.Username == s.playerName {
				s.draws++
			}
		}
	}
}

// The Render method is where the component appearance is defined.
func (s *stats) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Div().ID("main").Body(
				app.Table().Body(
					app.TBody().Body(
						app.Tr().Body(
							app.Td().ID("table-header").Text("Stats").ColSpan(6),
						),
						app.Tr().Body(
							app.Td().ColSpan(2).Text("Wins"),
							app.Td().ColSpan(2).Text("Draws"),
							app.Td().ColSpan(2).Text("Losses"),
						),
						app.Tr().Body(
							app.Td().ColSpan(2).Text(s.wins),
							app.Td().ColSpan(2).Text(s.draws),
							app.Td().ColSpan(2).Text(s.losses),
						),
						app.Tr().Body(
							app.Td().Text("Total Amount Won").ColSpan(3),
							app.Td().Text("Total Amount Lost").ColSpan(3),
						),
						app.Tr().Body(
							app.Td().ColSpan(3).Text("€"+strconv.FormatFloat(float64(float32(s.totalAmountWon)/100), 'f', 2, 32)),
							app.Td().ColSpan(3).Text("€"+strconv.FormatFloat(float64(float32(s.totalAmountLost)/100), 'f', 2, 32)),
						),
					),
				),
			),
		)
}
