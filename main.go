package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/apognu/gocal"
	"github.com/spf13/viper"
	"github.com/wneessen/go-mail"
)

type Conf struct {
	EmailFrom   string
	EmailTo     string
	STMPServer  string
	STMPPort    int
	Username    string
	Password    string
	IcsFileName string
}

func main() {
	m := mail.NewMsg()
	var conf Conf
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Fatal(err)
	}
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
		Classes: sortClasses(getTodayClasses(conf)),
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
	fmt.Fprintf(os.Stdout, "Send Success, %s", time.Now().String())
}

type class struct {
	Name      string
	Start     string
	End       string
	ClassRoom string
}

func getTodayClasses(conf Conf) []class {
	var classes []class
	// load ics file, you can get it from an app, download the app from https://www.wakeup.fun
	f, _ := os.Open(conf.IcsFileName)
	defer f.Close()

	// from now to end of the day
	loc, _ := time.LoadLocation("Asia/Shanghai")
	start := time.Now().Local().UTC().In(loc)
	end := time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, start.Location())

	c := gocal.NewParser(f)
	c.Start, c.End = &start, &end
	c.Parse()

	_, offsetSeconds := time.Now().Zone()
	offset := time.Duration(offsetSeconds * int(time.Second))
	if len(c.Events) > 0 {
		timeValue := c.Events[0].RawStart.Value
		if timeValue[len(timeValue)-1:] == "Z" {
			for _, e := range c.Events {
				classes = append(classes, class{Name: e.Summary, Start: e.Start.Add(offset).Format("15:04"), End: e.End.Add(offset).Format("15:04"), ClassRoom: e.Location})
			}
			return classes
		}
	}

	for _, e := range c.Events {
		classes = append(classes, class{Name: e.Summary, Start: e.Start.Format("15:04"), End: e.End.Format("15:04"), ClassRoom: e.Location})
	}
	return classes
}

func sortClasses(classes []class) []class {

	if len(classes) == 0 {
		return classes
	}
	// sort by start time
	for i := 0; i < len(classes); i++ {
		for j := i + 1; j < len(classes); j++ {
			if classes[i].Start > classes[j].Start {
				classes[i], classes[j] = classes[j], classes[i]
			}
		}
	}

	return classes
}
