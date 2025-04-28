package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Weso1ek/gator-blog-aggregator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, errExist := s.db.GetUser(context.Background(), name)
	if errExist != nil {
		return fmt.Errorf("user with name %s not exists", name)
	}

	err := s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	existUser, _ := s.db.GetUser(context.Background(), name)

	if existUser.Name != "" {
		return fmt.Errorf("user with name %s already exists", name)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: sql.NullTime{},
		UpdatedAt: sql.NullTime{},
	})

	if err != nil {
		return fmt.Errorf("couldn't save user: %w", err)
	}

	errSet := s.cfg.SetUser(name)
	if errSet != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User saved successfully!", user.Name)
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())

	if err != nil {
		return fmt.Errorf("couldn't list users: %w", err)
	}

	for _, j := range users {

		if j.Name == s.cfg.CurrentUserName {
			fmt.Println(j.Name + " (current)")
		} else {
			fmt.Println(j.Name)
		}
	}

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())

	if err != nil {
		return fmt.Errorf("couldn't delete users: %w", err)
	}

	fmt.Println("Users deleted successfully!")
	return nil
}
