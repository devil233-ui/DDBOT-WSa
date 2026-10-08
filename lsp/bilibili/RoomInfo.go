package bilibili

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cnxysoft/DDBOT-WSa/proxy_pool"
	"github.com/cnxysoft/DDBOT-WSa/requests"
	"github.com/cnxysoft/DDBOT-WSa/utils"
	"strconv"
	"time"
)

const (
	PathRoomInfo = "/room/v1/Room/get_info"
	// PathUcenterLiveInfo 查询当前 cookie 账号自己的直播信息（含直播间号）。
	// 订阅自己直播时用：FeedList 是关注列表的直播流、不含自己，
	// UserInfo 缓存里的 room_id 也可能缺失或为 0。
	PathUcenterLiveInfo = "/xlive/web-ucenter/user/live_info"
)

// UcenterLiveInfoResponse 自己直播信息响应，只取直播间号
type UcenterLiveInfoResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RoomId int64 `json:"room_id"`
	} `json:"data"`
}

type RoomInfoRequest struct {
	RoomId int64 `json:"room_id"`
}

func GetRoomInfo(roomId int64) (*LiveRoomData, error) {
	params, err := utils.ToDatas(&RoomInfoRequest{
		RoomId: roomId,
	})
	if err != nil {
		return nil, err
	}
	opts := []requests.Option{
		AddUAOption(),
		requests.TimeoutOption(time.Second * 15),
		requests.RequestAutoHostOption(),
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.WithCookieJar(cj.Load()),
	}
	opts = append(opts, GetVerifyOption()...)
	var resp = new(LiveRoomData)
	err = requests.Get(BPath(PathRoomInfo), params, resp, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetSelfLiveRoomId 查询当前 cookie 账号自己的直播间号。
// 订阅自己直播时不能依赖 UserInfo 缓存的 room_id（可能缺失或为 0），
// FeedList 也不含自己，因此单独走 ucenter 接口解析。
func GetSelfLiveRoomId() (int64, error) {
	opts := []requests.Option{
		AddUAOption(),
		requests.TimeoutOption(time.Second * 15),
		requests.RequestAutoHostOption(),
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.WithCookieJar(cj.Load()),
	}
	opts = append(opts, GetVerifyOption()...)
	var buf bytes.Buffer
	if err := requests.Get(BPath(PathUcenterLiveInfo), nil, &buf, opts...); err != nil {
		return 0, err
	}
	return parseSelfLiveRoomId(buf.Bytes())
}

// parseSelfLiveRoomId 从自己直播信息响应中解析直播间号（独立出来便于单测覆盖失败分支）。
func parseSelfLiveRoomId(body []byte) (int64, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return 0, errors.New("自己直播信息为空响应")
	}
	var resp UcenterLiveInfoResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, fmt.Errorf("解析自己直播信息失败 %v", err)
	}
	if resp.Code != 0 {
		return 0, fmt.Errorf("查询自己直播信息失败 code %v msg %v", resp.Code, resp.Message)
	}
	if resp.Data.RoomId <= 0 {
		return 0, errors.New("自己直播信息中没有有效的直播间号")
	}
	return resp.Data.RoomId, nil
}

func (l *LiveRoomData) GetLiveStatus() LiveStatus {
	if l.Data.LiveStatus == 1 {
		return LiveStatus_Living
	}
	return LiveStatus_NoLiving
}

func (l *LiveRoomData) GetCover() string {
	return l.Data.UserCover
}

func (l *LiveRoomData) GetTitle() string {
	return l.Data.Title
}

func (l *LiveRoomData) GetUrl() string {
	return "https://live.bilibili.com/" + strconv.FormatInt(l.Data.RoomId, 10)
}
