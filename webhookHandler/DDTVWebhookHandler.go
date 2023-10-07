package webhookHandler

import (
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
	"webhookTemplate/bilibiliInfo"
	"webhookTemplate/messageSender"
)

// ddtvTaskRunner 根据响应体内容，执行任务
func ddtvTaskRunner(content []byte) {
	log.Trace(string(content))
	webhookId := jsoniter.Get(content, "id").ToString()
	log.Infof("%s 收到 DDTV webhook 请求", webhookId)

	// 判断是否是重复的webhook请求
	webhookMessageIdListLock.Lock()
	if webhookMessageIdList.IsContain(webhookId) {
		webhookMessageIdListLock.Unlock()
		log.Warnf("%s 重复的webhook请求", webhookId)
		return
	} else {
		webhookMessageIdList.EnQueue(webhookId)
		webhookMessageIdListLock.Unlock()
	}

	// 判断事件类型
	hookType := jsoniter.Get(content, "type").ToInt()
	switch hookType {
	//	0 StartLive 主播开播
	case 0:
		// 输出日志
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 主播开播：")
		logBuilder.WriteString(jsoniter.Get(content, "user_info", "name").ToString())
		log.Info(logBuilder.String())
		// 构建消息
		// 构造消息标题
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "user_info", "name").ToString())
		msgTitleBuilder.WriteString(" 开播了")
		// 构造消息内容
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_name").ToString())
		msgContentBuilder.WriteString("\n- 开播时间：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "hook_time").ToString())
		// 发送消息
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
			ID:      webhookId,
			IconURL: jsoniter.Get(content, "user_info", "face").ToString(),
		}
		msg.Send()
		break

	//	1 StopLive 主播下播
	case 1:
		// 输出日志
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 主播下播：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Info(logBuilder.String())
		/*// 构造消息
		// 构造消息标题
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		// 封禁检测
		isLocked := jsoniter.Get(content, "room_Info", "is_locked").ToBool()
		if isLocked {
			// 主播被封号了
			msgTitleBuilder.WriteString(" 喜提直播间封禁！")
		} else {
			// 主播正常下播
			msgTitleBuilder.WriteString(" 下播了")
		}
		// 构造消息内容
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_name").ToString())
		if isLocked {
			msgContentBuilder.WriteString("\n- 封禁到：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "lock_till").ToString())
		}
		// 发送消息
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
			ID: webhookId,
			IconURL: jsoniter.Get(content, "user_info", "face").ToString(),
		}
		msg.Send()*/
		break

	//	2 StartRec 开始录制
	case 2:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 开始录制：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Info(logBuilder.String())
		break

	//	3 RecComplete 录制结束
	case 3:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 录制结束：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Info(logBuilder.String())
		break

	//	4 CancelRec 录制被取消
	case 4:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 录制被取消：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Info(logBuilder.String())
		break

	//	5 TranscodingComplete 完成转码
	case 5:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 完成转码：")
			logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
			log.Debug(logBuilder.String())
		}
		break

	//	6 SaveDanmuComplete 保存弹幕文件完成
	case 6:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 保存弹幕文件完成：")
			logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
			log.Debug(logBuilder.String())
		}
		break

	//	7 SaveSCComplete 保存SC文件完成
	case 7:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 保存SC文件完成：")
			logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
			log.Debug(logBuilder.String())
		}
		break

	//	8 SaveGiftComplete 保存礼物文件完成
	case 8:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 保存礼物文件完成：")
			logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
			log.Debug(logBuilder.String())
		}
		break

	//	9 SaveGuardComplete 保存大航海文件完成
	case 9:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 保存大航海文件完成：")
			logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
			log.Debug(logBuilder.String())
		}
		break

	//	10 RunShellComplete 执行Shell命令完成
	case 10:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 执行Shell命令完成：")
			logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
			log.Debug(logBuilder.String())
		}
		break

	//	11 DownloadEndMissionSuccess 下载任务成功结束
	case 11:
		// 记录日志
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 下载任务成功结束：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Info(logBuilder.String())

		// 构造消息
		// 构造消息内容
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_name").ToString())
		// 构造消息标题
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		// 判断是否是封禁
		isRoomLocked, lockTill := bilibiliInfo.IsRoomLocked(jsoniter.Get(content, "room_Info", "room_id").ToUint64(), webhookId)
		if isRoomLocked {
			// 主播被封号了
			msgTitleBuilder.WriteString(" 喜提直播间封禁！")
			msgContentBuilder.WriteString("\n- 封禁到：")
			msgContentBuilder.WriteString(time.Unix(lockTill, 0).Local().Format("2006-01-02 15:04:05"))
		} else {
			msgTitleBuilder.WriteString(" 录制完成")
		}
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
			ID:      webhookId,
			IconURL: jsoniter.Get(content, "user_info", "face").ToString(),
		}
		msg.Send()
		break

	//	12 SpaceIsInsufficientWarn 剩余空间不足
	case 12:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 剩余空间不足：")
		logBuilder.Write(content)
		log.Warn(logBuilder.String())
		break

	//	13 LoginFailure 登陆失效
	case 13:
		log.Errorf("%s DDTV 登录失效", webhookId)
		break

	//	14 LoginWillExpireSoon 登陆即将失效
	case 14:
		log.Warnf("%s DDTV 登录即将失效", webhookId)
		break

	//	15 UpdateAvailable 有可用新版本
	case 15:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 有可用新版本：")
		logBuilder.WriteString(jsoniter.Get(content, "version").ToString())
		log.Info(logBuilder.String())
		break

	//	16 ShellExecutionComplete 执行Shell命令结束
	case 16:
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" DDTV 执行Shell命令结束：")
			logBuilder.Write(content)
			log.Debug(logBuilder.String())
		}
		break

	//	以下是自编译版本的特有内容

	//	17 WarnedByAdmin 被管理员警告
	case 17:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 被管理员警告：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Warn(logBuilder.String())
		// 构造消息
		// 构造消息标题
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgTitleBuilder.WriteString(" 被管理员警告")
		// 构造消息内容
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_name").ToString())
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
			ID:      webhookId,
			IconURL: jsoniter.Get(content, "user_info", "face").ToString(),
		}
		msg.Send()
		break

	//	18 LiveCutOff 直播被管理员切断
	case 18:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 直播被管理员切断：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Warn(logBuilder.String())
		// 构造消息
		// 构造消息标题
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgTitleBuilder.WriteString(" 直播被管理员切断")
		// 构造消息内容
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_name").ToString())
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
			ID:      webhookId,
			IconURL: jsoniter.Get(content, "user_info", "face").ToString(),
		}
		msg.Send()
		break

	//	19 RoomLocked 直播间被封禁
	case 19:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 直播间被封禁：")
		logBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		log.Warn(logBuilder.String())
		// 构造消息
		// 构造消息标题
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgTitleBuilder.WriteString(" 喜提直播间封禁！")
		// 构造消息内容
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "uname").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "area_v2_name").ToString())
		msgContentBuilder.WriteString("\n- 封禁到：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "room_Info", "lock_till").ToString())
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
			ID:      webhookId,
			IconURL: jsoniter.Get(content, "user_info", "face").ToString(),
		}
		msg.Send()
		break

	//	别的不关心，所以没写
	default:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 未知的webhook请求类型：")
		logBuilder.WriteString(jsoniter.Get(content, "type").ToString())
		log.Warn(logBuilder.String())
	}
}

// DDTVWebhookHandler 处理 DDTV 的 webhook 请求
func DDTVWebhookHandler(w http.ResponseWriter, request *http.Request) {
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
	content, err := io.ReadAll(request.Body)
	if err != nil {
		log.Errorf("读取 DDTV webhook 请求失败：%s", err.Error())
		return
	}
	go ddtvTaskRunner(content)
}
