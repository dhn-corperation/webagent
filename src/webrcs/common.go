package webrcs

import (
	"strings"
	"encoding/json"

	"webagent/src/config"
)

var rcsAuthRequest *RcsAuthRequest

func SetAuthRequest() {
	rcsAuthRequest = &RcsAuthRequest{
		RcsId : config.RCSID,
		RcsSecret : config.RCSPW,
		GrantType : "clientCredentials",
	}
}

func (rcsAuthRequest *RcsAuthRequest) getTokenInfo() (RcsAuthResponse, error) {

	resp, err := config.Client.R().
		SetHeaders(map[string]string{"Content-Type": "application/json"}).
		SetBody(rcsAuthRequest).
		Post(config.Conf.RCSSENDURL + "/corp/v1/token")

	var rcsAuthResponse RcsAuthResponse
	if err != nil {
		return rcsAuthResponse, err
	} else {
		json.Unmarshal(resp.Body(), &rcsAuthResponse)
		return rcsAuthResponse, nil
	}
}

func getQuestionMark(column []interface{}) string {
	var placeholders []string
	numPlaceholders := len(column) // 원하는 물음표 수
	for i := 0; i < numPlaceholders; i++ {
	    placeholders = append(placeholders, "?")
	}
	return strings.Join(placeholders, ",")
}