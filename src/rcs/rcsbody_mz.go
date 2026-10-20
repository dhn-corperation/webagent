package rcs

// 💡 메가버드(Megabird) RCS 발송 요청 구조체
type MegabirdRcsRequest struct {
	MessageBaseID       string                 `json:"messageBaseId"`
	BrandID             string                 `json:"brandId"`
	RcsMsgType          string                 `json:"rcsMsgType"`
	RcsContent          MegabirdRcsContent     `json:"rcsContent"`
	SnPhnum             string                 `json:"snPhnum"`
	MessageReceiverList []MegabirdReceiver     `json:"messageReceiverList"`
	AdIncluYn           string                 `json:"adIncluYn"`
	ExpiryOpt           string                 `json:"expiryOpt"`
	UseDefault080       string                 `json:"useDefault080,omitempty"`
	BlockedPhnum        string                 `json:"blockedPhnum,omitempty"`
	CopyYn              string                 `json:"copyYn,omitempty"`
	SubTxtUseYn         string                 `json:"subTxtUseYn"`
	MsgMetaData         map[string]interface{} `json:"msgMetaData,omitempty"`
}

type MegabirdRcsContent struct {
	MsgData []MegabirdMsgData `json:"msgData"`
}

type MegabirdMsgData struct {
	Title       string                   `json:"title,omitempty"`
	Description string                   `json:"description,omitempty"`
	Buttons     []map[string]interface{} `json:"buttons,omitempty"`
}

type MegabirdReceiver struct {
	Mbnum string `json:"mbnum"`
}

// 💡 메가버드(Megabird) 발송 응답 구조체
type MegabirdRcsResponse struct {
	Success   bool        `json:"success"`
	Code      interface{} `json:"code"`
	Message   string      `json:"message"`
	Data      struct {
		MsgGrpNo int64 `json:"msgGrpNo"`
	} `json:"data"`
	Timestamp string `json:"timestamp"`
	// 💡 400 에러 상세 사유 파싱용
	Description []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"description"`
}

// 💡 KT 형식의 Body JSON 파싱용 임시 구조체
type KtBodyStruct struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// 메가버드 결과 조회용 응답 구조체
type MegabirdResultResponse struct {
	Code   string `json:"code"`
	Result struct {
		TotalCount int  `json:"totalCount"`
		HasNext    bool `json:"hasNext"`
		LastSeenId string `json:"lastSeenId"`
		Result     []struct {
			MsgGrpNo    int64  `json:"msgGrpNo"`
			MessageId   string `json:"messageId"`
			Mblnum      string `json:"mblnum"`
			SndStsCd    string `json:"sndStsCd"`    // MTC(성공), MTF(실패)
			SndRstCd    string `json:"sndRstCd"`    // 에러코드
			SndStsCzCmt string `json:"sndStsCzCmt"` // 에러 상세
		} `json:"result"`
	} `json:"result"`
}