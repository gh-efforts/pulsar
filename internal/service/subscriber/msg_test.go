package subscriber

import (
	"fmt"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/stretchr/testify/assert"
)

func Test_countMsg(t *testing.T) {
	type args struct {
		trace types.ExecutionTrace
	}

	t0123, err := address.NewFromString("t0123")
	assert.Nil(t, err)

	tests := []struct {
		name  string
		args  args
		want  []types.Message
		want1 int
	}{
		{
			name: "countMsg0",
			args: args{trace: types.ExecutionTrace{
				Msg:      &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
				Subcalls: nil,
			},
			},
			want1: 1,
		},
		{
			name: "countMsg1",
			args: args{trace: types.ExecutionTrace{
				Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
				Subcalls: []types.ExecutionTrace{
					{Subcalls: nil, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
				}},
			},
			want1: 2,
		},

		{
			name: "countMsg3",
			args: args{trace: types.ExecutionTrace{
				Subcalls: []types.ExecutionTrace{
					{
						Msg:      &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
						Subcalls: nil,
					},
					{
						Subcalls: []types.ExecutionTrace{
							{
								Subcalls: nil, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
							},
						},
						Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
					},
				},
				Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
			},
			},
			want1: 4,
		},

		{
			name: "countMsg5",
			args: args{trace: types.ExecutionTrace{
				Subcalls: []types.ExecutionTrace{
					{Subcalls: nil, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
					{Subcalls: []types.ExecutionTrace{
						{Subcalls: []types.ExecutionTrace{
							{Subcalls: []types.ExecutionTrace{
								{Subcalls: []types.ExecutionTrace{
									{Subcalls: nil, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
								}, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
							}, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
						}, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
					}, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
				},
				Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
			},
			},
			want1: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := countMsg(tt.args.trace)
			if len(got1) != got2 {
				t.Errorf("countMsg() got1 = %v, want %v", got1, tt.want1)
			}
			assert.Equal(t, got2, tt.want1, fmt.Sprintf("countMsg() got = %v, want %v", got2, tt.want))

		})
	}
}
