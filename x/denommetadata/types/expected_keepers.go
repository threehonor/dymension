package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// BankKeeper defines the expected interface needed
type BankKeeper interface {
	GetDenomMetaData(ctx context.Context, denom string) (types.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData types.Metadata)
}

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata types.Metadata) error
	HasDenomMetadata(ctx sdk.Context, base string) bool
}

type RollappKeeper interface {
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	GetValidTransfer(
		ctx sdk.Context,
		packetData []byte,
		raPortOnHub, raChanOnHub string,
	) (data rollapptypes.TransferData, err error)
	SetRegisteredDenom(ctx sdk.Context, rollappID, denom string) error
	HasRegisteredDenom(ctx sdk.Context, rollappID, denom string) (bool, error)
	ClearRegisteredDenoms(ctx sdk.Context, rollappID string) error
}
