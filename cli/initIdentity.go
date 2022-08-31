package cli

import (
	"weshare/engine"

	"github.com/manifoldco/promptui"
)

func initIdentity() error {
	prompt := promptui.Prompt{
		Label: "which nick do you want?",
	}
	nick, _ := prompt.Run()
	_, err := engine.Init(nick, "")
	return err
}
