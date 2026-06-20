package channeldata

import "html/template"

type MailData struct {
	ToName       string
	ToAddress    string
	FromName     string
	FromAddress  string
	AdditionalTo []string
	Subject      string
	Content      template.HTML
	Template     string
	CC           []string
	UseHermes    bool
	Attachments  []string
	StringMap    map[string]string
	IntMap       map[string]int
	FloatMap     map[string]float32
	RowSets      map[string]any
}

type MailJob struct {
	MailMessage MailData
}
