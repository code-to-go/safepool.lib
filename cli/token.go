package cli

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"weshare/core"
	"weshare/engine"
	"weshare/exchanges"
	"weshare/model"
	"weshare/security"
	"weshare/sql"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func s3Wizard() *exchanges.Config {
	var region, endpoint, bucket, accessKey, secret string
	var done bool
	for !done {
		prompt := promptui.Prompt{
			Label:   "region",
			Default: region,
		}
		region, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "endpoint",
			Default: endpoint,
		}
		endpoint, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "bucket",
			Default: bucket,
		}
		bucket, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "access key",
			Default: accessKey,
		}
		accessKey, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "secret",
			Default: secret,
		}
		secret, _ = prompt.Run()

		color.Green("Current setup")
		color.Green("region: %s", region)
		color.Green("endpoint: %s", endpoint)
		color.Green("bucket: %s", bucket)
		color.Green("access key: %s", accessKey)
		color.Green("secret: %s", secret)

		sel := promptui.Select{
			Label: "Confirm",
			Items: []string{"Ok, all good", "Go back, need to fix", "Wrong exchange type, back to main"},
		}
		i, _, _ := sel.Run()
		done = i == 0
		if i == 2 {
			return nil
		}
	}

	return &exchanges.Config{
		S3: &exchanges.S3Config{
			Region:    region,
			Endpoint:  endpoint,
			Bucket:    bucket,
			AccessKey: accessKey,
			Secret:    secret,
		},
	}
}

func exchangesWizard(domain string) (model.Access, error) {
	var configs []exchanges.Config
	var c *exchanges.Config

done:
	for {
		prompt := promptui.Select{
			Label: "Choose the exchange type or done to complete and generate the token",
			Items: []string{"S3", "SFTP", "Azure", "Done"},
		}
		_, s, _ := prompt.Run()
		switch s {
		case "S3":
			c = s3Wizard()
		case "Done":
			break done
		}
		if c != nil {
			configs = append(configs, *c)
		}
	}

	return model.Access{
		Domain:    domain,
		Exchanges: configs,
	}, nil
}

func loadConfig(domain string, configFile string) (model.Access, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		color.Red("cannot read file '%s': %v\n", configFile, err)
		return model.Access{}, err
	}
	var config []exchanges.Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		color.Red("file '%s' is not a valid json configuration: %v\n", configFile, err)
		return model.Access{}, err
	}

	return model.Access{
		Domain:    domain,
		Exchanges: config,
	}, nil
}

func processToken(commands []string, options Options) {
	if len(commands) == 0 {
		color.Red("domain is required\n")
		return
	}

	var err error
	var access model.Access
	domain := commands[0]
	if len(commands) == 2 {
		configFile := commands[1]
		access, err = loadConfig(domain, configFile)
		if err != nil {
			return
		}
	} else {
		access, err = sql.GetAccess(domain)
		if err != nil {
			access, err = exchangesWizard(domain)
		}
	}

	identity, err := security.MarshalIdentity(engine.Self, false)
	if core.IsErr(err, "cannot marshal identity: %v") {
		return
	}
	data, err := json.Marshal(model.AccessToken{
		Access:   access,
		Identity: identity,
	})
	if core.IsErr(err, "cannot marshal access token to json: %v") {
		return
	}
	sign, err := security.Sign(engine.Self, data)
	if core.IsErr(err, "cannot sign access token: %v") {
		return
	}

	data = append([]byte{byte(len(sign))}, append(sign, data...)...)
	color.Green(base64.StdEncoding.EncodeToString(data))
}
