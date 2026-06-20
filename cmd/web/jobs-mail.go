package main

import (
	"bytes"
	"html/template"
	"log"
	"strconv"
	"time"

	"github.com/aymerick/douceur/inliner"
	"github.com/halimi/vigilate/internal/channeldata"
	mail "github.com/xhit/go-simple-mail/v2"
	"jaytaylor.com/html2text"
)

type Worker struct {
	id         int
	jobQueue   chan channeldata.MailJob
	workerPool chan chan channeldata.MailJob
	quitChan   chan bool
}

func NewWorker(id int, workerPool chan chan channeldata.MailJob) Worker {
	return Worker{
		id:         id,
		jobQueue:   make(chan channeldata.MailJob),
		workerPool: workerPool,
		quitChan:   make(chan bool),
	}
}

func (w Worker) start() {
	go func() {
		for {
			// Add jobQueue to the worker pool.
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				w.processMailQueueJob(job.MailMessage)
			case <-w.quitChan:
				log.Printf("worker %d stopping", w.id)
				return
			}
		}
	}()
}

func (w Worker) stop() {
	go func() {
		w.quitChan <- true
	}()
}

func (w Worker) processMailQueueJob(mailMessage channeldata.MailData) {
	tmpl := "bootstrap.mail.tmpl"
	if mailMessage.Template != "" {
		tmpl = mailMessage.Template
	}

	t, ok := app.TemplateCache[tmpl]
	if !ok {
		log.Println("Could not get mail Template", mailMessage.Template)
		return
	}

	data := struct {
		Content       template.HTML
		From          string
		FromName      string
		PreferenceMap map[string]string
		IntMap        map[string]int
		StringMap     map[string]string
		FloatMap      map[string]float32
		RowSets       map[string]any
	}{
		Content:       mailMessage.Content,
		FromName:      mailMessage.FromName,
		From:          mailMessage.FromAddress,
		PreferenceMap: preferenceMap,
		IntMap:        mailMessage.IntMap,
		StringMap:     mailMessage.StringMap,
		FloatMap:      mailMessage.FloatMap,
		RowSets:       mailMessage.RowSets,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		log.Println(err)
	}

	result := tpl.String()

	plainText, err := html2text.FromString(result, html2text.Options{PrettyTables: true})
	if err != nil {
		plainText = ""
	}

	formattedMessage, err := inliner.Inline(result)
	if err != nil {
		log.Println(err)
		formattedMessage = result
	}

	port, _ := strconv.Atoi(preferenceMap["smtp_port"])

	server := mail.NewSMTPClient()
	server.Host = preferenceMap["smtp_server"]
	server.Port = port
	server.Username = preferenceMap["smtp_user"]
	server.Password = preferenceMap["smtp_password"]
	if preferenceMap["smtp_server"] == "localhost" {
		server.Authentication = mail.AuthPlain
	} else {
		server.Authentication = mail.AuthLogin
	}
	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		log.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(mailMessage.FromAddress).AddTo(mailMessage.ToAddress).SetSubject(mailMessage.Subject)

	if len(mailMessage.AdditionalTo) > 0 {
		for _, x := range mailMessage.AdditionalTo {
			email.AddTo(x)
		}
	}
	if len(mailMessage.CC) > 0 {
		for _, x := range mailMessage.CC {
			email.AddCc(x)
		}
	}
	if len(mailMessage.Attachments) > 0 {
		for _, x := range mailMessage.Attachments {
			email.AddAttachment(x)
		}
	}

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainText)

	err = email.Send(smtpClient)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent")
	}
}

type Dispatcher struct {
	workerPool chan chan channeldata.MailJob
	maxWorkers int
	jobQueue   chan channeldata.MailJob
}

func NewDispatcher(jobQueue chan channeldata.MailJob, maxWorkers int) *Dispatcher {
	workerPool := make(chan chan channeldata.MailJob, maxWorkers)
	return &Dispatcher{
		jobQueue:   jobQueue,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
	}
}

func (d *Dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i+1, d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func() {
				workerJobQueue := <-d.workerPool
				workerJobQueue <- job
			}()
		}
	}
}
