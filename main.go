package main

import (
	"io/ioutil"
	"time"

	"github.com/fatih/color"
	"github.com/getlantern/systray"
	"github.com/markbates/pkger"
	"github.com/turnage/graw/reddit"
)

var nomailIcon []byte
var mailIcon []byte

var mailCh chan bool
var exitCh chan int8

func readFile(name string) ([]byte, error) {
	f, err := pkger.Open(name)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(f)

	if err != nil {
		return nil, err
	}

	return data, nil

}

func checkMail(bot reddit.Bot) (bool, error) {

	h, err := bot.ListingWithParams("/message/unread", map[string]string{"limit": "1"})

	if err != nil {
		return false, err
	}

	return len(h.Messages) > 0, nil

}

func main() {

	pkger.Include("/assets")

	var err error

	color.Green("Startup")

	mailCh = make(chan bool)
	exitCh = make(chan int8)

	mailIcon, err = readFile("/assets/mail.ico")

	if err != nil {
		panic(err)
	}

	nomailIcon, err = readFile("/assets/nomail.ico")

	if err != nil {
		panic(err)
	}

	color.Green("Images loaded")

	go systray.Run(onReady, onExit)

	var bot reddit.Bot

	bot, err = reddit.NewBotFromAgentFile("agent.txt", 0)

	if err != nil {
		panic(err)
	}

	color.Green("Reddit API initialized")

	var mail bool

	mail, err = checkMail(bot)

	if err != nil {
		panic(err)
	}

	mailCh <- mail

	timer := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-timer.C:
			mail, err = checkMail(bot)
			if err != nil {
				panic(err)
			}
			mailCh <- mail
		case <-exitCh:
			return
		}
	}

}

func onReady() {

	systray.SetIcon(nomailIcon)
	systray.SetTitle("Reddit Mailer")

	quit := systray.AddMenuItem("Quit", "Stop Reddit Mailer")

	color.Green("Systray ready")

	for {
		select {
		case <-quit.ClickedCh:
			systray.Quit()
		case b := <-mailCh:
			if b {
				systray.SetIcon(mailIcon)
			} else {
				systray.SetIcon(nomailIcon)
			}
		}
	}

}

func onExit() {
	exitCh <- 1
}
