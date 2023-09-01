package main

import (
	"github.com/apognu/gocal"
	"github.com/spf13/viper"
	"github.com/wneessen/go-mail"
	"html/template"
	"log"
	"os"
	"time"
)

type Conf struct {
	EmailFrom  string
	EmailTo    string
	STMPServer string
	STMPPort   int
	Username   string
	Password   string
}

func main() {
	m := mail.NewMsg()
	var conf Conf
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	err := viper.Unmarshal(&conf)
	if err := m.From(conf.EmailFrom); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := m.To(conf.EmailTo); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}
	m.Subject("今日课程")
	//tpl := loadHtmlTemplate()
	tpl, err := template.ParseFiles("template.html")
	err = m.SetBodyHTMLTemplate(tpl, struct {
		Classes []class
	}{
		Classes: getTodayClasses(),
	})
	if err != nil {
		return
	}
	c, err := mail.NewClient(conf.STMPServer, mail.WithPort(conf.STMPPort), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(conf.Username), mail.WithPassword(conf.Password), mail.WithTLSPolicy(mail.TLSMandatory))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}
	if err := c.DialAndSend(m); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}
}

type class struct {
	Name  string
	Start string
	End   string
}

func getTodayClasses() []class {
	var classes []class
	// load ics file, you can get it from an app, download the app from https://www.wakeup.fun
	f, _ := os.Open("myclasses.ics")
	defer f.Close()

	// from now to end of the day
	loc, _ := time.LoadLocation("Asia/Shanghai")
	start := time.Now().In(loc)
	end := time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, start.Location())

	c := gocal.NewParser(f)
	c.Start, c.End = &start, &end
	c.Parse()

	for _, e := range c.Events {
		classes = append(classes, class{Name: e.Summary, Start: e.Start.Format("15:04"), End: e.End.Format("15:04")})
	}
	return classes
}
