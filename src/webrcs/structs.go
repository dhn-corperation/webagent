package webrcs

type RcsAuthRequest struct {
	RcsId       string        `json:"rcsId"`
	RcsSecret   string        `json:"rcsSecret"`
	GrantType   string        `json:"grantType"`
}

type RcsAuthResponse struct {
	Status       string        `json:"status"`
	Data  		 RcsData      `json:"data,omitempty"`
	Error  		 RcsErrorInfo     `json:"error,omitempty"`
}

type RcsResultRequest struct {
	QueryType string `json:"queryType"`
	MsgId string `json:"msgId,omitempty"`
	RcsId string `json:"rcsId,omitempty"`
	QueryId string `json:"queryId,omitempty"`
	RetryYn string `json:"retryYn,omitempty"`
}

type RcsData struct {
	TokenInfo RcsTokenInfo `json:"tokenInfo"`
}

type RcsTokenInfo struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn string `json:"expiresIn"`
}

type RcsFileInfo struct {
	FileId string `json:"fileId,omitempty"`
	UsageType string `json:"usageType"`
	UsageService string `json:"usageService"`
	MimeType string `json:"mimeType"`
	Status string `json:"status"`
	Size string `json:"size,omitempty"`
	ExpiryDate string `json:"expiryDate"`
	Url string `json:"url"`
	MessagebaseId string `json:"messagebaseId,omitempty"`
}

type RcsSendRequest struct {
	Common RcsCommonInfo `json:"common"`
	Rcs RcsInfo `json:"rcs"`
	Legacy *RcsLegacyInfo `json:"legacy,omitempty"`
}

type RcsCommonInfo struct {
	MsgId string `json:"msgId"`
	UserContact string `json:"userContact"`
	ScheduleType int `json:"scheduleType,omitempty"`
	MsgGroupId string `json:"msgGroupId,omitempty"`
	MsgServiceType string `json:"msgServiceType"`
}

//////////////////////// RcsInfo Area ////////////////////////

type RcsInfo struct {
	ChatbotId string `json:"chatbotId"`
	AgencyId string `json:"agencyId,omitempty"`
	AgencyKey string `json:"agencyKey,omitempty"`
	BrandKey string `json:"brandKey,omitempty"`
	MessagebaseId string `json:"messagebaseId"`
	ServiceType string `json:"serviceType"`
	ExpiryOption int `json:"expiryOption,omitempty"`
	Header string `json:"header"`
	Footer string `json:"footer,omitempty"`
	CdrId string `json:"cdrId,omitempty"`
	CopyAllowed bool `json:"copyAllowed,omitempty"`
	Body RcsBody `json:"body"`
	Buttons RcsButtons `json:"buttons,omitempty"`
	ChipList []RcsChipList `json:"chipList,omitempty"`
	ReplyId string `json:"replyId,omitempty"`
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

	SubTitle1        string        `json:"subTitle1,omitempty"`
	SubDesc1        string        `json:"subDesc1,omitempty"`
	SubTitle2        string        `json:"subTitle2,omitempty"`
	SubDesc2        string        `json:"subDesc2,omitempty"`
	SubTitle3        string        `json:"subTitle3,omitempty"`
	SubDesc3        string        `json:"subDesc3,omitempty"`

	SubMedia1        string        `json:"subMedia1,omitempty"`
	SubMediaUrl1        string        `json:"subMediaUrl1,omitempty"`
	SubMedia2        string        `json:"subMedia2,omitempty"`
	SubMediaUrl2        string        `json:"subMediaUrl2,omitempty"`
	SubMedia3        string        `json:"subMedia3,omitempty"`
	SubMediaUrl3        string        `json:"subMediaUrl3,omitempty"`

	MTitle        string        `json:"mTitle,omitempty"`
	MTitleMedia        string        `json:"mTitleMedia,omitempty"`

	ItemTitle        string        `json:"itemTitle,omitempty"`
	ItemDesc        string        `json:"itemDesc,omitempty"`
	ItemMedia        string        `json:"itemMedia,omitempty"`

	CellTitle        string        `json:"cellTitle,omitempty"`
	CellLeft1        string        `json:"cellLeft1,omitempty"`
	CellRight1        string        `json:"cellRight1,omitempty"`
	CellLeft2        string        `json:"cellLeft2,omitempty"`
	CellRight2        string        `json:"cellRight2,omitempty"`
}


type RcsButtons struct {
	Suggestions		[]RcsSuggestions `json:"suggestions,omitempty"`
}

type RcsSuggestions struct {
	Action 			RcsAction    `json:"action,omitempty"`
	Reply 			RcsAction    `json:"reply,omitempty"`
}

type RcsAction struct {
	UrlAction 			RcsUrlAction    	`json:"urlAction,omitempty"`
	DisplayText         string          	`json:"displayText,omitempty"`
	Postback            RcsPostback     	`json:"postback,omitempty"`
	ComposeAction       RcsComposeAction     	`json:"composeAction,omitempty"`
	MapAction       RcsMapAction     	`json:"mapAction,omitempty"`
	LocalBrowserAction RcsLocalBrowserAction `json:"localBrowserAction,omitempty"`
	DialerAction RcsDialerAction `json:"dialerAction,omitempty"`
	ClipboardAction RcsClipboardAction `json:"clipboardAction,omitempty"`
	CalendarAction RcsCalendarAction `json:"calendarAction,omitempty"`
}

type RcsUrlAction struct {
	OpenUrl 		RcsOpenUrl    `json:"openUrl,omitempty"`
}

type RcsOpenUrl struct {
	Url 		string    `json:"url,omitempty"`
	IsHalfView string `json:"isHalfView,omitempty"`
	PostParameter RcsPostParameter `json:"postParameter,omitempty"`
}

type RcsPostback struct {
	Data 		string    `json:"data,omitempty"`
}

type RcsChipList struct {
	Reply RcsAction `json:"reply"`
	Action RcsAction `json:"action"`
}

type RcsComposeAction struct {
	ComposeTextMessage RcsComposeTextMessage `json:"composeTextMessage,omitempty"`
}

type RcsComposeTextMessage struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Text string `json:"text,omitempty"`
}

type RcsMapAction struct {
	ShowLocation RcsShowLocation `json:"showLocation,omitempty"`
	RequestLocationPush RcsRequestLocationPush `json:"requestLocationPush,omitempty"`
}

type RcsShowLocation struct {
	FallbackUrl string `json:"fallbackUrl,omitempty"`
	Location RcsLocation `json:"location,omitempty"`
}

type RcsLocation struct {
	Query string `json:"query,omitempty"`
	Latitude string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
	Label string `json:"label,omitempty"`
}

type RcsRequestLocationPush struct {

}

type RcsLocalBrowserAction struct {
	OpenUrl RcsOpenUrl `json:"openUrl,omitempty"`
}

type RcsPostParameter struct {
	PName string `json:"P_NAME,omitempty"`
	PMid string `json:"P_MID,omitempty"`
}

type RcsDialerAction struct {
	DialPhoneNumber RcsDialPhoneNumber `json:"dialPhoneNumber,omitempty"`
}

type RcsDialPhoneNumber struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type RcsClipboardAction struct {
	CopyToClipboard RcsCopyToClipboard `json:"copyToClipboard,omitempty"`
}

type RcsCopyToClipboard struct {
	Text string `json:"text,omitempty"`
}

type RcsCalendarAction struct {
	CreateCalendarEvent RcsCreateCalendarEvent `json:"createCalendarEvent,omitempty"`
}

type RcsCreateCalendarEvent struct {
	StartTime string `json:"startTime,omitempty"`
	EndTime string `json:"endTime,omitempty"`
	Title string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

//////////////////////// RcsInfo Area ////////////////////////

type RcsLegacyInfo struct {
	ServiceType string `json:"serviceType"`
	Callback string `json:"callback"`
	Subject string `json:"subject,omitempty"`
	Msg string `json:"msg"`
	ContentCount int16 `json:"contentCount"`
	ContentData string `json:"contentData,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	KisaOrigCode int32 `json:"kisaOrigCode"`
}

type RcsStatusInfo struct {
	RcsId string `json:"rcsId,omitempty"`
	MsgId string `json:"msgId"`
	UserContact string `json:"userContact,omitempty"`
	Status string `json:"status"`
	ServiceType string `json:"serviceType,omitempty"`
	MnoInfo string `json:"mnoInfo,omitempty"`
	SentTime string `json:"sentTime,omitempty"`
	Reason RcsReasonInfo `json:"reason,omitempty"`
	Error RcsErrorInfo `json:"error,omitempty"`
	LegacyError RcsErrorInfo `json:"legacyError,omitempty"`
	Timestamp string `json:"timestamp"`
	AutoReplyMsgId string `json:"autoReplyMsgId,omitempty"`
	PostbackId string `json:"postbackId,omitempty"`
	ChatbotId string `json:"chatbotId,omitempty"`
	Bill int16 `json:"bill,omitempty"`
}

type RcsQuerystatusInfo struct {
	QueryId string `json:"queryId,omitempty"`
	MoreToSend int16 `json:"moreToSend,omitempty"`
}

type RcsErrorInfo struct {
	Code string `json:"code"`
	Message string `json:"message"`
}

type RcsResponseErrorInfo struct {
	Status string `json:"status"`
	Error RcsErrorInfo `json:"error"`
}

type RcsResponseInfo struct {
	Status string `json:"status"`
	Data interface{} `json:"data"`
}

type RcsReasonInfo struct {
	Code string `json:"code"`
	Message string `json:"message"`
}

type RcsResultResponse struct {
	StatusInfo		[]RcsStatusInfo `json:"statusInfo,omitempty"`
	QueryId			string        	`json:"queryId,omitempty"`
	MoreToSend		string        	`json:"moreToSend,omitempty"`
}

type DHNResultStr struct {
	Statuscode int
	BodyData   []byte
	Result     map[string]string
}





