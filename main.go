package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
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
	err := state.db.DeleteAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting all users: %v", err)
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
