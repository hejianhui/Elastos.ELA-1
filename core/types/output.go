package types

import (
	"errors"
	"fmt"
	"io"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/types/outputpayload"
)

// OutputType represents the type of a output payload.
type OutputType byte

const (
	// OTNone indicates there is no payload for this output.
	OTNone OutputType = 0x00

	// OTVote indicates the output payload is a vote.
	OTVote OutputType = 0x01
)

type OutputPayload interface {
	// Get payload data
	Data() []byte
	Serialize(w io.Writer) error
	Deserialize(r io.Reader) error
	GetVersion() byte
	Validate() error
}

type Output struct {
	AssetID     common.Uint256
	Value       common.Fixed64
	OutputLock  uint32
	ProgramHash common.Uint168
	Type        OutputType
	Payload     OutputPayload
}

func (o *Output) Serialize(w io.Writer, txVersion TransactionVersion) error {
	if err := o.AssetID.Serialize(w); err != nil {
		return err
	}

	if err := o.Value.Serialize(w); err != nil {
		return err
	}

	if err := common.WriteUint32(w, o.OutputLock); err != nil {
		return err
	}

	if err := o.ProgramHash.Serialize(w); err != nil {
		return err
	}

	if txVersion >= TxVersion09 {
		if err := common.WriteUint8(w, byte(o.Type)); err != nil {
			return err
		}
		if err := o.Payload.Serialize(w); err != nil {
			return err
		}
	}

	return nil
}

func (o *Output) Deserialize(r io.Reader, txVersion TransactionVersion) error {
	if err := o.AssetID.Deserialize(r); err != nil {
		return err
	}

	if err := o.Value.Deserialize(r); err != nil {
		return err
	}

	temp, err := common.ReadUint32(r)
	if err != nil {
		return err
	}
	o.OutputLock = temp

	if err := o.ProgramHash.Deserialize(r); err != nil {
		return err
	}

	if txVersion >= TxVersion09 {
		outputType, err := common.ReadUint8(r)
		if err != nil {
			return err
		}
		o.Type = OutputType(outputType)
		o.Payload, err = getOutputPayload(OutputType(outputType))
		if err != nil {
			return err
		}
		if err = o.Payload.Deserialize(r); err != nil {
			return err
		}
	}

	return nil
}

func (o *Output) String() string {
	outputStr := fmt.Sprint("Output: {\n\t",
		"AssetID: ", o.AssetID.String(), "\n\t",
		"Value: ", o.Value.String(), "\n\t",
		"OutputLock: ", o.OutputLock, "\n\t",
		"ProgramHash: ", o.ProgramHash.String(), "\n\t")

	if o.Payload != nil {
		outputStr += fmt.Sprint("Type: ", o.Type, "\n\t", "Payload: ", o.Payload, "\n\t")
	}
	outputStr += "}"

	return outputStr
}

func getOutputPayload(outputType OutputType) (OutputPayload, error) {
	var op OutputPayload
	switch outputType {
	case OTNone:
		op = new(outputpayload.DefaultOutput)
	case OTVote:
		op = new(outputpayload.VoteOutput)
	default:
		return nil, errors.New("invalid transaction output type")
	}
	return op, nil
}