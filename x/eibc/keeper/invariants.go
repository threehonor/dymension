package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

const (
	demandOrderCountInvariantName = "demand-order-count"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "demand-order-count", DemandOrderCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "underlying-packet-exist", UnderlyingPacketExistInvariant(k))
	ir.RegisterRoute(types.ModuleName, "coins", CoinsInvariant(k))
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []sdk.Invariant{
			DemandOrderCountInvariant(k),
			UnderlyingPacketExistInvariant(k),
			CoinsInvariant(k),
		} {
			res, stop := inv(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

func DemandOrderCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		allDemandOrders, err := k.ListAllDemandOrders(ctx)
		if err != nil {
			msg += fmt.Sprintf("list all demand orders failed: %v\n", err)
			broken = true
		}
		pendingDemandOrders, err := k.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING, 0)
		if err != nil {
			msg += fmt.Sprintf("list pending demand orders failed: %v\n", err)
			broken = true
		}
		finalizedDemandOrders, err := k.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED, 0)
		if err != nil {
			msg += fmt.Sprintf("list finalized demand orders failed: %v\n", err)
			broken = true
		}
		// Validate the count of demand orders is equal to the sum of demand orders in all statuses
		if len(allDemandOrders) != len(pendingDemandOrders)+len(finalizedDemandOrders) {
			msg += fmt.Sprintf("demand orders count mismatch: all(%d) != pending(%d)  + finalized(%d)\n",
				len(allDemandOrders), len(pendingDemandOrders), len(finalizedDemandOrders))
			broken = true
		}
		return sdk.FormatInvariant(types.ModuleName, demandOrderCountInvariantName, msg), broken
	}
}

func UnderlyingPacketExistInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		allDemandOrders, err := k.ListAllDemandOrders(ctx)
		if err != nil {
			msg += fmt.Sprintf("list all demand orders failed: %v\n", err)
			broken = true
		}
		for _, demandOrder := range allDemandOrders {
			// Get the underlying packet for the demand order
			_, err := k.dack.GetRollappPacket(ctx, demandOrder.TrackingPacketKey)
			if err != nil {
				msg += fmt.Sprintf("underlying packet for demand order %s not found: %v\n", demandOrder.Id, err)
				broken = true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "underlying-packet-exist", msg), broken
	}
}

// coins (price,fee) are sensible
func CoinsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		allDemandOrders, err := k.ListAllDemandOrders(ctx)
		if err != nil {
			msg += fmt.Sprintf("list all demand orders failed: %v\n", err)
			broken = true
		}
		for _, do := range allDemandOrders {
			for _, coins := range []sdk.Coins{do.Price, do.Fee} {
				if len(coins) == 0 {
					// This is OK, since coins will erase the zero coin and zero price/fee is allowed
					continue
				}
				if len(coins) > 1 {
					msg += fmt.Sprintf("multiple coins: %s\n", coins)
					broken = true
					continue
				}
				if coins[0].IsNegative() {
					msg += fmt.Sprintf("negative coins: %s\n", coins)
					broken = true
				}
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "coins", msg), broken
	}
}
