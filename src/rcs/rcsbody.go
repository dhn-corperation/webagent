package rcs

type MessageInfo struct {
	Common    	CommonInfo 	`json:"common, omitempty"`
	Rcs    		RcsInfo 	`json:"rcs, omitempty"`
}

type CommonInfo struct {
	Rcs_id 				int
	Msg_id 				string 		`json:"msgId, omitempty"`
	User_contact 		string 		`json:"userContact, omitempty"`
	Schedule_type 		int 		`json:"scheduleType, omitempty"`
	Msg_group_id 		string 		`json:"msgGroupId, omitempty"`
	Msg_service_type 	string 		`json:"msgServiceType, omitempty"`
}

type RcsInfo struct {
	Chatbot_id 			string 		`json:"chatbotId, omitempty"`
	Agency_id 			string 		`json:"agencyId, omitempty"`
	AgencyKey 			string 		`json:"agencyKey, omitempty"`
	BrandKey 			string 		`json:"brandKey, omitempty"`
	Messagebase_id 		string 		`json:"messagebaseId, omitempty"`
	Service_type 		string 		`json:"serviceType, omitempty"`
	Expiry_option 		int 		`json:"expiryOption, omitempty"`
	Header 				string 		`json:"header, omitempty"`
	Footer 				string 		`json:"footer, omitempty"`
	Copy_allowed 		bool 		`json:"copyAllowed, omitempty"`
	Body 				RcsBody		`json:"body, omitempty"`
	Buttons[]			RcsButton 	`json:"buttons, omitempty"`
}


type RcsBody struct {
	Title         string        `json:"title,omitempty"`
	Description   string        `json:"description,omitempty"`
    Media         string        `json:"media,omitempty"`
	Title1        string        `json:"title1,omitempty"`
	Description1  string        `json:"description1,omitempty"`
	Media1        string        `json:"media1,omitempty"`
	Title2        string        `json:"title2,omitempty"`
	Description2  string        `json:"description2,omitempty"`
	Media2        string        `json:"media2,omitempty"`
	Title3        string        `json:"title3,omitempty"`
	Description3  string        `json:"description3,omitempty"`
	Media3        string        `json:"media3,omitempty"`
	Title4        string        `json:"title4,omitempty"`
	Description4  string        `json:"description4,omitempty"`
	Media4        string        `json:"media4,omitempty"`
	Title5        string        `json:"title5,omitempty"`
	Description5  string        `json:"description5,omitempty"`
	Media5        string        `json:"media5,omitempty"`
}

type RcsButton struct {
	Suggestions[]		RcsSuggestions `json:"suggestions,omitempty"`
}

type RcsSuggestions struct {
	Action 			RcsAction    `json:"action,omitempty"`
}

type RcsAction struct {
	UrlAction 			RcsUrlAction    	`json:"urlAction,omitempty"`
	/*
	MapAction 			RcsMapAction    	`json:"mapAction,omitempty"`
	CalendarAction 		RcsCalendarAction   `json:"calendarAction,omitempty"`
	ClipboardAction 	RcsClipboardAction  `json:"clipboardAction,omitempty"`
	ComposeAction 		RcsComposeAction    `json:"composeAction,omitempty"`
	DialerAction 		RcsDialerAction    	`json:"dialerAction,omitempty"`
	ShareAction 		RcsShareAction    	`json:"shareAction,omitempty"`
	*/
	DisplayText         string          	`json:"displayText,omitempty"`
	Postback            RcsPostback     	`json:"postback,omitempty"`
}

type RcsUrlAction struct {
	OpenUrl 		RcsOpenUrl    `json:"openUrl,omitempty"`
}

type RcsOpenUrl struct {
	Url 		string    `json:"url,omitempty"`
}

type RcsPostback struct {
	Data 		string    `json:"data,omitempty"`
}


type RcsAuth struct {
	RcsId       string        `json:"rcsId,omitempty"`
	RcsSecret   string        `json:"rcsSecret,omitempty"`
	GrantType   string        `json:"grantType,omitempty"`
}

type RcsAuthResp struct {
	Status       string        `json:"status,omitempty"`
	Data  		 JsonData      `json:"data,omitempty"`
	Error  		 JsonError     `json:"error,omitempty"`
}

type JsonData struct {
	TokenInfo       JsonTokenInfo        `json:"tokenInfo,omitempty"`
}

type JsonError struct {
	Code       string        `json:"code,omitempty"`
	Message    string        `json:"message,omitempty"`
}

type RcsSendResp struct {
	Status       string        `json:"status,omitempty"`
	Error  		 JsonError     `json:"error,omitempty"`
}

type JsonTokenInfo struct {
	AccessToken   string        `json:"accessToken,omitempty"`
	ExpiresIn     string        `json:"expiresIn,omitempty"`
}

type RcsResultReq struct {
	QueryType	string        `json:"queryType,omitempty"`
	MsgId		string        `json:"msgId,omitempty"`
	RcsId		string        `json:"rcsId,omitempty"`
	QueryId		string        `json:"queryId,omitempty"`
	RetryYn		string        `json:"retryYn,omitempty"`
}

type RcsResultStatus struct {
	RcsId			string        `json:"rcsId,omitempty"`
	MsgId			string        `json:"msgId,omitempty"`
	UserContact		string        `json:"userContact,omitempty"`
	Status			string        `json:"status,omitempty"`
	ServiceType		string        `json:"serviceType,omitempty"`
	MnoInfo			string        `json:"mnoInfo,omitempty"`
	SentTime		string        `json:"sentTime,omitempty"`
	Timestamp		string        `json:"timestamp,omitempty"`
	Error			JsonError 	  `json:"error,omitempty"`
}

type RcsResultInfo struct {
	StatusInfo[]	RcsResultStatus  	`json:"statusInfo,omitempty"`
	QueryId			string        		`json:"queryId,omitempty"`
	MoreToSend		string        		`json:"moreToSend,omitempty"`
}
