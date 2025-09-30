package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const (
	dbRpsWallet = "rps_wallet"
)

// A component is a customizable, independent, and reusable UI element. It is created by
// embedding app.Compo into a struct.
type wallet struct {
	app.Compo
	sh              *shell.Shell
	myPeerID        string
	playerName      string
	debitAmount     float32
	creditAmount    float32
	transactionType TransactionType
	balance         int // cents
}

type Balance struct {
	ID     string `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`       // ID
	Amount int    `mapstructure:"amount" json:"amount" validate:"uuid_rfc4122"` // Amount
}

func (w *wallet) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	w.sh = sh

	myPeer, err := w.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	w.myPeerID = myPeer.ID

	ctx.GetState("playerName", &w.playerName)

	w.getBalance(ctx)

	w.transactionType = TypeDebit
}

func (w *wallet) OnNav(ctx app.Context) {
	url := ctx.Page().URL().Path
	path := strings.ReplaceAll(url, "/", "")
	linkElName := "link-" + path

	if !app.Window().GetElementByID(linkElName).IsNull() && !app.Window().GetElementByID(linkElName).IsNaN() && !app.Window().GetElementByID(linkElName).IsUndefined() {
		app.Window().GetElementByID(linkElName).Get("classList").Call("toggle", "active")
	}
}

func (w *wallet) getBalance(ctx app.Context) {
	ctx.Async(func() {
		accountJSON, err := w.sh.OrbitDocsGet(dbRpsWallet, w.playerName)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(accountJSON)) != "null" && len(accountJSON) > 0 {
			var balances []Balance

			err = json.Unmarshal(accountJSON, &balances)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				w.balance = balances[0].Amount
			})
		} else {
			w.createWallet(ctx)
		}
	})
}

func (w *wallet) createWallet(ctx app.Context) {
	wallet := Balance{
		ID:     w.playerName,
		Amount: 0,
	}

	walletJSON, err := json.Marshal(wallet)
	if err != nil {
		ctx.Notifications().New(app.Notification{
			Title: "Error",
			Body:  err.Error(),
		})
		return
	}

	ctx.Async(func() {
		err = w.sh.OrbitDocsPut(dbRpsWallet, walletJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			w.balance = 0
		})
	})
}

// The Render method is where the component appearance is defined.
func (w *wallet) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Form().
				Class("section").
				OnSubmit(w.doAction).
				Body(
					app.Div().
						Class("form-group").
						Body(
							app.H2().Text("Wallet"),
							app.Div().Class("balance-container").Body(
								app.P().Text("Balance: "),
								app.Span().ID("balance-amount").Text("€"+strconv.FormatFloat(float64(float32(w.balance)/100), 'f', 2, 32)),
							),
							app.Div().Class("tabs").Body(
								app.Button().ID("deposit-tablink").Class("tablink active").Text("Deposit").OnClick(w.openDepositTab),
								app.Button().ID("withdraw-tablink").Class("tablink").Text("Withdraw").OnClick(w.openWithdrawTab),
							),
							app.If(w.transactionType == TypeDebit, func() app.UI {
								return app.Div().ID("deposit-tab").Class("tabcontent").Body(
									app.Label().For("deposit-amount").Text("Deposit Amount"),
									app.Input().
										ID("deposit-amount").
										Name("deposit-amount").
										Type("number").
										Min(0.1).
										Step(0.1).
										Required(true).
										Placeholder("0.1").
										OnChange(w.ValueTo(&w.debitAmount)),
									app.Button().
										ID("deposit-btn").
										Type("submit").
										Text("Deposit"),
								)
							}).Else(func() app.UI {
								return app.Div().ID("withdraw-tab").Class("tabcontent").Body(
									app.Label().For("withdraw-amount").Text("Withdraw Amount"),
									app.Input().
										ID("withdraw-amount").
										Name("withdraw-amount").
										Type("number").
										Min(0.1).
										Step(0.1).
										Required(true).
										Placeholder("0.1").
										OnChange(w.ValueTo(&w.creditAmount)),
									app.Button().
										ID("withdraw-btn").
										Type("submit").
										Text("Withdraw"),
								)
							}),
						),
				),
		)
}

func (w *wallet) doAction(ctx app.Context, e app.Event) {
	e.PreventDefault()

	if w.transactionType == TypeDebit {
		amount := int(w.debitAmount * 100)

		w.updateBalance(ctx, amount)
	} else {
		amount := int(w.creditAmount * 100)
		if w.balance-amount < 0 {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  "Not enough funds",
			})
			return
		}

		w.updateBalance(ctx, amount)
	}
}

func (w *wallet) updateBalance(ctx app.Context, amount int) {
	var newBalance int
	if w.transactionType == TypeDebit {
		newBalance = w.balance + amount
	} else {
		newBalance = w.balance - amount
	}

	balance := Balance{
		ID:     w.playerName,
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

	ctx.Async(func() {
		err = w.sh.OrbitDocsPut(dbRpsWallet, ballanceJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		ctx.Dispatch(func(ctx app.Context) {
			w.balance = newBalance

			if w.transactionType == TypeDebit {
				ctx.Notifications().New(app.Notification{
					Title: "Success",
					Body:  "You have deposited €" + strconv.FormatFloat(float64(w.debitAmount), 'f', 2, 32),
				})
			} else {
				ctx.Notifications().New(app.Notification{
					Title: "Success",
					Body:  "You have withdrawn €" + strconv.FormatFloat(float64(w.creditAmount), 'f', 2, 32),
				})
			}

			w.storeTransaction(ctx, amount)

		})
	})
}

func (w *wallet) storeTransaction(ctx app.Context, amount int) {
	transaction := Transaction{
		ID:        uuid.NewString(),
		Username:  w.playerName,
		Type:      TransactionType(w.transactionType),
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

	ctx.Async(func() {
		err = w.sh.OrbitDocsPut(dbRpsTransaction, transactionJSON)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}
	})
}

func (w *wallet) openDepositTab(ctx app.Context, e app.Event) {
	e.PreventDefault()

	ctx.JSSrc().Set("value", 0.1)

	app.Window().GetElementByID("withdraw-amount").Set("value", 0.1)

	w.debitAmount = 0
	w.creditAmount = 0

	w.transactionType = TypeDebit

	app.Window().GetElementByID("withdraw-tablink").Get("classList").Call("remove", "active")
	app.Window().GetElementByID("deposit-tablink").Get("classList").Call("add", "active")
}

func (w *wallet) openWithdrawTab(ctx app.Context, e app.Event) {
	e.PreventDefault()

	ctx.JSSrc().Set("value", 0.1)

	app.Window().GetElementByID("deposit-amount").Set("value", 0.1)

	w.debitAmount = 0
	w.creditAmount = 0

	w.transactionType = TypeCredit

	app.Window().GetElementByID("deposit-tablink").Get("classList").Call("remove", "active")
	app.Window().GetElementByID("withdraw-tablink").Get("classList").Call("add", "active")
}
