package common

type RcsMsgRes struct {
	Rcs_id           interface{} `db:"rcs_id"`
	Msg_id           interface{} `db:"msg_id"`
	User_contact     interface{} `db:"user_contact"`
	Schedule_type    interface{} `db:"schedule_type"`
	Msg_group_id     interface{} `db:"msg_group_id"`
	Msg_service_type interface{} `db:"msg_service_type"`
	Chatbot_id       interface{} `db:"chatbot_id"`
	Agency_id        interface{} `db:"agency_id"`
	Messagebase_id   interface{} `db:"messagebase_id"`
	Service_type     interface{} `db:"service_type"`
	Expiry_option    interface{} `db:"expiry_option"`
	Header           interface{} `db:"header"`
	Footer           interface{} `db:"footer"`
	Copy_allowed     interface{} `db:"copy_allowed"`
	Body             interface{} `db:"body"`
	Buttons          interface{} `db:"buttons"`
	Status           interface{} `db:"status"`
	SentTime         interface{} `db:"senttime"`
	Timestamp        interface{} `db:"timestamp"`
	Error            interface{} `db:"error"`
	Proc             interface{} `db:"proc"`
}

type RcsStatusColumn struct {
	Rcs_id       interface{} `db:"rcs_id"`
	Msg_id       interface{} `db:"msg_id"`
	User_contact interface{} `db:"user_contact"`
	Status       interface{} `db:"status"`
	Service_type interface{} `db:"service_type"`
	Mno_info     interface{} `db:"mno_info"`
	Sent_time    interface{} `db:"sent_time"`
	Error        interface{} `db:"error"`
	Timestamp    interface{} `db:"timestamp"`
}

type AmtColumn struct {
	Amt_datetime interface{} `db:"amt_datetime"`
	Amt_kind     interface{} `db:"amg_kind"`
	Amt_amount   interface{} `db:"amt_amount"`
	Amt_memo     interface{} `db:"amt_memo"`
	Amt_reason   interface{} `db:"amt_reason"`
	Amt_payback  interface{} `db:"amt_payback"`
	Amt_admin    interface{} `db:"amt_admin"`
}

type OSmsColumn struct {
	Sender     interface{} `db:"Sender"`
	Receiver   interface{} `db:"Receiver"`
	Msg        interface{} `db:"Msg"`
	URL        interface{} `db:"URL"`
	ReserveDT  interface{} `db:"ReserveDT"`
	TimeoutDT  interface{} `db:"TimeoutDT"`
	SendResult interface{} `db:"SendResult"`
	Mst_id     interface{} `db:"mst_id"`
	Cb_msg_id  interface{} `db:"cb_msg_id"`
}

type OMmsColumn struct {
	MsgGroupID interface{} `db:"MsgGroupID"`
	Sender     interface{} `db:"Sender"`
	Receiver   interface{} `db:"Receiver"`
	Subject    interface{} `db:"Subject"`
	Msg        interface{} `db:"Msg"`
	ReserveDT  interface{} `db:"ReserveDT"`
	TimeoutDT  interface{} `db:"TimeoutDT"`
	SendResult interface{} `db:"SendResult"`
	File_Path1 interface{} `db:"File_Path1"`
	File_Path2 interface{} `db:"File_Path2"`
	File_Path3 interface{} `db:"File_Path3"`
	Mst_id     interface{} `db:"mst_id"`
	Cb_msg_id  interface{} `db:"cb_msg_id"`
}
