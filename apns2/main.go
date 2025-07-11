package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/uuneo/apns2"
	"github.com/uuneo/apns2/certificate"
)

type CLI struct {
	CertificatePath string `kong:"name=certificate-path,help='Path to certificate file.',required,short=c"`
	Topic           string `kong:"help='The topic of the remote notification, which is typically the bundle ID for your app',required,short=t"`
	Mode            string `kong:"help='APNS server to send notifications to. production or development. Defaults to production',default=production,short=m"`
}

func main() {
	var cli CLI
	parser := kong.Must(&cli,
		kong.Description(`Listens to STDIN to send notifications and writes APNS response code and reason to STDOUT.
The expected format is: <DeviceToken> <APNS Payload>
Example: aff0c63d9eaa63ad161bafee732d5bc2c31f66d552054718ff19ce314371e5d0 {"aps": {"alert": "hi"}}`),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	_, err := parser.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	cert, pemErr := certificate.FromPemFile(cli.CertificatePath, "")
	if pemErr != nil {
		log.Fatalf("Error retrieving certificate `%v`: %v", cli.CertificatePath, pemErr)
	}

	client := apns2.NewClient(cert)

	if cli.Mode == "development" {
		client.Development()
	} else {
		client.Production()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		in := scanner.Text()
		notificationArgs := strings.SplitN(in, " ", 2)
		if len(notificationArgs) < 2 {
			log.Println("Invalid input format, expected: <DeviceToken> <APNS Payload>")
			continue
		}
		token := notificationArgs[0]
		payload := notificationArgs[1]

		notification := &apns2.Notification{
			DeviceToken: token,
			Topic:       cli.Topic,
			Payload:     payload,
		}

		res, err := client.Push(notification)
		if err != nil {
			log.Fatal("Error: ", err)
		} else {
			fmt.Printf("%v: '%v'\n", res.StatusCode, res.Reason)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
