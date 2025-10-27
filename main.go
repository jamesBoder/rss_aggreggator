package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/google/uuid"

	"github.com/jamesBoder/rss_aggreggator/internal/config"

	"github.com/jamesBoder/rss_aggreggator/internal/database"
)

// Create a state struct that holds a pointer to a config struct
type state struct {
	cfg *config.Config // The * goes before the package.Type

	// points to database queries
	db *database.Queries

	// point to raw *sql.DB
	dbSQL *sql.DB
}

// Create a command struct. Contains a name and a slice of string args
type command struct {
	Name string
	Args []string
}

// Create a commands struct. This will hold all the commands the CLI supports. Add a map[string]func(*State, *Command)error to hold command name to handler function mappings
type commands struct {
	Handlers map[string]func(*state, command) error
}

// implement register method
func (c *commands) register(name string, f func(*state, command) error) {
	c.Handlers[name] = f

}

// add a run method that runs a given command
func (c *commands) run(state *state, command command) error {
	// return error if command name not found in handlers map
	handler, exists := c.Handlers[command.Name]
	if !exists {
		return fmt.Errorf("unknown command: %s", command.Name)
	}
	// call the handler function
	return handler(state, command)
}

// handlerLogin handles the "login" command
func handlerLogin(state *state, command command) error {
	// ensure an arg exists
	if len(command.Args) < 1 {
		return fmt.Errorf("username argument is required")
	}
	// call GetUserByName with context.Background() and the username arg
	userName := command.Args[0]
	_, err := state.db.GetUserByName(context.Background(), userName)
	// compare error to sql.ErrNoRows and return error if so
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %s", userName)
		}
		return fmt.Errorf("error fetching user: %v", err)
	}

	// call SetUser on the config struct in state to set the current user name. exit with code 1 if name already exists
	if err := state.cfg.SetUser(userName); err != nil {
		return fmt.Errorf("error setting current user: %v", err)
	}

	fmt.Printf("Logged in as user: %s\n", userName)
	return nil
}

// Create handlerRegister to handle "register" command
func handlerRegister(state *state, command command) error {
	// ensure an arg exists
	if len(command.Args) < 1 {
		return fmt.Errorf("username argument is required")
	}
	// get username from args
	userName := command.Args[0]

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userName,
	}

	// call CreateUser with context.Background() and the username arg
	_, err := state.db.CreateUser(context.Background(), params)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			fmt.Println("user already exists")
			os.Exit(1)
		}
		return fmt.Errorf("error creating user: %v", err)
	}

	if err := state.cfg.SetUser(userName); err != nil {
		return fmt.Errorf("error setting current user: %v", err)
	}

	fmt.Printf("Registered and logged in as user: %s\n", userName)
	return nil
}

// add a reset command that deletes all users from the database

func handlerReset(state *state, command command) error {

	entries, err := os.ReadDir("sql/schema")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		fmt.Println("Applying:", e.Name())
		fmt.Println("Applied:", e.Name())

	}

	// // before running migrations, db.Exec a DROP TABLE IF EXISTS users CASCADE; db.Exec a DROP TABLE IF EXISTS feeds CASCADE; db.Exec a DROP EXTENSION IF EXISTS "pgcrypto";
	// db.Exec a CREATE EXTENSION IF NOT EXISTS "pgcrypto";
	_, err = state.dbSQL.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
	if err != nil {
		return fmt.Errorf("error creating extension: %v", err)
	}
	fmt.Println("Created extension successfully")

	// db.Exec a DROP TABLE IF EXISTS feeds CASCADE;
	_, err = state.dbSQL.Exec(`DROP TABLE IF EXISTS feeds CASCADE;`)
	if err != nil {
		return fmt.Errorf("error dropping feeds table: %v", err)
	}
	fmt.Println("Dropped feeds table successfully")

	// db.Exec a DROP TABLE IF EXISTS users CASCADE;
	_, err = state.dbSQL.Exec(`DROP TABLE IF EXISTS users CASCADE;`)
	if err != nil {
		return fmt.Errorf("error dropping users table: %v", err)
	}
	fmt.Println("Dropped users table successfully")

	// run migrations

	if err := runMigrations(state.dbSQL, "sql/schema"); err != nil {
		return fmt.Errorf("error running migrations: %v", err)
	}

	fmt.Println("All users deleted successfully")
	return nil
}

// add a users command that lists all users and prints them to the console in this format: * <user>. make sure current user is marked with (current)
func handlerUsers(state *state, command command) error {
	users, err := state.db.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching all users: %v", err)
	}

	currentUser := state.cfg.CurrentUserName

	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

// define RSSfeed struct to hold feed data
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

// define RSSItem struct to hold individual feed items
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// create fetchFeed function. Fetch a feed from the URL and return a RSSfeed struct pointer and error
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// http.NewRequestWithContext to create a new GET request with the given context and feedURL
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// set User-Agent header to "gator" with request.Header.Set
	req.Header.Set("User-Agent", "gator")

	// use http.Client.Do to send the request and get a response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %v", err)
	}

	// ensure resp.Body is closed after reading
	defer resp.Body.Close()

	// check if response status code is 200
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	// use io.ReadAll to read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// use xml.Unmarshal to parse the body into a RSSfeed struct
	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("error unmarshaling feed: %v", err)
	}

	// ensure both Title and Description are decoded using html.UnescapeString. Make sure no case changes, trimming, or other modifications are made
	feed.Channel.Title = decodeHTMLEntities(feed.Channel.Title)
	feed.Channel.Description = decodeHTMLEntities(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = decodeHTMLEntities(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = decodeHTMLEntities(feed.Channel.Item[i].Description)
	}

	// return the RSSfeed struct pointer

	return &feed, nil
}

// use html.UnescapeString to decode escaped HTML entities. Run Title and Description through this function before storing or displaying
func decodeHTMLEntities(s string) string {
	return html.UnescapeString(s)
}

// add agg command to fetch a single RSS feed and print the titles of the items to the console. It takes no arguments and should fetch the feed found at https://www.wagslane.dev/index.xml. Print entire parsed struct. Don't call decodeHTMLEntities here; it's called in fetchFeed.
func handlerAgg(state *state, command command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}

	// print the entire parsed struct
	// go
	fmt.Println(feed.Channel.Description)
	fmt.Printf("%+v\n", feed)

	return nil

}

// add command addFeed. It takes the name of the feed and the URL as arguments. At the top of the handler, get current user from the database and connect the feedto that user. the print out the fields of the new feed record.
func handlerAddFeed(state *state, command command) error {

	// get current user from the database
	if len(command.Args) < 2 {
		return fmt.Errorf("feed name and URL arguments are required")
	}

	// confirm a current user exists
	if state.cfg.CurrentUserName == "" {
		return fmt.Errorf("no user logged in")
	}

	// print current user
	fmt.Println("current user:", state.cfg.CurrentUserName)

	feedName := command.Args[0]
	feedURL := command.Args[1]

	user, err := state.db.GetUserByName(context.Background(), state.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error fetching current user: %v", err)
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		Name:      feedName,
		Url:       feedURL,
	}

	feed, err := state.db.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}

	fmt.Printf("Feed created: %+v\n", feed)
	return nil
}

func runMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		s := string(b)
		parts := strings.Split(s, "-- +goose Up")
		if len(parts) < 2 {
			continue
		}
		up := parts[1]
		if i := strings.Index(up, "-- +goose Down"); i >= 0 {
			up = up[:i]
		}
		up = strings.TrimSpace(up)
		if up == "" {
			continue
		}
		if _, err := db.Exec(up); err != nil {
			return fmt.Errorf("migration %s failed: %w", e.Name(), err)
		}
	}
	return nil
}

// add a feeds handler. it takes no arguments and prints all the fees in the database. Include name, Url, and name of user that created the feed.
func handlerFeeds(state *state, command command) error {
	feeds, err := state.db.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching all feeds: %v", err)
	}

	for _, feed := range feeds {
		user, err := state.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error fetching user for feed %s: %v", feed.Name, err)
		}
		fmt.Printf("* Name: %s, URL: %s, User: %s\n", feed.Name, feed.Url, user.Name)
	}
	return nil
}

func main() {

	// read config file
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	// store config file in a new instance of state struct
	s := &state{
		cfg: cfg,
	}

	// create new instance of commands struct with map of handler functions
	cmds := &commands{
		Handlers: make(map[string]func(*state, command) error),
	}

	// register the login command
	cmds.register("login", handlerLogin)

	// register the register command
	cmds.register("register", handlerRegister)

	// register the reset command
	cmds.register("reset", handlerReset)

	// register the users command
	cmds.register("users", handlerUsers)

	// register the agg command
	cmds.register("agg", handlerAgg)

	// register the addFeed command
	cmds.register("addfeed", handlerAddFeed)

	// register the feeds command
	cmds.register("feeds", handlerFeeds)

	// load database URL to the config struct and open a connection to dbURL using sql.Open

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// use generated database to create a new *database.Queries
	dbQueries := database.New(db)

	// store database queries in state struct
	s.db = dbQueries

	// store raw *sql.DB in state struct
	s.dbSQL = db

	// use os.Args to get command line args. if fewer than 2 args print error and exit
	args := os.Args
	if len(args) < 2 {
		log.Fatalf("No command provided")
	}

	// first arg is command name, rest are command args
	cmdName := args[1]
	cmdArgs := args[2:]
	cmd := command{
		Name: cmdName,
		Args: cmdArgs,
	}

	// run the command
	if err := cmds.run(s, cmd); err != nil {
		log.Fatalf("Error running command: %v", err)
	}

}
