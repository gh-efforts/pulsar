package subscriber

import (
	"testing"

	"github.com/filecoin-project/go-state-types/abi"

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
				Msg:      &types.Message{To: t0123, From: t0123, Method: 1, Nonce: 1},
				Subcalls: nil,
			},
			},
			want1: 1,
		},
		{
			name: "countMsg1",
			args: args{trace: types.ExecutionTrace{
				Msg: &types.Message{To: t0123, From: t0123, Method: 1, Nonce: 1},
				Subcalls: []types.ExecutionTrace{
					{Subcalls: nil, Msg: &types.Message{To: t0123, From: t0123, Method: 2, Nonce: 1}},
				}},
			},
			want1: 2,
		},

		{
			name: "countMsg3",
			args: args{trace: types.ExecutionTrace{
				Msg: &types.Message{To: t0123, From: t0123, Method: 1, Nonce: 1},
				Subcalls: []types.ExecutionTrace{
					{
						Msg:      &types.Message{To: t0123, From: t0123, Method: 2, Nonce: 1},
						Subcalls: nil,
					},
					{
						Msg: &types.Message{To: t0123, From: t0123, Method: 3, Nonce: 1},
						Subcalls: []types.ExecutionTrace{
							{
								Msg:      &types.Message{To: t0123, From: t0123, Method: 4, Nonce: 1},
								Subcalls: nil,
							},
						},
					},
				},
			},
			},
			want1: 4,
		},

		{
			name: "countMsg5",
			args: args{trace: types.ExecutionTrace{
				Msg: &types.Message{To: t0123, From: t0123, Method: 1, Nonce: 1},
				Subcalls: []types.ExecutionTrace{
					{
						Msg:      &types.Message{To: t0123, From: t0123, Method: 2, Nonce: 1},
						Subcalls: nil,
					},
					{
						Msg: &types.Message{To: t0123, From: t0123, Method: 3, Nonce: 1},
						Subcalls: []types.ExecutionTrace{
							{
								Msg: &types.Message{To: t0123, From: t0123, Method: 4, Nonce: 1},
								Subcalls: []types.ExecutionTrace{
									{
										Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
										Subcalls: []types.ExecutionTrace{
											{
												Msg: &types.Message{To: t0123, From: t0123, Method: 6, Nonce: 1},
												Subcalls: []types.ExecutionTrace{
													{
														Msg:      &types.Message{To: t0123, From: t0123, Method: 7, Nonce: 1},
														Subcalls: nil,
													},
												}},
										},
									},
								},
							},
						},
					},
				},
			},
			},
			want1: 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := countMsg(tt.args.trace)
			assert.Equal(t, len(got1), got2, "countMsg() got1 = %v, want %v", got1, tt.want1)
			for i := 0; i < len(got1); i++ {
				assert.Equal(t, got1[i].Method, abi.MethodNum(i+1), "countMsg() got1[%d] = %v, want %v", i, got1[i].Method, abi.MethodNum(i+1))
			}
		})
	}
}
