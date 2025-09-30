package main

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbRpsChallenge = "rps_challenge"
const dbRpsItem = "rps_item"

type Status string

const (
	StatusPending   Status = "pending"
	StatusDeclined  Status = "declined"
	StatusDraw      Status = "draw"
	StatusCompleted Status = "completed"
)

type Outcome string

const (
	OutcomeWin  Outcome = "win"
	OutcomeDraw Outcome = "draw"
	OutcomeLoss Outcome = "loss"
)

type ItemType string

const (
	ItemRock     ItemType = "rock"
	ItemPaper    ItemType = "paper"
	ItemScissors ItemType = "scissors"
)

type match struct {
	app.Compo
	sh           *shell.Shell
	myPeerID     string
	playerName   string
	balance      int
	matchID      string
	match        Match
	items        []Item
	betAmount    float32
	selectedItem ItemType
	itemSelected bool
	outcome      Outcome
	winner       string
	loser        string
}

type Item struct {
	ID    string `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`     // ID
	Name  string `mapstructure:"name" json:"name" validate:"uuid_rfc4122"`   // Name
	Image string `mapstructure:"image" json:"image" validate:"uuid_rfc4122"` // Image base64
}

type Selection struct {
	Username string
	ItemName string
	Bet      int
}

type Match struct {
	ID           string    `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`                     // ID
	Status       Status    `mapstructure:"status" json:"status" validate:"uuid_rfc4122"`               // Status - pending, completed
	BetAmount    int       `mapstructure:"bet_amount" json:"bet_amount" validate:"uuid_rfc4122"`       // Amount in cents
	Host         Selection `mapstructure:"host" json:"host" validate:"uuid_rfc4122"`                   // Host selection
	Opponent     Selection `mapstructure:"opponent" json:"opponent" validate:"uuid_rfc4122"`           // Opponent selection
	Winner       string    `mapstructure:"winner" json:"winner" validate:"uuid_rfc4122"`               // Winner username
	Loser        string    `mapstructure:"loser" json:"loser" validate:"uuid_rfc4122"`                 // Loser username
	HostNotified bool      `mapstructure:"host_notified" json:"host_notified" validate:"uuid_rfc4122"` // Host Notified
}

func (m *match) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	m.sh = sh

	myPeer, err := m.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	m.myPeerID = myPeer.ID

	path := ctx.Page().URL().Path

	id := strings.TrimPrefix(path, "/match/")

	m.matchID = id

	match, err := m.matchExists()
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if !reflect.DeepEqual(match, Match{}) {
		m.match = match
	} else {
		ctx.Navigate("404")
		return
	}

	ctx.GetState("playerName", &m.playerName)

	balance, err := m.getBalance(ctx, m.playerName)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if !reflect.DeepEqual(balance, Balance{}) {
		if balance.Amount == 0 {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  "Your balance is zero. Top up and come back.",
			})
			ctx.Navigate("/wallet")
			return
		}

		m.balance = balance.Amount
	} else {
		ctx.Navigate("/wallet")
		return
	}

	m.getItems(ctx)
}

func (m *match) matchExists() (Match, error) {
	matchJSON, err := m.sh.OrbitDocsGet(dbRpsChallenge, m.matchID)
	if err != nil {
		return Match{}, err
	}

	var matches []Match

	if strings.TrimSpace(string(matchJSON)) != "null" && len(matchJSON) > 0 {
		err = json.Unmarshal(matchJSON, &matches)
		if err != nil {
			return Match{}, err
		}
	} else {
		return Match{}, nil
	}

	return matches[0], nil
}

func (m *match) getItems(ctx app.Context) {
	ctx.Async(func() {
		itemsJSON, err := m.sh.OrbitDocsQuery(dbRpsItem, "all", "")
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(itemsJSON)) != "null" && len(itemsJSON) > 0 {
			var items []Item

			err = json.Unmarshal(itemsJSON, &items)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				m.items = items

				// Sort by ID
				sort.Slice(items, func(i, j int) bool {
					vi, _ := strconv.Atoi(m.items[i].ID) // Convert string to int
					vj, _ := strconv.Atoi(m.items[j].ID)
					return vi < vj
				})
			})
		}
	})
}

func (m *match) getBalance(ctx app.Context, playerName string) (Balance, error) {
	accountJSON, err := m.sh.OrbitDocsGet(dbRpsWallet, playerName)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return Balance{}, err
	}

	if strings.TrimSpace(string(accountJSON)) != "null" && len(accountJSON) > 0 {
		var balances []Balance

		err = json.Unmarshal(accountJSON, &balances)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return Balance{}, err
		}

		return balances[0], nil
	} else {
		return Balance{}, nil
	}
}

func (m *match) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Form().
				Class("section").
				OnSubmit(m.updateMatch).
				Body(
					app.Div().
						Class("form-group").
						Body(
							app.H2().Text("Match"),
							app.Div().Class("span-container").Body(
								app.Span().Text("Balance: €"+strconv.FormatFloat(float64(float32(m.balance)/100), 'f', 2, 32)),
								app.Span().Text("Opponent: "+m.match.Opponent.Username),
							),
							app.Div().Body(
								app.Label().For("bet-amount").Text("Bet Amount"),
								app.Input().
									ID("bet-amount").
									Name("bet-amount").
									Type("number").
									Min(0.1).
									Step(0.1).
									Required(true).
									Placeholder("0.1").
									OnChange(m.ValueTo(&m.betAmount)),
								app.Span().Class("label").Text("Select Option"),
								app.Div().ID("inventory").Body(
									app.Range(m.items).Slice(func(i int) app.UI {
										return app.Div().ID("card-" + m.items[i].Name).Class("card").Body(
											app.Img().Class("selectable").DataSet("value", m.items[i].Name).Src("data:image/jpeg;base64," + m.items[i].Image).OnClick(m.selectItem),
										)
									}),
								),
								app.Button().
									ID("bet-btn").
									Type("submit").
									Text("Place Bet"),
							),
						),
				),
		)
}

func (m *match) selectItem(ctx app.Context, e app.Event) {
	e.PreventDefault()

	v := ctx.JSSrc().Call("getAttribute", "data-value").String()

	m.selectedItem = ItemType(v)

	items := app.Window().Get("document").Call("querySelectorAll", ".selectable")

	for i := 0; i < items.Length(); i++ {
		items.Index(i).Get("classList").Call("remove", "active")
	}

	ctx.JSSrc().Get("classList").Call("add", "active")

	m.itemSelected = true
}

func (m *match) updateMatch(ctx app.Context, e app.Event) {
	e.PreventDefault()

	if !m.itemSelected {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  "Please select option",
		})
		return
	}

	betAmount := int(m.betAmount * 100)

	if m.balance-betAmount < 0 {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  "Not enough funds",
		})
		return
	}

	// update balance
	newBalance := m.balance - betAmount

	m.updateBalance(ctx, m.playerName, newBalance)

	m.storeTransaction(ctx, m.playerName, TypeCredit, betAmount)

	if m.match.Opponent.Username == m.playerName {
		m.settleOutcome()
	}

	m.saveMatch(ctx, betAmount)

	if m.match.Opponent.Username == m.playerName {
		switch m.outcome {
		case OutcomeWin, OutcomeLoss:
			m.payWinner(ctx)
		case OutcomeDraw:
			m.doRefunds(ctx)
		}
	}

	m.notifyPlayer(ctx)

	ctx.Navigate("/challenges")
}

func (m *match) updateBalance(ctx app.Context, playerName string, newBalance int) {
	balance := Balance{
		ID:     playerName,
		Amount: newBalance,
	}

	ballanceJSON, err := json.Marshal(balance)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	err = m.sh.OrbitDocsPut(dbRpsWallet, ballanceJSON)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	m.balance = newBalance
}

func (m *match) storeTransaction(ctx app.Context, playerName string, transactionType TransactionType, amount int) {
	transaction := Transaction{
		ID:        uuid.NewString(),
		Username:  playerName,
		Type:      transactionType,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	err = m.sh.OrbitDocsPut(dbRpsTransaction, transactionJSON)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}
}

func (m *match) settleOutcome() {
	switch m.match.Host.ItemName {
	case string(ItemRock):
		switch m.selectedItem {
		case ItemRock:
			m.outcome = OutcomeDraw
		case ItemPaper:
			m.outcome = OutcomeWin
		case ItemScissors:
			m.outcome = OutcomeLoss
		}
	case string(ItemPaper):
		switch m.selectedItem {
		case ItemRock:
			m.outcome = OutcomeLoss
		case ItemPaper:
			m.outcome = OutcomeDraw
		case ItemScissors:
			m.outcome = OutcomeWin
		}
	case string(ItemScissors):
		switch m.selectedItem {
		case ItemRock:
			m.outcome = OutcomeWin
		case ItemPaper:
			m.outcome = OutcomeLoss
		case ItemScissors:
			m.outcome = OutcomeDraw
		}
	}

	switch m.outcome {
	case OutcomeWin:
		m.winner = m.match.Opponent.Username
		m.loser = m.match.Host.Username
	case OutcomeLoss:
		m.winner = m.match.Host.Username
		m.loser = m.match.Opponent.Username
	}
}

func (m *match) saveMatch(ctx app.Context, betAmount int) {
	if m.match.Host.Username == m.playerName {
		m.match.BetAmount = betAmount
		m.match.Host.ItemName = string(m.selectedItem)
		m.match.Host.Bet = betAmount
	} else {
		m.match.BetAmount += betAmount
		m.match.Opponent.ItemName = string(m.selectedItem)
		m.match.Opponent.Bet = betAmount

		if m.outcome != OutcomeDraw {
			m.match.Status = StatusCompleted
			m.match.Winner = m.winner
			m.match.Loser = m.loser
		} else {
			m.match.Status = StatusDraw
		}
	}

	matchJSON, err := json.Marshal(m.match)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	err = m.sh.OrbitDocsPut(dbRpsChallenge, matchJSON)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}
}

func (m *match) payWinner(ctx app.Context) {
	switch m.outcome {
	case OutcomeWin:
		newBalance := m.balance + m.match.BetAmount
		m.updateBalance(ctx, m.playerName, newBalance)
		m.storeTransaction(ctx, m.playerName, TypeDebit, m.match.BetAmount)
	case OutcomeLoss:
		balanceHost, err := m.getBalance(ctx, m.match.Host.Username)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if !reflect.DeepEqual(balanceHost, Balance{}) {
			newBalance := balanceHost.Amount + m.match.BetAmount
			m.updateBalance(ctx, m.match.Host.Username, newBalance)
			m.storeTransaction(ctx, m.match.Host.Username, TypeDebit, m.match.BetAmount)
		} else {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  "Could not transfer funds to winner. Wallet not found.",
			})
			return
		}
	}
}

func (m *match) doRefunds(ctx app.Context) {
	balanceOpponent := m.balance + m.match.Opponent.Bet
	m.updateBalance(ctx, m.playerName, balanceOpponent)
	m.storeTransaction(ctx, m.playerName, TypeDebit, m.match.Opponent.Bet)

	balanceHost, err := m.getBalance(ctx, m.match.Host.Username)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	if !reflect.DeepEqual(balanceHost, Balance{}) {
		newBalance := balanceHost.Amount + m.match.Host.Bet
		m.updateBalance(ctx, m.match.Host.Username, newBalance)
		m.storeTransaction(ctx, m.match.Host.Username, TypeDebit, m.match.Host.Bet)
	} else {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  "Could not transfer funds to winner. Wallet not found.",
		})
		return
	}
}

func (m *match) notifyPlayer(ctx app.Context) {
	switch m.outcome {
	case OutcomeWin:
		ctx.Notifications().New(app.Notification{
			Title: "Congrats",
			Body:  "You won the match",
		})
	case OutcomeLoss:
		ctx.Notifications().New(app.Notification{
			Title: "Try again",
			Body:  "You lost the match",
		})
	case OutcomeDraw:
		ctx.Notifications().New(app.Notification{
			Title: "A tie",
			Body:  "It was a draw. Bets refunded.",
		})
	case "":
		ctx.Notifications().New(app.Notification{
			Title: "Success",
			Body:  "Challenge created.",
		})
	}
}
