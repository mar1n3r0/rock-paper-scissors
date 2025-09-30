package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	shell "github.com/stateless-minds/go-ipfs-api"
	// shell "github.com/stateless-minds/go-ipfs-api"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	// sh := shell.NewShell("localhost:5001")

	// populateItems(sh)

	// The first thing to do is to associate the hello component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", func() app.Composer { return &auth{} })
	app.Route("/home", func() app.Composer { return &home{} })
	app.Route("/wallet", func() app.Composer { return &wallet{} })
	app.Route("/players", func() app.Composer { return &player{} })
	app.RouteWithRegexp(`/match/([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$`, func() app.Composer { return &match{} })
	app.Route("/challenges", func() app.Composer { return &challenge{} })
	app.Route("/transactions", func() app.Composer { return &transaction{} })
	app.Route("/stats", func() app.Composer { return &stats{} })
	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	app.RunWhenOnBrowser()

	// Finally, launching the server that serves the app is done by using the Go
	// standard HTTP package.
	//
	// The Handler is an HTTP handler that serves the client and all its
	// required resources to make it work into a web browser. Here it is
	// configured to handle requests with a path that starts with "/".
	http.Handle("/", &app.Handler{
		Name:        "Rock Paper Scissors",
		Description: "Rock Paper Scissors - the classic game",
		Styles: []string{
			"web/app.css",
		},
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}

func populateItems(sh *shell.Shell) {
	err := sh.OrbitDocsDelete(dbRpsItem, "all")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("completed deletion of existing items")

	folder := "./web/assets"
	imagesMap, err := scanAndEncodeImages(folder)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var i int

	for filename, image := range imagesMap {
		re := regexp.MustCompile(`^(rock|paper|scissors)\.jpeg$`)
		matches := re.FindStringSubmatch(filename)

		switch matches[1] {
		case "rock":
			i = 1
		case "paper":
			i = 2
		case "scissors":
			i = 3
		}

		item := Item{
			ID:    strconv.Itoa(i),
			Name:  matches[1],
			Image: image,
		}

		itemJSON, err := json.Marshal(item)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("Filename: %s\n", filename) // Print a snippet for brevity

		err = sh.OrbitDocsPut(dbRpsItem, itemJSON)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("completed inserting items")
	}
}

// scanAndEncodeImages scans the given folder for image files and returns a map
// of filename to Base64-encoded image content.
func scanAndEncodeImages(folderPath string) (map[string]string, error) {
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
	}
	result := make(map[string]string)

	// Walk through the directory
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && imageExtensions[strings.ToLower(filepath.Ext(info.Name()))] {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			encoded := base64.StdEncoding.EncodeToString(data)
			result[info.Name()] = encoded
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
