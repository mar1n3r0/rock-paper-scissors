package main

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
)

const dbRpsTransaction = "rps_transaction"

type TransactionType string

const (
	TypeDebit  TransactionType = "debit"
	TypeCredit TransactionType = "credit"
)

type transaction struct {
	app.Compo
	sh           *shell.Shell
	myPeerID     string
	playerName   string
	transactions []Transaction
}

type Transaction struct {
	ID        string          `mapstructure:"_id" json:"_id" validate:"uuid_rfc4122"`             // ID
	Username  string          `mapstructure:"username" json:"username" validate:"uuid_rfc4122"`   // Username
	Type      TransactionType `mapstructure:"type" json:"type" validate:"uuid_rfc4122"`           // Type
	Amount    int             `mapstructure:"amount" json:"amount" validate:"uuid_rfc4122"`       // Amount
	Timestamp time.Time       `mapstructure:"timestamp" json:"timestamp" validate:"uuid_rfc4122"` // Timestamp
}

func (t *transaction) OnMount(ctx app.Context) {
	var loggedIn bool
	ctx.GetState("loggedIn", &loggedIn)
	if !loggedIn {
		ctx.Navigate("/")
		return
	}

	sh := shell.NewShell("localhost:5001")
	t.sh = sh

	myPeer, err := t.sh.ID()
	if err != nil {
		ctx.Navigate("/")
		return
	}

	t.myPeerID = myPeer.ID

	ctx.GetState("playerName", &t.playerName)

	t.getTransactions(ctx)
}

func (t *transaction) OnNav(ctx app.Context) {
	url := ctx.Page().URL().Path
	path := strings.ReplaceAll(url, "/", "")
	linkElName := "link-" + path

	if !app.Window().GetElementByID(linkElName).IsNull() && !app.Window().GetElementByID(linkElName).IsNaN() && !app.Window().GetElementByID(linkElName).IsUndefined() {
		app.Window().GetElementByID(linkElName).Get("classList").Call("toggle", "active")
	}
}

func (t *transaction) getTransactions(ctx app.Context) {
	ctx.Async(func() {
		transactionsJSON, err := t.sh.OrbitDocsQuery(dbRpsTransaction, "username", t.playerName)
		if err != nil {
			ctx.Notifications().New(app.Notification{
				Title: "Error",
				Body:  err.Error(),
			})
			return
		}

		if strings.TrimSpace(string(transactionsJSON)) != "null" && len(transactionsJSON) > 0 {
			var transactions []Transaction

			err = json.Unmarshal(transactionsJSON, &transactions)
			if err != nil {
				ctx.Notifications().New(app.Notification{
					Title: "Error",
					Body:  err.Error(),
				})
				return
			}

			ctx.Dispatch(func(ctx app.Context) {
				t.transactions = transactions

				// Sorting slice by timestamp descending
				sort.Slice(t.transactions, func(i, j int) bool {
					return t.transactions[i].Timestamp.After(t.transactions[j].Timestamp)
				})
			})
		}
	})
}

// The Render method is where the component appearance is defined.
func (t *transaction) Render() app.UI {
	return app.Div().
		Class("container").
		Body(
			newNav(),
			app.Div().ID("main").Body(
				app.Table().Body(
					app.TBody().Body(
						app.Tr().Body(
							app.Td().ID("table-header").Text("Transactions").ColSpan(3),
						),
						app.Tr().Body(
							app.Td().Text("ID"),
							app.Td().Text("Type"),
							app.Td().Text("Amount"),
							app.Td().Text("Timestamp"),
						),
						app.Range(t.transactions).Slice(func(i int) app.UI {
							return app.Tr().Body(
								app.Td().Text(t.transactions[i].ID),
								app.Td().Text(t.transactions[i].Type),
								app.Td().Text("€"+strconv.FormatFloat(float64(float32(t.transactions[i].Amount)/100), 'f', 2, 32)),
								app.Td().Text(t.transactions[i].Timestamp),
							)

						}),
					),
				),
			),
		)
}
