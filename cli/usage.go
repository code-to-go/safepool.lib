package cli

import "fmt"

func usage() {
	fmt.Printf("usage: weshare <command>\n\n" +
		"Commands:\n" +
		"\tjoin <token>                       join a new domain; token is a base64 encoding of the access configuration\n" +
		"\tstate      				          returns the current identity and list the domains\n" +
		"\tstate <domain>[/path]              check the local and remote state of all files in the domain and path\n" +
		"\tadd <path>                         add files to the staging area\n" +
		"\ttrust <domain> [identity]       	  trust a new identity for a domain; identity is a base64 identity\n" +
		"\tupdate <path> ['']              synchronize with the closest exchange; when \n" +
		"\t-v                                 shows verbose log\n" +
		"\t-vv                                shows a very verbose log\n")
}
