package cli

import (
	"encoding/json"
	"weshare/exchanges"
	"weshare/model"

	"github.com/fatih/color"
)

func processJoin(commands []string, options Options) {
	if len(commands) == 0 {
		color.Red("join commands requires a token; to create a token fill the below json and encode in base64")

		sampleToken := model.AccessToken{
			Access: model.Access{
				Domain: "domain name",
				Exchanges: []exchanges.Config{
					exchanges.SampleConfig,
				},
			},
			Identity: []byte("identity return by init or state"),
		}

		data, _ := json.Marshal(sampleToken)
		color.Green(string(data))
		return
	}

}
