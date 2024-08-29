package handler

type OshotSmsTable struct {
	MsgId int 			`db:"MsgID"`
	Sender string		`db:"Sender"`
	Receiver string		`db:"Receiver"`
	Msg string			`db:"Msg"`
	Url string			`db:"URL"`
	ReserveDt string	`db:"ReserveDT"`
	TimeoutDt string	`db:"TimeoutDT"`
	SendDt string		`db:"SendDT"`
	SendResult int		`db:"SendResult"`
	Telecom string		`db:"Telecom"`
	InsertDt string		`db:"InsertDT"`
	MstId int			`db:"mst_id"`
	ProcFlag string		`db:"proc_flag"`
	CbMsgId string		`db:"cb_msg_id"`
	ResendFlag int 		`db:"resend_flag"`
}

type OshotMmsTable struct {
	MsgId int			`db:"MsgID"`
	MsgGroupId string	`db:"MsgGroupID"`
	Sender string		`db:"Sender"`
	Receiver string		`db:"Receiver"`
	Subject string		`db:"Subject"`
	Msg string			`db:"Msg"`
	ReserveDt string	`db:"ReserveDT"`
	TimeoutDt string	`db:"TimeoutDT"`
	SendDt string		`db:"SendDT"`
	SendResult int		`db:"SendResult"`
	Telecom string		`db:"Telecom"`
	FilePath1 string	`db:"File_Path1"`
	FilePath2 string	`db:"File_Path2"`
	FilePath3 string	`db:"File_Path3"`
	FilePath4 string	`db:"File_Path4"`
	InsertDt string		`db:"InsertDT"`
	MstId int			`db:"mst_id"`
	ProcFlag string		`db:"proc_flag"`
	CbMsgId string		`db:"cb_msg_id"`
	ResendFlag int		`db:"resend_flag"`
}

type NanoSmsTable struct {
	Num int						`db:"TR_NUM"`
	SendDate string				`db:"TR_SENDDATE"`
	SerialNum int 				`db:"TR_SERIALNUM"`
	Id string					`db:"TR_ID"`
	SendStat string				`db:"TR_SENDSTAT"`
	RsltStat string				`db:"TR_RSLTSTAT"`
	MsgType string				`db:"TR_MSGTYPE"`
	Phone string				`db:"TR_PHONE"`
	Callback string				`db:"TR_CALLBACK"`
	OrgCallback string			`db:"TR_ORG_CALLBACK"`
	BillId string				`db:"TR_BILL_ID"`
	RsltDate string				`db:"TR_RSLTDATE"`
	Modified string				`db:"TR_MODIFIED"`
	Msg string					`db:"TR_MSG"`
	Net string					`db:"TR_NET"`
	IdentificationCode string	`db:"TR_IDENTIFICATION_CODE"`
	Etc1 string					`db:"TR_ETC1"`
	Etc2 string					`db:"TR_ETC2"`
	Etc3 string					`db:"TR_ETC3"`
	Etc4 string					`db:"TR_ETC4"`
	Etc5 string					`db:"TR_ETC5"`
	Etc6 string					`db:"TR_ETC6"`
	Etc7 string					`db:"TR_ETC7"`
	Etc8 string					`db:"TR_ETC8"`
	Etc9 string					`db:"TR_ETC9"`
	Etc10 string				`db:"TR_ETC10"`
	RealSendData string			`db:"TR_REALSENDDATE"`
}

type NanoMmsTable struct {
	MsgKey string				`db:"MSGKEY"`
	Subject string				`db:"SUBJECT"`
	Phone string				`db:"PHONE"`
	Callback string				`db:"CALLBACK"`
	OrgCallback string			`db:"ORG_CALLBACK"`
	BillId string				`db:"BILL_ID"`
	Status string				`db:"STATUS"`
	ReqDate string				`db:"REQDATE"`
	Msg string					`db:"MSG"`
	FileCnt int 				`db:"FILE_CNT"`
	FileCntReal int 			`db:"FILE_CNT_REAL"`
	FilePath1 string			`db:"FILE_PATH1"`
	FilePath1Siz int 			`db:"FILE_PATH1_SIZ"`
	FilePath2 string			`db:"FILE_PATH2"`
	FilePath2Siz int 			`db:"FILE_PATH2_SIZ"`
	FilePath3 string			`db:"FILE_PATH3"`
	FilePath3Siz int 			`db:"FILE_PATH3_SIZ"`
	FilePath4 string			`db:"FILE_PATH4"`
	FilePath4Siz int 			`db:"FILE_PATH4_SIZ"`
	FilePath5 string			`db:"FILE_PATH5"`
	FilePath5Siz int 			`db:"FILE_PATH5_SIZ"`
	ExpireTime string			`db:"EXPIRETIME"`
	SendDate string				`db:"SENDDATE"`
	RsltDate string				`db:"RSLTDATE"`
	ReportDate string			`db:"REPORTDATE"`
	TerminatedDate string		`db:"TERMINATEDDATE"`
	Rslt string					`db:"RSLT"`
	RepCnt int 					`db:"REPCNT"`
	Type string					`db:"TYPE"`
	TelcoInfo string			`db:"TELCOINFO"`
	Id string					`db:"ID"`
	Post string					`db:"POST"`
	IdentificationCode string	`db:"IDENTIFICATION_CODE"`
}