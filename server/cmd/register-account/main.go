package main

import (
	"flag"
	"log"

	"github.com/google/uuid"

	"lunar-tear/server/internal/auth"
	"lunar-tear/server/internal/database"
	"lunar-tear/server/internal/model"
	"lunar-tear/server/internal/store/sqlite"
)

func main() {
	dbPath := flag.String("db", "db/game.db", "SQLite database path")
	authdbPath := flag.String("auth-db", "db/auth.db", "SQLite auth server database path")

	name := flag.String("name", "", "Nickname of the new account to-be")
	password := flag.String("password", "", "Password of the new account to-be")
	platform := flag.String("platform", "android", "Platform of the user. Can be: \"android\", \"ios\"")

	flag.Parse()

	if *name == "" {
		log.Fatal("--name flag is required")
	}

	if *password == "" {
		log.Fatal("--password flag is required")
	}

	if (*platform != "android") && (*platform != "ios") {
		log.Fatal("--platform can be either \"android\" or \"ios\"")
	}

	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	userStore := sqlite.New(db, nil)

	authdb, err := database.Open(*authdbPath)
	if err != nil {
		log.Fatalf("open auth database: %v", err)
	}
	defer db.Close()

	authStore, err := auth.NewAuthStore(authdb)
	if err != nil {
		log.Fatalf("init auth store: %v", err)
	}

	// Auth user check

	userExists := authStore.UserExists(*name)
	if userExists {
		log.Fatal("Username is already taken")
	}

	// lunar-tear user

	var userPlatform model.ClientPlatform

	if *platform == "android" {
		userPlatform.OsType = 2
		userPlatform.PlatformType = 2
	} else {
		userPlatform.OsType = 1
		userPlatform.PlatformType = 1
	}

	userUuid := uuid.New().String()
	id, err := userStore.CreateUser(userUuid, userPlatform)

	if err == nil {
		log.Printf("Registered user %d in database successfully", id)
	} else {
		log.Fatalf("Register user in database: %v", err)
	}

	// Bind

	authUser, err := authStore.CreateUser(*name, *password)
	if err != nil {
		log.Fatalf("Register auth account: %v", err)
	}

	err = userStore.SetFacebookId(id, authUser.ID)
	if err == nil {
		log.Printf("Bound user %d with facebook account %v", id, authUser.Username)
	} else {
		log.Fatalf("failed to bind user with facebook account: %v", err)
	}

	log.Printf("Account %v created successfully.", *name)
}
