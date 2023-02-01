package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/david_mbuvi/go_asterisks"
)

func GetCredentials() (string, string, error) {
	var (
		reader             = bufio.NewReader(os.Stdin)
		username, password string
		err                error
	)

	// read user input from terminal until the user input is not empty
	for username == "" {
		fmt.Print("Username or email: ")

		username, err = reader.ReadString('\n')
		if err != nil {
			return "", "", fmt.Errorf("error while reading username: %w", err)
		}

		username = strings.TrimSpace(username)
		if username == "" {
			log.Error("Username cannot be empty. Please try again.")
		}
	}

	// read user input from terminal until the user input is not empty
	for password == "" {
		fmt.Print("Password: ")

		bytePassword, err := go_asterisks.GetUsersPassword("", true, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Println(err.Error())
		}

		if bytePassword == nil {
			fmt.Println()
			log.Error("Password cannot be empty. Please try again.")
		}

		fmt.Println()

		password = string(bytePassword)
	}

	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

// HomeDir returns the invoking user's home directory.
func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicln(err)
	}

	return home
}
