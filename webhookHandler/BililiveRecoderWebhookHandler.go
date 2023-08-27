package webhookHandler

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"webhookTemplate/messageSender"
)

func BililiveRecoderWebhookHandler(w http.ResponseWriter, request *http.Request) {
	// defer request.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(request.Body)
	// return 200 at first
	w.WriteHeader(http.StatusOK)
	// 读取请求内容
	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	log.Infof("收到webhook请求")
	log.Debugf(string(content))

	// 判断是否是重复的webhook请求
	webhookId := jsoniter.Get(content, "EventId").ToString()
	if webhookMessageIdList.IsContain(webhookId) {
		log.Warnf("重复的webhook请求：%s", webhookId)
		return
	} else {
		webhookMessageIdList.EnQueue(webhookId)
	}

	// 判断事件类型
	eventType := jsoniter.Get(content, "EventType").ToString()
	switch eventType {
	//录制开始 SessionStarted
	case "SessionStarted":
		log.Debugf("B站录播姬 录制开始 %s", jsoniter.Get(content, "EventData", "Name").ToString())
		break

	//文件打开 FileOpening
	case "FileOpening":
		log.Debugf("B站录播姬 文件打开 %s", jsoniter.Get(content, "EventData", "RelativePath").ToString())
		break

	//文件关闭 FileClosed
	case "FileClosed":
		log.Debugf("B站录播姬 文件关闭 %s", jsoniter.Get(content, "EventData", "RelativePath").ToString())
		break

	//录制结束 SessionEnded
	case "SessionEnded":
		log.Debugf("B站录播姬 录制结束 %s", jsoniter.Get(content, "EventData", "Name").ToString())
		break

	//直播开始 StreamStarted
	case "StreamStarted":
		log.Debugf("B站录播姬 直播开始 %s", jsoniter.Get(content, "EventData", "Name").ToString())
		var msg = messageSender.Message{
			Title: fmt.Sprintf("%s 开播了", jsoniter.Get(content, "EventData", "Name").ToString()),
			Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 开播时间：%s",
				jsoniter.Get(content, "EventData", "Name").ToString(),
				jsoniter.Get(content, "EventData", "Title").ToString(),
				jsoniter.Get(content, "EventData", "AreaNameParent").ToString(),
				jsoniter.Get(content, "EventData", "AreaNameChild").ToString(),
				jsoniter.Get(content, "EventTimestamp").ToString()),
		}
		err := msg.Send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		break

	//直播结束 StreamEnded
	case "StreamEnded":
		log.Debugf("B站录播姬 直播结束 %s", jsoniter.Get(content, "EventData", "Name").ToString())
		var msg = messageSender.Message{
			Title: fmt.Sprintf("%s 直播结束", jsoniter.Get(content, "EventData", "Name").ToString()),
			Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s",
				jsoniter.Get(content, "EventData", "Name").ToString(),
				jsoniter.Get(content, "EventData", "Title").ToString(),
				jsoniter.Get(content, "EventData", "AreaNameParent").ToString(),
				jsoniter.Get(content, "EventData", "AreaNameChild").ToString()),
		}
		err := msg.Send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		break

	//	别的不关心，所以没写
	default:
		log.Warnf("B站录播姬 未知的webhook请求：%+v", content)
	}
}
