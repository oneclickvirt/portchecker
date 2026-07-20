package model

var Version = "v0.0.4"
var LocalServers = []string{"25", "465", "110", "995", "143", "993"}
var Platforms = []string{
	"QQ", "163", "Sohu", "Yandex", "Gmail", "Outlook", "Office365",
	"Yahoo", "MailCOM", "MailRU", "AOL", "GMX", "Sina", "Apple",
	"FastMail", "ProtonMail", "MXRoute", "Namecrane", "XYAMail",
	"ZohoMail", "Inbox_eu", "Free_fr",
}

// Domains contains the mail domain whose MX records should be tested for each
// platform. These are explicit because removing one label from an SMTP host is
// not valid for hosts such as smtp.mail.yahoo.com.
var Domains = map[string]string{
	"QQ": "qq.com", "163": "163.com", "Sohu": "sohu.com", "Yandex": "yandex.com",
	"Gmail": "gmail.com", "Outlook": "outlook.com", "Office365": "office365.com",
	"Yahoo": "yahoo.com", "MailCOM": "mail.com", "MailRU": "mail.ru", "AOL": "aol.com",
	"GMX": "gmx.com", "Sina": "sina.com", "Apple": "me.com", "FastMail": "fastmail.com",
	"ProtonMail": "proton.me", "MXRoute": "mxrouting.net", "Namecrane": "workspace.org",
	"XYAMail": "xyamail.com", "ZohoMail": "zoho.com", "Inbox_eu": "inbox.eu", "Free_fr": "free.fr",
}

var SmtpServers = map[string]string{
	"Gmail":      "smtp.gmail.com",
	"163":        "smtp.163.com",
	"Yandex":     "smtp.yandex.com",
	"Office365":  "smtp.office365.com",
	"QQ":         "smtp.qq.com",
	"Outlook":    "smtp.outlook.com",
	"Yahoo":      "smtp.mail.yahoo.com",
	"Apple":      "smtp.mail.me.com",
	"MailRU":     "smtp.mail.ru",
	"AOL":        "smtp.aol.com",
	"GMX":        "smtp.gmx.com",
	"MailCOM":    "smtp.mail.com",
	"Sohu":       "smtp.sohu.com",
	"Sina":       "smtp.sina.com",
	"FastMail":   "smtp.fastmail.com",
	"ProtonMail": "smtp.proton.me",
	"MXRoute":    "lisa.mxrouting.net", // 官方示例
	"Namecrane":  "us1.workspace.org",  // 示例
	"XYAMail":    "smtp.xyamail.com",
	"ZohoMail":   "smtp.zoho.com", //
	"Inbox_eu":   "mail.inbox.eu", // 官方示例
	"Free_fr":    "smtp.free.fr",  //
}

var Pop3Servers = map[string]string{
	"Gmail":      "pop.gmail.com",
	"163":        "pop.163.com",
	"Yandex":     "pop.yandex.com",
	"Office365":  "outlook.office365.com",
	"QQ":         "pop.qq.com",
	"Outlook":    "pop-mail.outlook.com",
	"Yahoo":      "pop.mail.yahoo.com",
	"Apple":      "pop.mail.me.com",
	"MailRU":     "pop.mail.ru",
	"AOL":        "pop.aol.com",
	"GMX":        "pop.gmx.com",
	"MailCOM":    "pop.mail.com",
	"Sohu":       "pop.sohu.com",
	"Sina":       "pop.sina.com",
	"FastMail":   "pop.fastmail.com",
	"ProtonMail": "pop.proton.me",
	"MXRoute":    "lisa.mxrouting.net",
	"Namecrane":  "us1.workspace.org",
	"XYAMail":    "pop.xyamail.com",
	"ZohoMail":   "pop.zoho.com",
	"Inbox_eu":   "mail.inbox.eu",
	"Free_fr":    "pop.free.fr",
}

var ImapServers = map[string]string{
	"Gmail":      "imap.gmail.com",
	"163":        "imap.163.com",
	"Yandex":     "imap.yandex.com",
	"Office365":  "outlook.office365.com",
	"QQ":         "imap.qq.com",
	"Outlook":    "imap-mail.outlook.com",
	"Yahoo":      "imap.mail.yahoo.com",
	"Apple":      "imap.mail.me.com",
	"MailRU":     "imap.mail.ru",
	"AOL":        "imap.aol.com",
	"GMX":        "imap.gmx.com",
	"MailCOM":    "imap.mail.com",
	"Sohu":       "imap.sohu.com",
	"Sina":       "imap.sina.com",
	"FastMail":   "imap.fastmail.com",
	"ProtonMail": "imap.proton.me",
	"MXRoute":    "lisa.mxrouting.net",
	"Namecrane":  "us1.workspace.org",
	"XYAMail":    "imap.xyamail.com",
	"ZohoMail":   "imap.zoho.com",
	"Inbox_eu":   "mail.inbox.eu",
	"Free_fr":    "imap.free.fr",
}
