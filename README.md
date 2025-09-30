# Rock Paper Scissors - demo game

---

## Game logic assumptions

- **A match has three possible outcomes - a win, draw or loss**
- **On draw - all bets are refunded**
- **All matches are non-realtime**
- **This means that the host and the opponent do not have time constraints to be in the game at the same time**
- **For reference to such a game check out <a href="https://github.com/stateless-minds/cyber-derive">Cyber-Derive</a> - A gamified delivery app which is based on concurrent play with time constraints**
- **Instead the outcome is resolved asynchronously**
- **The initiator of the game also called the host gets a notification next time he goes to his challenges if the match has been resolved meanwhile**
- **The challenged player also called the opponent gets an immediate notification about the outcome because he is always closing the match with his/her choice**

---

## Tech Stack

- **<a href="https://github.com/stateless-minds/kubo">IPFS Kubo fork with DB integration in the local daemon for access to db local files from browser wasm</a>**
- **<a href="https://go-app.dev/">go-app - A Golang frontend/fullstack/edge framework similar to React which compiles to browser-native wasm code</a>**

## How to run locally

1. Install Linux
2. Install IPFS Companion http://docs.ipfs.io/install/ipfs-companion/
3. Install golang 1.20 or later version - https://go.dev/doc/install
4.  Clone https://github.com/stateless-minds/kubo
`git clone https://github.com/stateless-minds/kubo.git`
5. Build IPFS
`make build`
6. Init IPFS
`./cmd/ipfs/ipfs init`
7.  Follow the instructions here to open your config file: https://github.com/ipfs/kubo/blob/master/docs/config.md. Usually it's `~/.ipfs/config` on Linux. Add the following snippet to the `HTTPHeaders`:
```
  "API": {
    "HTTPHeaders": {
      "Access-Control-Allow-Origin": ["webui://-", "http://localhost:3000", "http://127.0.0.1:5001", "https://webui.ipfs.io"],
      "Access-Control-Allow-Credentials": ["true"],
      "Access-Control-Allow-Methods": ["PUT", "POST", "GET"]
    }
  },
 ```
8. Run the daemon:
+ `.cmd/ipfs/ipfs daemon --enable-pubsub-experiment`

9.  Clone https://github.com/mar1n3r0/rock-paper-scissors
`git clone https://github.com/mar1n3r0/rock-paper-scissors.git`
10.  Do `make run`.
11. Head to localhost:3000 and you should see the authentication screen
12. In auth.go rename `a.myPeerID = myPeer.ID` for each new user you want to test with. For example `a.myPeerID = myPeer.ID + "player1"`

## How to run in online multiplayer mode

The app runs on the public IPFS network. In order to use it follow the steps below:

1. Install Linux
2. Install IPFS Companion http://docs.ipfs.io/install/ipfs-companion/
3. Install golang 1.20 or later version - https://go.dev/doc/install
4.  Clone https://github.com/stateless-minds/kubo to your local machine
`git clone https://github.com/stateless-minds/kubo.git`
5. Build IPFS
`make build`
6. Init IPFS
`./cmd/ipfs/ipfs init`
7.  Follow the instructions here to open your config file: https://github.com/ipfs/kubo/blob/master/docs/config.md. Usually it's `~/.ipfs/config` on Linux. Add the following snippet to the `HTTPHeaders`:
```
  "API": {
    "HTTPHeaders": {
      "Access-Control-Allow-Origin": ["webui://-", "http://k51qzi5uqu5dl77fp7duac3b9lkz4ytu44ipr4iwb83thn1vzoosfruuul624g.ipns.localhost:8080", "http://127.0.0.1:5001", "https://webui.ipfs.io"],
      "Access-Control-Allow-Credentials": ["true"],
      "Access-Control-Allow-Methods": ["PUT", "POST", "GET"]
    }
  },
 ```
8. Run the daemon:
+ `.cmd/ipfs/ipfs daemon --enable-pubsub-experiment`

9.  Navigate to <a href="https://ipfs.io/ipns/k51qzi5uqu5dl77fp7duac3b9lkz4ytu44ipr4iwb83thn1vzoosfruuul624g">Rock Paper Scissors</a>
10.  Pin it to your local node so that you cohost it every time your IPFS daemon is running
```
10.1. Open your IPFS dashboard

http://127.0.0.1:5001/webui

10.1 In your CLI with a running daemon run:

./cmd/ipfs/ipfs name resolve k51qzi5uqu5dl77fp7duac3b9lkz4ytu44ipr4iwb83thn1vzoosfruuul624g

Expected result for example:
/ipfs/QmZeRmYK1ugNVUGJ2VsSEUVwX6nqdfJbvKP6hAPBscVdoR

10.2. In the search bar of the web UI search for QmHash by pasting: QmZeRmYK1ugNVUGJ2VsSEUVwX6nqdfJbvKP6hAPBscVdoR
10.3 Click on More
10.4 Click Set Pinning
10.5 Mark local node and hit Apply
```

## FAQ

**1. How to add more choices other than the default rock, paper, scissors?**
 - Add new icons to web/assets/ with extension .jpeg
 - Uncomment the code at the top of main.go in main function which includes launching a new shell to local daemon and calling a function populateItems()
 - Adjust index for each item accordingly if you want them ordered
 - Add new constants for your new items at the top of match.go
 - Find settleOutcome() function in match.go and add your new outcome logic accordingly