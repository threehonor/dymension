package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) SaveDistribution(ctx sdk.Context, d types.Distribution) error {
	return k.distribution.Set(ctx, d)
}

func (k Keeper) GetDistribution(ctx sdk.Context) (types.Distribution, error) {
	return k.distribution.Get(ctx)
}

// UpdateDistribution updates the distribution by applying the provided function to the current distribution.
// It retrieves the current distribution from the state, applies the update function to it, saves the updated distribution
// back to the state, and returns the updated distribution. If any error occurs during these steps, it returns an error.
func (k Keeper) UpdateDistribution(ctx sdk.Context, fn func(types.Distribution) types.Distribution) (types.Distribution, error) {
	// Get the current plan from the state
	current, err := k.GetDistribution(ctx)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to get distribution: %w", err)
	}

	// Apply the update
	result := fn(current)

	// Save the updated distribution
	err = k.SaveDistribution(ctx, result)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to save distribution: %w", err)
	}

	// Return the updated distribution
	return result, nil
}

func (k Keeper) SaveDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress, power math.Int) error {
	return k.delegatorValidatorPower.Set(ctx, collections.Join(voterAddr, valAddr), power)
}

func (k Keeper) GetDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress) (math.Int, error) {
	return k.delegatorValidatorPower.Get(ctx, collections.Join(voterAddr, valAddr))
}

func (k Keeper) HasDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress) (bool, error) {
	return k.delegatorValidatorPower.Has(ctx, collections.Join(voterAddr, valAddr))
}

func (k Keeper) IterateDelegatorValidatorPower(
	ctx sdk.Context,
	voterAddr sdk.AccAddress,
	fn func(valAddr sdk.ValAddress, power math.Int) (stop bool, err error),
) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](voterAddr)
	iterator, err := k.delegatorValidatorPower.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			return err
		}

		stop, err := fn(kv.Key.K2(), kv.Value)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	return nil
}

func (k Keeper) DeleteDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return k.delegatorValidatorPower.Remove(ctx, collections.Join(voterAddr, valAddr))
}

func (k Keeper) DeleteDelegatorPower(ctx sdk.Context, voterAddr sdk.AccAddress) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.ValAddress](voterAddr)
	return k.delegatorValidatorPower.Clear(ctx, rng)
}

func (k Keeper) SaveVote(ctx sdk.Context, voterAddr sdk.AccAddress, v types.Vote) error {
	return k.votes.Set(ctx, voterAddr, v)
}

func (k Keeper) GetVote(ctx sdk.Context, voterAddr sdk.AccAddress) (types.Vote, error) {
	return k.votes.Get(ctx, voterAddr)
}

func (k Keeper) Voted(ctx sdk.Context, voterAddr sdk.AccAddress) (bool, error) {
	return k.votes.Has(ctx, voterAddr)
}

func (k Keeper) DeleteVote(ctx sdk.Context, voterAddr sdk.AccAddress) error {
	return k.votes.Remove(ctx, voterAddr)
}

func (k Keeper) IterateVotes(
	ctx sdk.Context,
	fn func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error),
) error {
	iterator, err := k.votes.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			return err
		}

		stop, err := fn(kv.Key, kv.Value)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	return nil
}

func (k Keeper) GetEndorsement(ctx sdk.Context, rollappID string) (types.Endorsement, error) {
	return k.raEndorsements.Get(ctx, rollappID)
}

func (k Keeper) SaveEndorsement(ctx sdk.Context, e types.Endorsement) error {
	return k.raEndorsements.Set(ctx, e.RollappId, e)
}

func (k Keeper) HasEndorsement(ctx sdk.Context, rollappID string) (bool, error) {
	return k.raEndorsements.Has(ctx, rollappID)
}

// UpdateEndorsement updates the endorsement by applying the provided function to the current endorsement.
// It retrieves the current endorsement from the state, applies the update function to it, saves the updated endorsement
// back to the state, and returns the updated endorsement. If any error occurs during these steps, it returns an error.
func (k Keeper) UpdateEndorsement(ctx sdk.Context, rollappID string, updates ...types.EndorsementUpdateFn) error {
	current, err := k.GetEndorsement(ctx, rollappID)
	if err != nil {
		return fmt.Errorf("get endorsement: %w", err)
	}

	for _, update := range updates {
		current = update(current)
	}

	err = k.SaveEndorsement(ctx, current)
	if err != nil {
		return fmt.Errorf("save endorsement: %w", err)
	}

	return nil
}

func (k Keeper) GetAllEndorsements(ctx sdk.Context) ([]types.Endorsement, error) {
	iterator, err := k.raEndorsements.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iterator.Close() // nolint: errcheck
	return iterator.Values()
}

func (k Keeper) CanClaim(ctx sdk.Context, addr sdk.AccAddress) (bool, error) {
	blacklisted, err := k.claimBlacklist.Has(ctx, addr)
	if err != nil {
		return false, fmt.Errorf("check claim blacklist: %w", err)
	}
	voted, err := k.votes.Has(ctx, addr)
	if err != nil {
		return false, fmt.Errorf("check vote: %w", err)
	}
	return !blacklisted && voted, err
}

func (k Keeper) BlacklistClaim(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.claimBlacklist.Set(ctx, addr)
}

func (k Keeper) RefreshClaimBlacklist(ctx sdk.Context) error {
	return k.claimBlacklist.Clear(ctx, nil)
}
