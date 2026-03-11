package rcs

import "database/sql"

type MessageInfo struct {
	Common CommonInfo `json:"common, omitempty"`
	Rcs    RcsInfo    `json:"rcs, omitempty"`
}

type CommonInfo struct {
	Rcs_id           int
	Msg_id           string `json:"msgId, omitempty"`
	User_contact     string `json:"userContact, omitempty"`
	Schedule_type    int    `json:"scheduleType, omitempty"`
	Msg_group_id     string `json:"msgGroupId, omitempty"`
	Msg_service_type string `json:"msgServiceType, omitempty"`
}

type RcsInfo struct {
	Chatbot_id     string      `json:"chatbotId, omitempty"`
	Agency_id      string      `json:"agencyId, omitempty"`
	AgencyKey      string      `json:"agencyKey, omitempty"`
	BrandKey       string      `json:"brandKey, omitempty"`
	Messagebase_id string      `json:"messagebaseId, omitempty"`
	Service_type   string      `json:"serviceType, omitempty"`
	Expiry_option  int         `json:"expiryOption, omitempty"`
	Header         string      `json:"header, omitempty"`
	Footer         string      `json:"footer, omitempty"`
	Copy_allowed   bool        `json:"copyAllowed, omitempty"`
	Body           RcsBody     `json:"body, omitempty"`
	Buttons        []RcsButton `json:"buttons, omitempty"`
}

type RcsBody struct {
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	Media        string `json:"media,omitempty"`
	Title1       string `json:"title1,omitempty"`
	Description1 string `json:"description1,omitempty"`
	Media1       string `json:"media1,omitempty"`
	Title2       string `json:"title2,omitempty"`
	Description2 string `json:"description2,omitempty"`
	Media2       string `json:"media2,omitempty"`
	Title3       string `json:"title3,omitempty"`
	Description3 string `json:"description3,omitempty"`
	Media3       string `json:"media3,omitempty"`
	Title4       string `json:"title4,omitempty"`
	Description4 string `json:"description4,omitempty"`
	Media4       string `json:"media4,omitempty"`
	Title5       string `json:"title5,omitempty"`
	Description5 string `json:"description5,omitempty"`
	Media5       string `json:"media5,omitempty"`
}

type RcsButton struct {
	Suggestions []RcsSuggestions `json:"suggestions,omitempty"`
}

type RcsSuggestions struct {
	Action RcsAction `json:"action,omitempty"`
}

type RcsAction struct {
	UrlAction RcsUrlAction `json:"urlAction,omitempty"`
	/*
		MapAction 			RcsMapAction    	`json:"mapAction,omitempty"`
		CalendarAction 		RcsCalendarAction   `json:"calendarAction,omitempty"`
		ClipboardAction 	RcsClipboardAction  `json:"clipboardAction,omitempty"`
		ComposeAction 		RcsComposeAction    `json:"composeAction,omitempty"`
		DialerAction 		RcsDialerAction    	`json:"dialerAction,omitempty"`
		ShareAction 		RcsShareAction    	`json:"shareAction,omitempty"`
	*/
	DisplayText string      `json:"displayText,omitempty"`
	Postback    RcsPostback `json:"postback,omitempty"`
}

type RcsUrlAction struct {
	OpenUrl RcsOpenUrl `json:"openUrl,omitempty"`
}

type RcsOpenUrl struct {
	Url string `json:"url,omitempty"`
}

type RcsPostback struct {
	Data string `json:"data,omitempty"`
}

type RcsAuth struct {
	RcsId     string `json:"rcsId,omitempty"`
	RcsSecret string `json:"rcsSecret,omitempty"`
	GrantType string `json:"grantType,omitempty"`
}

type RcsAuthResp struct {
	Status string    `json:"status,omitempty"`
	Data   JsonData  `json:"data,omitempty"`
	Error  JsonError `json:"error,omitempty"`
}

type JsonData struct {
	TokenInfo JsonTokenInfo `json:"tokenInfo,omitempty"`
}

type JsonError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type RcsSendResp struct {
	Status string    `json:"status,omitempty"`
	Error  JsonError `json:"error,omitempty"`
}

type JsonTokenInfo struct {
	AccessToken string `json:"accessToken,omitempty"`
	ExpiresIn   string `json:"expiresIn,omitempty"`
}

type RcsResultReq struct {
	QueryType string `json:"queryType,omitempty"`
	MsgId     string `json:"msgId,omitempty"`
	RcsId     string `json:"rcsId,omitempty"`
	QueryId   string `json:"queryId,omitempty"`
	RetryYn   string `json:"retryYn,omitempty"`
}

type RcsResultStatus struct {
	RcsId       string    `json:"rcsId,omitempty"`
	MsgId       string    `json:"msgId,omitempty"`
	UserContact string    `json:"userContact,omitempty"`
	Status      string    `json:"status,omitempty"`
	ServiceType string    `json:"serviceType,omitempty"`
	MnoInfo     string    `json:"mnoInfo,omitempty"`
	SentTime    string    `json:"sentTime,omitempty"`
	Timestamp   string    `json:"timestamp,omitempty"`
	Error       JsonError `json:"error,omitempty"`
}

type RcsResultInfo struct {
	StatusInfo []RcsResultStatus `json:"statusInfo,omitempty"`
	QueryId    string            `json:"queryId,omitempty"`
	MoreToSend string            `json:"moreToSend,omitempty"`
}

type bulkUpdateItem struct {
	MsgId       string
	UserContact string
	Status      string
	ErrorMsg    sql.NullString
}

func GetRcsErrorMsg(code string) string {
	if msg, ok := RcsErrorMsg[code]; ok {
		return msg
	}
	return code
}

// RcsErrorMsg RCS 에러코드 메시지 매핑 (총 605개)
var RcsErrorMsg = map[string]string{
	"40001": "Missing Authorization header",
	"40002": "Missing token",
	"40003": "Invalid token",
	"40004": "Token has expired",
	"40005": "Malformed token payload",
	"40006": "Invalid client id",
	"40007": "Insufficient scope",
	"41000": "Internal Server Error",
	"41001": "RCS Request Timeout",
	"41002": "Revocation Failed",
	"41003": "Throttled by message rate",
	"41004": "RCS Server Busy",
	"41005": "RCS Server Temporarily unavailable",
	"41006": "Session does not exist",
	"41007": "Expired before session establishment",
	"41008": "Session already expired",
	"41009": "Device not support revocation",
	"41010": "IMDN received even already revoked",
	"41011": "Message was already revoked",
	"41100": "Connecting RCS Allocator Failed",
	"41101": "Connecting RCS Allocator Timeout",
	"41102": "RCS Allocation Failed",
	"41103": "RCS Allocation Timeout",
	"41104": "Connecting RCS Failed",
	"41105": "Connecting RCS Timeout",
	"41106": "Sending Message to RCS Failed",
	"41107": "Sending Message to RCS Timeout",
	"41108": "RCS handle request Failed",
	"41109": "RCS Internal Server Error",
	"41111": "Connecting RCS Presence Failed",
	"41112": "Connecting RCS Presence Timeout",
	"41113": "RCS Presence Failed",
	"41114": "RCS Presence Timeout",
	"41115": "MaaP Registry Failed",
	"41116": "MaaP Registry Timeout",
	"41117": "Process Revocation Request Failed",
	"41118": "File Upload Failed",
	"41200": "User Not Found. Non-RCS User",
	"41210": "User is not capable for TEXT",
	"41211": "User is not capable for FT",
	"41212": "User is not capable for RICHCARD",
	"41220": "User is not capable for XBOTMESSAGE 1.0",
	"41221": "User is not capable for XBOTMESSAGE 1.1",
	"41222": "User is not capable for XBOTMESSAGE 1.2",
	"41230": "User is not capable for OPENRICHCARD 1.0",
	"41231": "User is not capable for OPENRICHCARD 1.1",
	"41232": "User is not capable for OPENRICHCARD 1.2",
	"41240": "User is not capable for GEOLOCATION PUSH REQUEST",
	"41250": "Failed to get message content type",
	"41300": "File download failed",
	"42001": "Bot Not Found",
	"42002": "Missing MessageID",
	"42003": "Missing BotID",
	"42004": "Missing User Contact",
	"42005": "User Contact did not accepted",
	"42006": "Emulator access only",
	"42007": "Contact not in allowlist",
	"42008": "Missing message content",
	"42009": "Invalid message content",
	"42010": "Ambiguous message",
	"42011": "Invalid message status",
	"42012": "Invalid isTyping status",
	"42013": "Invalid traffic type",
	"42014": "Invalid suggested chiplist association",
	"42015": "Empty text message",
	"42031": "Missing richcard",
	"42032": "Ambiguous richcard",
	"42033": "Too many richcards",
	"42034": "Missing richcard layout",
	"42035": "Missing richcard content",
	"42036": "Invalid cardOrientation",
	"42037": "Missing image alignment",
	"42038": "Invalid image alignment",
	"42039": "Redundant image alignment",
	"42040": "Invalid richcard carousel cardWidth",
	"42041": "Mismatched media height",
	"42042": "Invalid richcard content",
	"42043": "Invalid suggestions",
	"42044": "Too many suggestions in chiplist (max 11)",
	"42045": "Invalid suggestion",
	"42046": "Ambiguous suggestion",
	"42047": "Ambiguous suggested action",
	"42048": "Too many suggestions in richcard (max 4)",
	"42049": "Too long postback data (max 2048)",
	"42050": "Invalid DisplayText in reply or action (max25)",
	"42051": "Invalid location",
	"42052": "Ambiguous location",
	"42053": "Invalid mapAction",
	"42054": "Ambiguous mapAction",
	"42055": "Invalid dialerAction",
	"42056": "Ambiguous dialerAction",
	"42057": "Invalid composeAction",
	"42058": "Ambiguous composeAction",
	"42059": "Invalid settingsAction",
	"42060": "Ambiguous settingsAction",
	"42061": "Invalid ClipboardAction",
	"42062": "Missing localBrowserAction",
	"42063": "Invalid ShareAction",
	"42064": "Missing CalendarAction",
	"42065": "Invalid CalendarAction",
	"42066": "Invalid CalendarAction title (min1 max100)",
	"42067": "Invalid CalendarAction description (min1 max500)",
	"42068": "Invalid ComposeAction composeTextMessage (missing phone number) or (text min 1 max 100)",
	"42069": "Invalid ComposeAction composeRecodingMessage (missing phone number) or (type 'AUDIO' or 'VIDEO')",
	"42070": "Missing DeviceAction",
	"42071": "Invalid dialerAction (missing phone number) or (subject max 60)",
	"42072": "Missing mapAction showLocation location",
	"42073": "Missing urlAction openUrl",
	"42100": "Invalid thumbnail",
	"42101": "Missing fileUrl",
	"42102": "Missing audio fileUrl",
	"42103": "Missing pos",
	"42104": "Missing Media information",
	"42105": "Missing Media Content Type",
	"42106": "Missing Media FileSize",
	"42107": "Missing Media Height",
	"42108": "Missing postback data",
	"42200": "Invalid URI format",
	"42201": "Invalid Location Label",
	"42202": "Invalid media content description",
	"42203": "Invalid media title",
	"42204": "Invalid media description",
	"42205": "Too many contents in richard carousel",
	"42206": "Invalid media file size",
	"42207": "Invalid expiry format",
	"42208": "Invalid expiry",
	"42301": "Missing Openrichcard layout",
	"42302": "Missing Openrichcard",
	"42303": "Missing Openrichcard Layout Widget",
	"42304": "Invalid Openrichcard View Content",
	"42305": "Invalid Openrichcard LinearLayout Content",
	"42306": "Missing Openrichcard Textview Content",
	"42307": "Invalid Contents In Openrichcard TextView",
	"42308": "Invalid Text Length In Openrichcard TextView",
	"42309": "Missing Openrichcard ImageView Content",
	"42310": "Too Small Media In Openrichcard ImageView",
	"42311": "Invalid Openrichcard ImageView Scaletype",
	"42312": "Missing Openrichcard width or height",
	"42313": "Invalid Openrichcard width or height",
	"42314": "Invalid Openrichcard Common Contents",
	"42315": "Too many child in open rich card",
	"42316": "Missing Openrichcard Click for Button",
	"42401": "Invalid file type",
	"42402": "Download failure",
	"42501": "Missing contact",
	"42502": "Missing content",
	"42503": "Missing title",
	"42504": "Missing description",
	"42505": "Missing image url",
	"42506": "Missing image type",
	"42507": "Missing button link",
	"42508": "Missing button text",
	"42509": "Invalid title",
	"42510": "Invalid description",
	"42511": "Invalid image url address",
	"42512": "Invalid button url address",
	"42513": "Invalid button text",
	"42514": "Duplicated message ID",
	"42601": "Too Many Request",
	"42602": "message filtered",
	"45000": "Internal Server Error",
	"45001": "Handle bot data error",
	"45002": "Handle capability data error",
	"45003": "Handle xml data error",
	"45004": "Access MessageInfo Cache Failed",
	"45005": "Access ChatInfo Cache Failed",
	"45006": "Create Http client for maap registry",
	"45007": "Create Http client for bot",
	"50000": "Matched to New Specifications",
	"50001": "Missing Authorization header",
	"50002": "Missing token",
	"50003": "Invaild token",
	"50004": "Token has expired",
	"50005": "Authorization Error",
	"50006": "Invalid client id",
	"50007": "Invalid sender id",
	"50008": "Invalid password",
	"50009": "NonAllowedIp",
	"50100": "Invalid state",
	"50201": "Message TPS Exceeded",
	"50202": "Message Quota Exceeded",
	"51003": "DuplicationError",
	"51004": "ParameterError",
	"51005": "JsonParsingError",
	"51006": "DataNotFound",
	"51007": "Duplicated AutoReplyMsgId",
	"51008": "Duplicated PostbackId",
	"51101": "AgencyKey Not Found",
	"51102": "Invaild AgencyKey",
	"51103": "BrandKey Not Found",
	"51104": "Invaild BrandKey",
	"51900": "Not Found Handler",
	"51901": "Samsung MaaP Connect IF Error",
	"51902": "Samsung MaaP Service IF Error",
	"51903": "Capri IF Error",
	"51904": "Webhook Execute Imposible Status Failure",
	"51905": "Webhook CDR Log Writing Failure",
	"51906": "Invalid Webhook Url",
	"51907": "Expired Webhook Message",
	"51908": "Exceed Retry Count to Send Message",
	"51909": "Non Existing Webhook Message",
	"51910": "Non Existing Webhook GW Vendor",
	"51911": "No Subscription",
	"51912": "Failure in Preperation for sending Webhook",
	"51913": "Failure in Updating Webhook Sending Result State",
	"51914": "Failure in Updating CDR Result State",
	"51915": "Failure in Updating Completion Result State",
	"51916": "Failure in Updating Expiration Result State",
	"51917": "Webhook Msg Log Creation Error",
	"51918": "Webhook Msg Log Creation Failure",
	"51919": "Invalid Webhook Receive Request Error",
	"51920": "Non Existing Chatbot For Request",
	"51921": "Non Usable Chatbot State Error",
	"51922": "Non Existing Gw Vendor For Request",
	"51923": "Non Definition Mo Message Url for Chatbot",
	"51924": "Webhook Gateway Execution Error",
	"51925": "Not Allowed Request Event Error",
	"51926": "Webhook Receive Execution Error",
	"51927": "Webhook Receive Async Execution Error",
	"51928": "Non Definition Webhook Url for Gw Vendor Cid",
	"51929": "Non Target GwVendor for CDR Log",
	"51930": "Non Received Webhook Command Error Log",
	"51931": "Mo Message Registration Error",
	"51932": "Mo Message Registration Failuer",
	"51933": "Auto Reply Message Sending Error",
	"51934": "Non Service Supported Error",
	"51935": "Samsung MaaP Core File Server Connection Error",
	"51936": "Message FileMessage Event File Download Error",
	"51937": "Message FileMessage Event File Download Error",
	"51938": "Message FileMessage Registration to DB Error",
	"51939": "Message FileMessage Registration to DB Failure",
	"51950": "Webhook Scheduler Async Execution Error",
	"51951": "Webhook Scheduler DB Execution Error",
	"51952": "Webhook Scheduler DB Execution Failuer",
	"51953": "Webhook Scheduler Process Execution Error",
	"51954": "Webhook Scheduler Process Execution Failuer",
	"51955": "Webhook Scheduler Processor Execution Error",
	"51956": "Webhook Scheduler Processor Execution Failuer",
	"51957": "Non Existing Mo Message Error",
	"51958": "Non-Defined Webhook Scheduler Type Error",
	"52001": "Invalid phone number format",
	"52002": "Invalid message status",
	"52003": "Bot Aleady Exists",
	"52004": "Bot Creation Failed",
	"52005": "Bot Update Failed",
	"52006": "Brand Delete Failed",
	"52007": "InvalidChatbotServiceType",
	"52008": "MismatchedChatbotId",
	"52009": "Persistent Menu Permission Error",
	"52010": "Invalid Persistent Menu Data",
	"52016": "Message Transmission Time Exceeding",
	"52023": "Messagebase Id Stopped Temporarily",
	"52101": "Invalid Webhook Request Parameter",
	"52102": "Webhook Host Connect Error",
	"52103": "Webhook Host Server Request Failure",
	"52104": "Webhook Response Receive Failure",
	"52105": "Webhook Message Non Receive Failure",
	"52106": "Webhook Message Process Failure",
	"52107": "Invalid Webhook Msg A2P Status",
	"52108": "Response Event Postback Stat Log DB Error",
	"52109": "Response Event Postback Stat Log DB Failuer",
	"52201": "Invalid Auto Reply Message ID",
	"52202": "Auto Reply Message Contents Error",
	"53001": "Invalid file type",
	"53002": "File Attribute Error",
	"53003": "FileID format Error",
	"53004": "FileUploadError",
	"53005": "InvalidMultiPartRequest",
	"53006": "Attached File Size Error",
	"53007": "File Info Extraction Error",
	"53008": "File Size Format Error",
	"53009": "Validation Error for OpenRichCard MMS Msg",
	"53010": "Invalid OpenRichCard MMS Msg File",
	"53011": "Invalid OpenRichCard MMS Msg File Size",
	"53012": "Invalid OpenRichCard MMS Media IframeplayB Type Error",
	"53013": "OpenRichCard MMS IfameplayB Media Building Error",
	"53014": "Non Existing OpenRichCard MMS MessageBase",
	"53015": "Non Existing OpenRichCard MMS MessageBase Parameters",
	"53016": "Non Existing OpenRichCard MMS Media Data",
	"53017": "Mismated OpenRichCard MMS Media File MessagebaseId",
	"53018": "UnSupported OpenRichCard Product MediaStreaming Error",
	"54001": "Invaild Contact number User",
	"54002": "No Rcs Capability",
	"54003": "Unable Sending to Recipient",
	"54004": "MaaP Internal Error",
	"55001": "Corp Content Error",
	"55002": "Invalid Property",
	"55101": "Agency Content Error",
	"55102": "Invalid AgencyId",
	"55103": "AgencyID permission error",
	"55104": "Contract Content Error",
	"55201": "Brand Content Error",
	"55202": "Brand name Error",
	"55203": "Brand profile image Error",
	"55204": "Brand CS number Error",
	"55205": "Brand menu Error",
	"55206": "Brand category Error",
	"55207": "Brand homepage Error",
	"55208": "Brand email Error",
	"55209": "Brand address Error",
	"55210": "Invalid BrandID",
	"55301": "Bot Content Error",
	"55302": "Invalid BotID",
	"55303": "BotID permission error",
	"55501": "Messagebase Content Error",
	"55502": "Invalid MessagebaseID",
	"55503": "MessagebaseID permission error",
	"55504": "Invalid formatstring",
	"55505": "Invalid messagebase Policy Info",
	"55506": "Invalid messagebase param",
	"55507": "Invalid messagebase attribute",
	"55508": "Invalid messagebase type",
	"55509": "mismatching Product type",
	"55510": "Non Exists Messagebase Policy",
	"55511": "Non Exists Messagebase Param",
	"55512": "Invalid Messagebase Status",
	"55513": "Invalid Message Product Version",
	"55601": "MessagebaseForm Content Error",
	"55602": "Invalid messagebaseformID",
	"55603": "Invalid MessageBase ProductCode",
	"55701": "prohibited text content",
	"55702": "Action button Permission error",
	"55703": "prohibited header value",
	"55704": "prohibited footer field",
	"55705": "missing footer content",
	"55706": "footer content syntax Error",
	"55707": "content pattern error",
	"55708": "Exceeded max character of title",
	"55709": "Exceeded max character of description",
	"55710": "Exceeded max number of buttons",
	"55711": "mismatching number of carousel card",
	"55712": "Exceeded max size of media",
	"55714": "Expired ReplyId",
	"55715": "Not Found ReplyId",
	"55716": "Msg Send Request Invald Error",
	"55717": "Non Defined Bot Agency Id for Chat Service",
	"55718": "Mismatching AgencyId for Chat Service",
	"55719": "Non Existing Gw Vendor/Cid Info",
	"55720": "Mismatching Request Gw Vendor Cid for Chat Service",
	"55721": "Duplication Gw Vendor Cid",
	"55722": "Exceeded Individual Size Of Media",
	"55723": "Invalid Ambi-Directional Data Status",
	"55730": "RCS Message Validation Check Error",
	"55731": "Invalid OpenRichCard MMS Request Policy",
	"55732": "Non-CouplingId Value for OpenRichCard MMS Param",
	"55733": "Non-Parameter for OpenRichCard MMS Param",
	"55734": "Invalid OpenRichCard MMS Coupled Param Value",
	"55735": "OpenRichCard MMS Visiblity Value Setting Failure",
	"55736": "Invalid OpenRichCard Media Url Format Error",
	"55737": "OpenRichCard MMS MaapUrl translate to MediaUrl Error",
	"55738": "OpenRichCard MMS Message Format Error",
	"55739": "OpenRichCard MMS Media Url Building Failure",
	"55740": "OpenRichCard MMS Media Url Building Error",
	"55741": "OpenRichCard MMS Button Non Existance Error",
	"55742": "Not Allowed Parameter for OpenRichCard",
	"55743": "Invalid OpenRichCard MMS Optional Param Policy",
	"55744": "OpenRichCard MMS Content Building Failure",
	"55745": "OpenRichCard MMS Parameter Validation Error",
	"55801": "GwVendor Content Error",
	"55802": "Invalid message",
	"55803": "Message Syntax Error",
	"55804": "Missing message content",
	"55805": "Invalid message content",
	"55806": "Duplicated MessageId",
	"55807": "Invalid Chatbot Permission",
	"55808": "Invalid Chatbot Status",
	"55809": "Invalid Agency Permission",
	"55810": "Invalid Expiry Field",
	"55811": "Exceeded Max Character Of Param",
	"55812": "Buttons Not Allowed",
	"55813": "Exceed Button Text Length",
	"55814": "Invalid MessageBase Buttons",
	"55815": "Message Body File Not Found",
	"55816": "Mismatched Suggestions Count",
	"55817": "Invalid Dest Phone Number",
	"55818": "Invalid MessageBase Id",
	"55819": "Invalid Chatbot Id",
	"55820": "Revoked Message",
	"55821": "Etc TimeOut",
	"55822": "Canceled Message",
	"55823": "Mismatched suggestedChipList Count",
	"55824": "Chiplist Not Allowed",
	"55825": "Buttons Reply Not Allowed",
	"55880": "Mismatched CarouselButton Count",
	"55881": "Mismatched ChipList Count",
	"55882": "ChipList Not Allowed",
	"55883": "Invalid ReplyId",
	"55884": "Missing ReplyId",
	"55885": "Invalid User Contact For SessionMessage",
	"55886": "Invalid ChatbotId For Session Message",
	"55887": "Invalid MessageBase ProductCode For Session Message",
	"55888": "Not Allowed Chatbot For Session Message",
	"55900": "Non Retryable Error Caused By Invalid Message",
	"56002": "Revocation Failed",
	"56007": "Expired Before Session Establishment",
	"57001": "Exceeded Number Of Limit",
	"57002": "Invalid Offset Number",
	"57003": "Invaild Stat Type Error",
	"59001": "SystemError",
	"59002": "IOError",
	"59003": "Backend Timeout",
	"59999": "Etc Error",
	"60004": "No Content",
	"61001": "Missing Authorization header",
	"61002": "Missing Token",
	"61003": "Invalid token",
	"61004": "Token has expired",
	"61005": "Invalid client id",
	"61006": "Invalid secret key",
	"63001": "No Brand Permission",
	"64001": "Missing X-RCS-BrandKey header",
	"64002": "Invalid Brand Key",
	"64101": "Invalid brandId on path parameter",
	"64102": "Invalid agencyId on path parameter",
	"64103": "Invalid corpRegNum on path parameter",
	"64104": "Invalid personId on path parameter",
	"64105": "Invalid chatbotId on path parameter",
	"64106": "Invalid messagebaseId on path parameter",
	"64107": "Invalid messagebaseformId on path parameter",
	"64201": "Invalid query parameter ({name of paramter})",
	"64202": "Invalid query parameter value ({name of paramter}:{value})",
	"64203": "query paramter required ({name of paramter})",
	"64301": "Missing Body data",
	"64302": "Invalid JSON format",
	"64303": "Invalid type of Attribute ({name  of attribute})",
	"64304": "Over specified size ({name of attribute})",
	"64305": "Missing Certification document",
	"64306": "Exceed MDN registration quantity",
	"64307": "Missing MDN",
	"64308": "Invalid MDN format",
	"64309": "Missing chatbot name",
	"64310": "Invalid display format",
	"64311": "Invalid smsmo format",
	"64312": "Missing messagebaseformId",
	"64313": "Invalid messagebaseformId",
	"64314": "Missing Template name",
	"64315": "Missing brandId",
	"64316": "Invalid brandId",
	"64317": "Invalid agencyId",
	"64318": "Invalid formattedString format",
	"70001": "Missing Authorization header",
	"70002": "Missing token",
	"70003": "Invaild token",
	"70004": "Token has expired",
	"70005": "Authorization Error",
	"70006": "Invalid client id",
	"70007": "Invalid sender id",
	"70008": "Invalid password",
	"70009": "Non allowed ip",
	"71001": "System Error",
	"71002": "IOError",
	"71003": "Duplication Error",
	"71004": "Parameter Error",
	"71005": "Json Parsing Error",
	"71006": "Data Not Found",
	"71007": "Not Found Handler",
	"71008": "Missing Mandatory Parameter",
	"71009": "Invalid Data State",
	"71010": "MAAP-FE API Error",
	"71011": "Kisa GW Error",
	"71816": "Whloesale Simulater Fail",
	"72100": "Invalid State",
	"72101": "Common Info Error",
	"72102": "Legacy Info Error",
	"72103": "Rcs Info Error",
	"72104": "Message TPS Exceeded",
	"72106": "Message Media Not Found",
	"72107": "Invalid Media File Type",
	"72108": "Invalid Message Type",
	"72109": "Invalid Content Type",
	"72110": "Exceeded number of contents",
	"72111": "Data Type Error",
	"72112": "Attached File Error",
	"72113": "File Upload Error",
	"72114": "Convert Timeout",
	"72115": "Attached file size Error",
	"72116": "Webfile Download Error",
	"72117": "Exceeded message content size",
	"72118": "Sub Type Error",
	"72119": "Invalied Data Type",
	"72120": "Message Format Error",
	"72121": "Exceed Content Size",
	"72122": "Invalid Chatbot Permission",
	"72123": "Invalid Service Type",
	"72124": "Invalid Service Permission",
	"72125": "Invalid Expiry Option",
	"72126": "Invalid Header",
	"72127": "Invalid Footer",
	"72128": "MessageSMSQuotaExceeded",
	"72129": "MessageLMSQuotaExceeded",
	"72130": "MessageMMSQuotaExceeded",
	"72131": "MessageTMPLTQuotaExceeded",
	"72132": "Message CHAT Quota Exceeded",
	"72133": "Message ITMPLT Quota Exceeded",
	"72141": "Chiplist Not Allowed",
	"72142": "Invalid Chatbot Chat Permission",
	"72143": "Not Found Corp Chat",
	"72144": "Invalid Chat Message Permission",
	"72145": "Not Found ReplyId",
	"72146": "Expired ReplyId",
	"73001": "Unavailable Service",
	"73002": "UnsubscribedLegacyService",
	"74001": "Message Content Spam",
	"74002": "Sender Number Spam",
	"74003": "Receipient Number Spam",
	"74004": "Callback Number Spam",
	"74005": "Same Message Limit",
	"74006": "Same Receipient Number Limit",
	"74007": "Number Theft Block",
	"74008": "Callback Number Block",
	"74009": "Number Rule Violation Block",
	"74101": "080 Spam Block Number",
	"75001": "NPDB user (send fail)",
	"75002": "There is no subscriber",
	"75003": "CAPRI & NPDB Error",
	"75004": "There is no end user",
	"75005": "Receipient's number Error",
	"75006": "Sender number Error",
	"75007": "No Rcs Capability",
	"75008": "Invalid User Contact",
	"76001": "Capri interface Error",
	"76002": "Schedule Manager Internal Error",
	"76003": "Not Found Rcs Subscriber",
	"76004": "Xroshot Sender Internal Error",
	"76005": "Xroshot Manager Internal Error",
	"77001": "Expired message received time",
	"77002": "Invalied Message Sequence",
	"77003": "Non Existing Webhook Message",
	"77004": "Invalid Webhook Message",
	"77005": "Non Existing Webhook Corporation",
	"77006": "Webhook Msg Log Writing Failure",
	"77007": "Disabled Send Webhook Msg",
	"77008": "Webhook Failure Status",
	"77009": "Invalid Webhook Message",
	"77010": "Invalid Webhook EventType",
	"77011": "Invalid Webhook Message Status",
	"77012": "Not Found Agency",
	"77013": "Non Existing MO Message",
	"77701": "Non Existing Webhook Message",
	"77703": "Non Existing Webhook GW Vendor",
	"77704": "Invalid Webhook GW Vendor",
	"77705": "Webhook CDR Log Writing Failure",
	"77707": "Failure in updating Webhook Message",
	"77708": "Invalid Webhook Request Parameter",
	"77709": "Webhook Host Connect Error",
	"77710": "Webhook Host Server Request Failure",
	"77711": "Webhook Response Receive Failure",
	"77712": "Invalid Webhook Url",
	"77713": "Expired Webhook Message",
	"77714": "Exceed Retry Count to Send Message",
	"77715": "Failure in Sending Message to MaaP Core FE",
	"77716": "Failure in Linkage to BrandPortal Api Server",
	"77719": "NonExistingResultMsgRcsMapping",
	"77720": "NonExistingResultMsg",
	"77721": "ResultMsgRegisterFailuer",
	"77722": "Mo Message Registration Error",
	"77723": "Maap-FE File Server Connection Error",
	"77724": "Message FileMessage Event File Donwload Error",
	"77725": "Message FileMessage Registration to DB Error",
	"77726": "Failure FileMessage Registration Error",
	"77727": "File Download Error",
	"77728": "Mo Reply Registration Error",
	"77800": "Non CDR Target Message",
	"77801": "Invalid Notification Param Error",
	"77802": "Brand Portal Auth Token Error",
	"77803": "Brand Portal Host Connection Failure",
	"77804": "Brand Portal Auth Token Request Failure",
	"77805": "Brand Portal Auth Token Response Failure",
	"77806": "Brand Portal Api Request Format Error",
	"77807": "Brand Portal UnAuthorized Token Error",
	"77808": "Brand Portal Api Request Error",
	"77809": "Brand Portal Api Response Error",
	"77810": "Brand Portal Api Processing Error",
	"77811": "Brand Portal Api No Data",
	"77812": "Notification Internal Error",
	"77813": "Failure Notification Hist Insert",
	"77814": "Invalid Notification Method",
	"77815": "Invalid Notificiation Type",
	"78001": "Unsupport Error",
	"78002": "Calling",
	"78003": "Device No response",
	"78004": "Device Power Off",
	"78005": "Shaded Area",
	"78006": "Device Message Full",
	"78007": "SMS Forward Exceed",
	"78008": "Invalid Subscriber Name",
	"78009": "No CallbackUrl User",
	"78010": "Invalid Device Error",
	"79001": "Retry Count Exceeded",
	"79002": "Concurrent Max Request Exceeded",
	"79003": "RCS ID Mismatched",
	"79004": "Invalid Query Request",
	"79005": "Exceed Query Limit Count",
	"79006": "Request Query Execution Failure",
	"79007": "Duplicated Query Id",
	"79008": "InvaildRequestError",
	"79009": "Reserved",
	"79010": "Invalid Result Policy",
	"79011": "NonExistQueryId",
	"79994": "KT MaaP-FE Fail",
	"79995": "SKT MaaP-FE Fail",
	"79996": "LGU MaaP-FE Fail",
	"79997": "TimeOut",
	"79998": "Etc TimeOut",
	"79999": "Unknown Error",
}
