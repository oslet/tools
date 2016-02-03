package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
)

type Smtp struct {
	User string
	Pwd  string
	Host string
	Port string
}

type Config struct {
	Smtp        Smtp
	From        string
	To          string
	Subject     string
	Body        string
	AttachFile  string
	ContentType string
}

var (
	config      Config
	smtpUser    string
	smtpPwd     string
	smtpHost    string
	smtpPort    string
	from        string
	to          string
	subject     string
	body        string
	attachFile  string
	contentType string
	show        bool
	help        bool
	auth        smtp.Auth
	boundary    string
)

func parse(body string) string {
	if len(strings.Split(body, "\n")) == 1 &&
		exists(body) {
		text, err := ioutil.ReadFile(body)
		if err != nil {
			log.Fatal(err)
		}
		body = string(text)
	}
	return body
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func getHeader() string {
	mailHeader := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\nMIME-Version: 1.0\r\n"+
			"Content-Type: multipart/mixed; boundary=%s\r\n--%s",
		from, to, subject, boundary, boundary)
	return mailHeader
}

func getBody() string {
	b := parse(body)
	mailBody := fmt.Sprintf(
		"\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Transfer-Encoding:8bit\r\n"+
			"\r\n%s\r\n--%s", contentType, b, boundary)
	return mailBody
}

func getAttach() string {
	var mailAttach string
	attachs := strings.Split(attachFile, ",")
	if len(attachs) > 0 {
		for i, a := range attachs {
			content, _ := ioutil.ReadFile(a)
			encoded := base64.StdEncoding.EncodeToString(content)
			fileName := path.Base(a)

			lineMaxLength := 500
			nbrLines := len(encoded) / lineMaxLength

			var buf bytes.Buffer
			for j := 0; j < nbrLines; j++ {
				buf.WriteString(encoded[j*lineMaxLength:(j+1)*lineMaxLength] + "\n")
			}

			buf.WriteString(encoded[nbrLines*lineMaxLength:])

			attachBytes, err := ioutil.ReadFile(a)
			if err != nil {
				log.Fatal(err)
			}
			mimeType := http.DetectContentType(attachBytes)
			if i == len(attachs)-1 {
				boundary += "--"
			}
			mailAttach += fmt.Sprintf(
				"\r\n"+
					"Content-Type: %s; name=\"%s\"\r\n"+
					"Content-Transfer-Encoding:base64\r\n"+
					"Content-Disposition: attachment; filename=\"%s\"\r\n\r\n%s\r\n--%s",
				mimeType, a, fileName, buf.String(), boundary)
		}
	}
	return mailAttach
}

func doSendMail(header, body, attach string) {
	err := smtp.SendMail(smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{to},
		[]byte(
			header+
				body+
				attach))
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	out := `
usage:
	go run src/sendmail.go [option] (default-config: config/default.toml)

option:
	-u, --user 			smtp login user
	-p, --password 		smtp login password
	-h, --host			smtp server host
	-P, --port 			stmp server port
	-f, --from			email sender
	-t, --to 			email recipient
	-s, --subject 		email subject
	-a, --attach        email attach file
	-c, --content-type	email body content-type
	-b, --body 			email body (message body or require file path)
	--show				view config
	--help			 	view usage

example:
	go run src/sendmail.go \
		-u account@gmail.com \
		-p password \
		-h smtp.gmail.com \
		-P 587 \
		-f sender@example.org \
		-t recipient@example.net \
		-s "Hello" \
		-a "/image.png" \
		-b "message body or require file path"
	`
	fmt.Println(out)
}

func showConfig() {
	fmt.Println(
		"[smtp]\n" +
			"user: " + smtpUser + "\n" +
			"pwd: " + smtpPwd + "\n" +
			"host: " + smtpHost + "\n" +
			"port: " + smtpPort + "\n" +
			"[mail]\n" +
			"from: " + from + "\n" +
			"to: " + to + "\n" +
			"subject: " + subject + "\n" +
			"attache: " + attachFile + "\n" +
			"content-type: " + contentType + "\n" +
			"body: " + body + "\n")
}

func setFlag() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.StringVar(&smtpUser, "u", config.Smtp.User, "smtp login user")
	f.StringVar(&smtpUser, "user", config.Smtp.User, "smtp login user")
	f.StringVar(&smtpPwd, "p", config.Smtp.Pwd, "smtp login password")
	f.StringVar(&smtpPwd, "password", config.Smtp.Pwd, "smtp login password")
	f.StringVar(&smtpHost, "h", config.Smtp.Host, "smtp server host")
	f.StringVar(&smtpHost, "host", config.Smtp.Host, "smtp server host")
	f.StringVar(&smtpPort, "P", config.Smtp.Port, "stmp server port")
	f.StringVar(&smtpPort, "port", config.Smtp.Port, "stmp server port")
	f.StringVar(&from, "f", config.From, "email sender")
	f.StringVar(&from, "from", config.From, "email sender")
	f.StringVar(&to, "t", config.To, "email recipient")
	f.StringVar(&to, "to", config.To, "email recipient")
	f.StringVar(&subject, "s", config.Subject, "email subject")
	f.StringVar(&subject, "subject", config.Subject, "email subject")
	f.StringVar(&attachFile, "a", config.AttachFile, "email attach file")
	f.StringVar(&attachFile, "attach", config.AttachFile, "email attach file")
	f.StringVar(&contentType, "c", config.ContentType, "email body content-type")
	f.StringVar(&contentType, "content-type", config.ContentType, "email body content-type")
	f.StringVar(&body, "b", config.Body, "email body text or require file path")
	f.StringVar(&body, "body", config.Body, "email body text or require file path")
	f.BoolVar(&show, "show", false, "View config")
	f.BoolVar(&help, "help", false, "View usage")
	f.Parse(os.Args[1:])
	for 0 < f.NArg() {
		f.Parse(f.Args()[1:])
	}
}

func setDefaultConfig(path string) {
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Fatal(err)
	}
}

func main() {
	//	setDefaultConfig("config/default.toml")
	setFlag()

	auth = smtp.PlainAuth("", smtpUser, smtpPwd, smtpHost)
	boundary = "PART"

	if help {
		usage()
		os.Exit(0)
	}

	if show {
		showConfig()
		os.Exit(0)
	}

	h := getHeader()
	b := getBody()
	a := getAttach()
	doSendMail(h, b, a)
}
