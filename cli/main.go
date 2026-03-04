package main

import (
	"github.com/hookie-sh/hookie/cli/cmd"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	cmd.Execute()
}

