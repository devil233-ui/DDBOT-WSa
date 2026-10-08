package bilibili

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseSelfLiveRoomId 覆盖"订阅自己直播"时解析直播间号的各分支。
// 关键回归：无效房间号（如 0）时 bilibili 返回空数组，取真实响应体自行解析，
// 必须返回明确错误而不是把 0 当作有效房间号继续查询。
func TestParseSelfLiveRoomId(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    int64
		wantErr bool
	}{
		{
			name: "正常返回直播间号",
			body: `{"code":0,"message":"OK","ttl":1,"data":{"room_id":1930400660,"main_site_level_rank":10000}}`,
			want: 1930400660,
		},
		{
			name:    "cookie失效",
			body:    `{"code":-101,"message":"账号未登录","data":null}`,
			wantErr: true,
		},
		{
			name:    "缺少直播间号",
			body:    `{"code":0,"message":"OK","data":{}}`,
			wantErr: true,
		},
		{
			name:    "直播间号为0",
			body:    `{"code":0,"message":"OK","data":{"room_id":0}}`,
			wantErr: true,
		},
		{
			name:    "空响应",
			body:    "   ",
			wantErr: true,
		},
		{
			name:    "非JSON响应",
			body:    `<html>风控</html>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSelfLiveRoomId([]byte(tt.body))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
