package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestLatestStateInfoIndexQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLatestStateInfoIndex(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetLatestStateInfoIndexRequest
		response *types.QueryGetLatestStateInfoIndexResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLatestStateInfoIndexRequest{
				RollappId: msgs[0].RollappId,
			},
			response: &types.QueryGetLatestStateInfoIndexResponse{LatestStateInfoIndex: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetLatestStateInfoIndexRequest{
				RollappId: msgs[1].RollappId,
			},
			response: &types.QueryGetLatestStateInfoIndexResponse{LatestStateInfoIndex: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetLatestStateInfoIndexRequest{
				RollappId: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LatestStateInfoIndex(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestLatestStateInfoIndexQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLatestStateInfoIndex(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLatestStateInfoIndexRequest {
		return &types.QueryAllLatestStateInfoIndexRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LatestStateInfoIndexAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LatestStateInfoIndex), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LatestStateInfoIndex),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LatestStateInfoIndexAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LatestStateInfoIndex), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LatestStateInfoIndex),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LatestStateInfoIndexAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.LatestStateInfoIndex),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LatestStateInfoIndexAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
