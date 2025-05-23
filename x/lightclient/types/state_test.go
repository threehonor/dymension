package types_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

var (
	sequencerPubKey = ed25519.GenPrivKey().PubKey()
	timestamp       = time.Unix(1724392989, 0)

	seq = sequencertypes.NewTestSequencer(sequencerPubKey)

	validIBCState = ibctm.ConsensusState{
		Root:               commitmenttypes.NewMerkleRoot([]byte("root")),
		Timestamp:          timestamp,
		NextValidatorsHash: seq.MustValsetHash(),
	}
	validRollappState = types.RollappState{
		BlockDescriptor: rollapptypes.BlockDescriptor{
			StateRoot: []byte("root"),
			Timestamp: timestamp,
		},
		NextBlockSequencer: seq,
	}
)

func TestCheckCompatibility(t *testing.T) {
	type input struct {
		ibcState ibctm.ConsensusState
		raState  types.RollappState
	}
	testCases := []struct {
		name  string
		input func() input
		err   error
	}{
		{
			name: "roots are not equal",
			input: func() input {
				invalidRootRaState := validRollappState
				invalidRootRaState.BlockDescriptor.StateRoot = []byte("not same root")
				return input{
					ibcState: validIBCState,
					raState:  invalidRootRaState,
				}
			},
			err: types.ErrStateRootMismatch,
		},
		{
			name: "nextValidatorHash does not match the sequencer who submitted the next block descriptor",
			input: func() input {
				invalidNextValidatorHashIBCState := validIBCState
				invalidNextValidatorHashIBCState.NextValidatorsHash = []byte("wrong next validator hash")
				return input{
					ibcState: invalidNextValidatorHashIBCState,
					raState:  validRollappState,
				}
			},
			err: types.ErrNextValHashMismatch,
		},
		{
			name: "timestamps is empty. ignore timestamp check",
			input: func() input {
				emptyTimestampRAState := validRollappState
				emptyTimestampRAState.BlockDescriptor.Timestamp = time.Time{}
				return input{
					ibcState: validIBCState,
					raState:  emptyTimestampRAState,
				}
			},
			err: nil,
		},
		{
			name: "timestamps are not equal",
			input: func() input {
				invalidTimestampRAState := validRollappState
				invalidTimestampRAState.BlockDescriptor.Timestamp = timestamp.Add(1)
				return input{
					ibcState: validIBCState,
					raState:  invalidTimestampRAState,
				}
			},
			err: types.ErrTimestampMismatch,
		},
		{
			name: "all fields are compatible",
			input: func() input {
				return input{
					ibcState: validIBCState,
					raState:  validRollappState,
				}
			},
			err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.input()
			err := types.CheckCompatibility(input.ibcState, input.raState)
			if err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
