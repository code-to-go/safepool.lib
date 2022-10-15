package cli

import (
	"fmt"
	"weshare/engine"
	"weshare/model"
	"weshare/sql"

	"github.com/fatih/color"
)

func processState(commands []string, options Options) {
	if len(commands) == 0 {
		domains, err := sql.GetDomains()
		if err != nil {
			color.Red("internal error - cannot query db")
		}
		for _, d := range domains {
			color.Green(d)
		}
		return
	}

	domain := commands[0]

	files, err := engine.State(domain)
	if err != nil {
		color.Red("internal error - cannot query db")
		return
	}

	color.Blue("Local   Exchange  Staged   Name")
	for _, f := range files {
		line := " "
		if f.State&model.LocalCreated > 0 {
			line += "C"
		}
		if f.State&model.LocalModified > 0 {
			line += "M"
		}
		if f.State&model.LocalDeleted > 0 {
			line += "D"
		}
		if f.State&model.LocalRenamed > 0 {
			line += "R"
		}
		line = fmt.Sprintf("%-10s", line)

		if f.State&model.ExchangeCreated > 0 {
			line += "C"
		}
		if f.State&model.ExchangeModified > 0 {
			line += "M"
		}
		if f.State&model.ExchangeDeleted > 0 {
			line += "D"
		}
		if f.State&model.ExchangeRenamed > 0 {
			line += "R"
		}
		line = fmt.Sprintf("%-18s", line)
		if f.State&model.Staged > 0 {
			line += "⤒"
		}
		if f.State&model.Watched > 0 {
			line += "⤓"
		}

		line = fmt.Sprintf("%-24s", line)
		line += f.Name
		color.Green(line)
	}

}
