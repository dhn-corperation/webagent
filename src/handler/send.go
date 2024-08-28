package handler

import(
	"time"
	s "strings"

	"github.com/gin-gonic/gin"
)

func SendNano(c *gin.Context){
	uid := c.Query("uid")
	sd := c.Query("sd")
	uip := c.ClientIP()
	if s.EqualFold(uid, "dhn") {
		ed := time.Now().Format("2006-01-02 15:04:05")
		parseSd, err := time.Parse("20060102150405", sd)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    "error",
				"message": "잘못된 시간형식 입니다.",
				"sd":  sd,
			})
		}
		formattedSd := parseSd.Format("2006-01-02 15:04:05")
		nanoResendProcess(formattedSd, ed)
		loopSd := ed
		for i:=0;i<3;i++ {
			time.Sleep(1 * time.Minute)
			ed = time.Now().Format("2006-01-02 15:04:05")
			nanoResendProcess(loopSd, ed)
			loopSd = ed
		}

		c.JSON(200, gin.H{
			"code":    "ok",
			"message": "처리가 완료되었습니다.",
			"sd":  sd,
			"ed":  ed,
		})

	} else {
		c.JSON(404, gin.H{
			"code":    "error",
			"message": "허용되지 않은 사용자 입니다",
			"userid":  uid,
			"ip":      uip,
		})
	}
}

func nanoResendProcess(sd, ed string){

}