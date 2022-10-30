package cli

import (
	"github.com/fatih/color"
)

func processJoin(commands []string, options Options) {
	if len(commands) == 0 {
		color.Red("join commands requires a token; to create a token get it from an admin or use the command token")
		return
	}

	// token, err := base64.StdEncoding.DecodeString(commands[0])
	// if err != nil {
	// 	color.Red("token is not in base64 format")
	// }

	// signLen := token[0]
	// sign, data := token[1:signLen+1], token[signLen+1:]

	// var accessToken model.AccessToken
	// err = json.Unmarshal(data, &accessToken)
	// if err != nil {
	// 	color.Red("token is invalid")
	// }

	// if !security.Verify(accessToken.Identity, data, sign) {
	// 	color.Red("invalid signature in token")
	// }

	// err = engine.Join(accessToken.Transport)
	// if err != nil {
	// 	color.Red("cannot save access token: %v", err)
	// 	return
	// }

	color.Green("access saved")
}
