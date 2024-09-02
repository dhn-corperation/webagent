package handler

import (
	"database/sql"
)

type OshotSmsTable struct {
	MsgId int 					`db:"MsgID"`
	Sender string				`db:"Sender"`
	Receiver string				`db:"Receiver"`
	Msg string					`db:"Msg"`
	Url string					`db:"URL"`
	ReserveDt sql.NullString	`db:"ReserveDT"`
	TimeoutDt sql.NullString	`db:"TimeoutDT"`
	SendDt sql.NullString		`db:"SendDT"`
	SendResult int				`db:"SendResult"`
	Telecom sql.NullString		`db:"Telecom"`
	InsertDt string				`db:"InsertDT"`
	MstId int					`db:"mst_id"`
	ProcFlag string				`db:"proc_flag"`
	CbMsgId string				`db:"cb_msg_id"`
	ResendFlag int 				`db:"resend_flag"`
}

type OshotMmsTable struct {
	MsgId int					`db:"MsgID"`
	MsgGroupId string			`db:"MsgGroupID"`
	Sender string				`db:"Sender"`
	Receiver string				`db:"Receiver"`
	Subject string				`db:"Subject"`
	Msg string					`db:"Msg"`
	ReserveDt sql.NullString	`db:"ReserveDT"`
	TimeoutDt sql.NullString	`db:"TimeoutDT"`
	SendDt sql.NullString		`db:"SendDT"`
	SendResult int				`db:"SendResult"`
	Telecom sql.NullString		`db:"Telecom"`
	FilePath1 sql.NullString	`db:"File_Path1"`
	FilePath2 sql.NullString	`db:"File_Path2"`
	FilePath3 sql.NullString	`db:"File_Path3"`
	FilePath4 sql.NullString	`db:"File_Path4"`
	InsertDt string				`db:"InsertDT"`
	MstId int					`db:"mst_id"`
	ProcFlag string				`db:"proc_flag"`
	CbMsgId string				`db:"cb_msg_Id"`
	ResendFlag int				`db:"resend_flag"`
}

type NanoSmsTable struct {
	Num int								`db:"TR_NUM"`
	SendDate sql.NullString				`db:"TR_SENDDATE"`
	SerialNum sql.NullInt64 			`db:"TR_SERIALNUM"`
	Id sql.NullString					`db:"TR_ID"`
	SendStat string						`db:"TR_SENDSTAT"`
	RsltStat string						`db:"TR_RSLTSTAT"`
	MsgType string						`db:"TR_MSGTYPE"`
	Phone string						`db:"TR_PHONE"`
	Callback sql.NullString				`db:"TR_CALLBACK"`
	OrgCallback string					`db:"TR_ORG_CALLBACK"`
	BillId string						`db:"TR_BILL_ID"`
	RsltDate sql.NullString				`db:"TR_RSLTDATE"`
	Modified sql.NullString				`db:"TR_MODIFIED"`
	Msg string							`db:"TR_MSG"`
	Net sql.NullString					`db:"TR_NET"`
	IdentificationCode sql.NullString	`db:"TR_IDENTIFICATION_CODE"`
	Etc1 sql.NullString					`db:"TR_ETC1"`
	Etc2 sql.NullString					`db:"TR_ETC2"`
	Etc3 sql.NullString					`db:"TR_ETC3"`
	Etc4 sql.NullString					`db:"TR_ETC4"`
	Etc5 sql.NullString					`db:"TR_ETC5"`
	Etc6 sql.NullString					`db:"TR_ETC6"`
	Etc7 sql.NullString					`db:"TR_ETC7"`
	Etc8 sql.NullString					`db:"TR_ETC8"`
	Etc9 sql.NullString					`db:"TR_ETC9"`
	Etc10 sql.NullString				`db:"TR_ETC10"`
	RealSendData sql.NullString			`db:"TR_REALSENDDATE"`
}

type NanoMmsTable struct {
	MsgKey string						`db:"MSGKEY"`
	Subject sql.NullString				`db:"SUBJECT"`
	Phone string						`db:"PHONE"`
	Callback string						`db:"CALLBACK"`
	OrgCallback string					`db:"ORG_CALLBACK"`
	BillId string						`db:"BILL_ID"`
	Status string						`db:"STATUS"`
	ReqDate string						`db:"REQDATE"`
	Msg sql.NullString					`db:"MSG"`
	FileCnt int 						`db:"FILE_CNT"`
	FileCntReal int 					`db:"FILE_CNT_REAL"`
	FilePath1 sql.NullString			`db:"FILE_PATH1"`
	FilePath1Siz sql.NullInt64 			`db:"FILE_PATH1_SIZ"`
	FilePath2 sql.NullString			`db:"FILE_PATH2"`
	FilePath2Siz sql.NullInt64 			`db:"FILE_PATH2_SIZ"`
	FilePath3 sql.NullString			`db:"FILE_PATH3"`
	FilePath3Siz sql.NullInt64 			`db:"FILE_PATH3_SIZ"`
	FilePath4 sql.NullString			`db:"FILE_PATH4"`
	FilePath4Siz sql.NullInt64 			`db:"FILE_PATH4_SIZ"`
	FilePath5 sql.NullString			`db:"FILE_PATH5"`
	FilePath5Siz sql.NullInt64 			`db:"FILE_PATH5_SIZ"`
	ExpireTime sql.NullString			`db:"EXPIRETIME"`
	SentDate sql.NullString				`db:"SENTDATE"`
	RsltDate sql.NullString				`db:"RSLTDATE"`
	ReportDate sql.NullString			`db:"REPORTDATE"`
	TerminatedDate sql.NullString		`db:"TERMINATEDDATE"`
	Rslt sql.NullString					`db:"RSLT"`
	RepCnt sql.NullInt64 				`db:"REPCNT"`
	Type string							`db:"TYPE"`
	TelcoInfo sql.NullString			`db:"TELCOINFO"`
	Id sql.NullString					`db:"ID"`
	Post sql.NullString					`db:"POST"`
	IdentificationCode sql.NullString	`db:"IDENTIFICATION_CODE"`
	Etc1 sql.NullString					`db:"ETC1"`
	Etc2 sql.NullString					`db:"ETC2"`
	Etc3 sql.NullString					`db:"ETC3"`
	Etc4 sql.NullString					`db:"ETC4"`
	Etc5 sql.NullString					`db:"ETC5"`
	Etc6 sql.NullString					`db:"ETC6"`
	Etc7 sql.NullString					`db:"ETC7"`
	Etc8 sql.NullString					`db:"ETC8"`
	Etc9 sql.NullString					`db:"ETC9"`
	Etc10 sql.NullString				`db:"ETC10"`
}