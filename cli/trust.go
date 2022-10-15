package cli

import (
	"weshare/security"
	"weshare/sql"

	"github.com/fatih/color"
)

func processTrust(commands []string, options Options) {
	if len(commands) < 2 {
		color.Red("domain, nick or identity are required\n")
		return
	}

	domain := commands[0]

	identityData := commands[1]
	identity, err := security.IdentityFromBase64(identityData)
	if err == nil {
		err = sql.SetTrusted(domain, identity, true)
		if err == nil {
			color.Green("%s is now a trusted user", identity.Nick)

		} else {
			color.Red("cannot set trust state in db")
		}
		return
	}

	nick := commands[1]
	users, err := sql.GetUsersByNick(domain, nick, true)
	if err != nil {
		color.Red("internal error - cannot query for nick")
		return
	}

	switch len(users) {
	case 0:
		color.Red("no nick '%s' in domain '%s'", nick)
	case 1:
		err = sql.SetTrusted(domain, users[0].Identity, true)
		if err != nil {
			color.Red("internal error - cannot trust nick '%s'", nick)
		}
	default:
		color.Red("more than one identities for nick '%s'; please use identity instead than nick")
	}

}
