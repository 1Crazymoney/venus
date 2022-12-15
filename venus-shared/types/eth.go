package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathbig "math/big"
	"strconv"
	"strings"

	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/go-state-types/builtin/v10/eam"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/multiformats/go-varint"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/venus/pkg/constants"
)

// mainnet
var Eip155ChainID = 314

func SetEip155ChainId(val int) {
	Eip155ChainID = val
}

type EthUint64 uint64

func (e EthUint64) MarshalJSON() ([]byte, error) {
	if e == 0 {
		return json.Marshal("0x0")
	}
	return json.Marshal(fmt.Sprintf("0x%x", e))
}

func (e *EthUint64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsedInt, err := strconv.ParseUint(strings.Replace(s, "0x", "", -1), 16, 64)
	if err != nil {
		return err
	}
	eint := EthUint64(parsedInt)
	*e = eint
	return nil
}

type EthBigInt big.Int

var (
	EthBigIntZero = EthBigInt{Int: big.Zero().Int}
)

func (e EthBigInt) MarshalJSON() ([]byte, error) {
	if e.Int == nil {
		return json.Marshal("0x0")
	}
	return json.Marshal(fmt.Sprintf("0x%x", e.Int))
}

func (e *EthBigInt) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	replaced := strings.Replace(s, "0x", "", -1)
	if len(replaced)%2 == 1 {
		replaced = "0" + replaced
	}

	i := new(mathbig.Int)
	i.SetString(replaced, 16)

	*e = EthBigInt(big.NewFromGo(i))
	return nil
}

type EthBytes []byte

func (e EthBytes) MarshalJSON() ([]byte, error) {
	if len(e) == 0 {
		return json.Marshal("0x00")
	}
	s := hex.EncodeToString(e)
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return json.Marshal("0x" + s)
}

func (e *EthBytes) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	s = strings.Replace(s, "0x", "", -1)
	if len(s)%2 == 1 {
		s = "0" + s
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	*e = decoded
	return nil
}

type EthBlock struct {
	ParentHash       EthHash    `json:"parentHash"`
	Sha3Uncles       EthHash    `json:"sha3Uncles"`
	Miner            EthAddress `json:"miner"`
	StateRoot        EthHash    `json:"stateRoot"`
	TransactionsRoot EthHash    `json:"transactionsRoot"`
	ReceiptsRoot     EthHash    `json:"receiptsRoot"`
	// TODO: include LogsBloom
	Difficulty    EthUint64 `json:"difficulty"`
	Number        EthUint64 `json:"number"`
	GasLimit      EthUint64 `json:"gasLimit"`
	GasUsed       EthUint64 `json:"gasUsed"`
	Timestamp     EthUint64 `json:"timestamp"`
	Extradata     []byte    `json:"extraData"`
	MixHash       EthHash   `json:"mixHash"`
	Nonce         EthNonce  `json:"nonce"`
	BaseFeePerGas EthBigInt `json:"baseFeePerGas"`
	Size          EthUint64 `json:"size"`
	// can be []EthTx or []string depending on query params
	Transactions []interface{} `json:"transactions"`
	Uncles       []EthHash     `json:"uncles"`
}

var (
	EmptyEthHash  = EthHash{}
	EmptyEthInt   = EthUint64(0)
	EmptyEthNonce = [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
)

func NewEthBlock() EthBlock {
	return EthBlock{
		Sha3Uncles:       EmptyEthHash,
		StateRoot:        EmptyEthHash,
		TransactionsRoot: EmptyEthHash,
		ReceiptsRoot:     EmptyEthHash,
		Difficulty:       EmptyEthInt,
		Extradata:        []byte{},
		MixHash:          EmptyEthHash,
		Nonce:            EmptyEthNonce,
		GasLimit:         EthUint64(constants.BlockGasLimit), // TODO we map Ethereum blocks to Filecoin tipsets; this is inconsistent.
		Uncles:           []EthHash{},
		Transactions:     []interface{}{},
	}
}

type EthCall struct {
	From     EthAddress  `json:"from"`
	To       *EthAddress `json:"to"`
	Gas      EthUint64   `json:"gas"`
	GasPrice EthBigInt   `json:"gasPrice"`
	Value    EthBigInt   `json:"value"`
	Data     EthBytes    `json:"data"`
}

func (c *EthCall) UnmarshalJSON(b []byte) error {
	type TempEthCall EthCall
	var params TempEthCall

	if err := json.Unmarshal(b, &params); err != nil {
		return err
	}
	*c = EthCall(params)
	return nil
}

type EthTxReceipt struct {
	TransactionHash  EthHash     `json:"transactionHash"`
	TransactionIndex EthUint64   `json:"transactionIndex"`
	BlockHash        EthHash     `json:"blockHash"`
	BlockNumber      EthUint64   `json:"blockNumber"`
	From             EthAddress  `json:"from"`
	To               *EthAddress `json:"to"`
	// Logs
	// LogsBloom
	StateRoot         EthHash     `json:"root"`
	Status            EthUint64   `json:"status"`
	ContractAddress   *EthAddress `json:"contractAddress"`
	CumulativeGasUsed EthUint64   `json:"cumulativeGasUsed"`
	GasUsed           EthUint64   `json:"gasUsed"`
	EffectiveGasPrice EthBigInt   `json:"effectiveGasPrice"`
	LogsBloom         EthBytes    `json:"logsBloom"`
	Logs              []string    `json:"logs"`
}

func NewEthTxReceipt(tx EthTx, lookup *MsgLookup, replay *InvocResult) (EthTxReceipt, error) {
	receipt := EthTxReceipt{
		TransactionHash:  tx.Hash,
		TransactionIndex: tx.TransactionIndex,
		BlockHash:        tx.BlockHash,
		BlockNumber:      tx.BlockNumber,
		From:             tx.From,
		To:               tx.To,
		StateRoot:        EmptyEthHash,
		LogsBloom:        []byte{0},
		Logs:             []string{},
	}

	if receipt.To == nil && lookup.Receipt.ExitCode.IsSuccess() {
		// Create and Create2 return the same things.
		var ret eam.CreateReturn
		if err := ret.UnmarshalCBOR(bytes.NewReader(lookup.Receipt.Return)); err != nil {
			return EthTxReceipt{}, fmt.Errorf("failed to parse contract creation result: %w", err)
		}
		addr := EthAddress(ret.EthAddress)
		receipt.ContractAddress = &addr
	}

	if lookup.Receipt.ExitCode.IsSuccess() {
		receipt.Status = 1
	}
	if lookup.Receipt.ExitCode.IsError() {
		receipt.Status = 0
	}

	receipt.GasUsed = EthUint64(lookup.Receipt.GasUsed)

	// TODO: handle CumulativeGasUsed
	receipt.CumulativeGasUsed = EmptyEthInt

	effectiveGasPrice := big.Div(replay.GasCost.TotalCost, big.NewInt(lookup.Receipt.GasUsed))
	receipt.EffectiveGasPrice = EthBigInt(effectiveGasPrice)
	return receipt, nil
}

const (
	ETH_ADDRESS_LENGTH = 20 // nolint
	ETH_HASH_LENGTH    = 32 // nolint
)

type EthNonce [8]byte

func (n EthNonce) String() string {
	return "0x" + hex.EncodeToString(n[:])
}

func (n EthNonce) MarshalJSON() ([]byte, error) {
	return json.Marshal((n.String()))
}

type EthAddress [ETH_ADDRESS_LENGTH]byte

func (a EthAddress) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a EthAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal((a.String()))
}

func (a *EthAddress) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	addr, err := EthAddressFromHex(s)
	if err != nil {
		return err
	}
	copy(a[:], addr[:])
	return nil
}

func (a EthAddress) ToFilecoinAddress() (address.Address, error) {
	expectedPrefix := [12]byte{0xff}
	if !bytes.Equal(a[:12], expectedPrefix[:]) {
		// TODO: handle f4 once we've added support to go-address.
		return address.Address{}, fmt.Errorf("not a valid id-in-eth address: %s", a)
	}
	id := binary.BigEndian.Uint64(a[12:])
	return address.NewIDAddress(id)
}

func TryEthAddressFromFilecoinAddress(addr address.Address, allowId bool) (EthAddress, bool, error) {
	switch addr.Protocol() {
	case address.ID:
		if !allowId {
			return EthAddress{}, false, nil
		}
		id, err := address.IDFromAddress(addr)
		if err != nil {
			return EthAddress{}, false, err
		}
		var ethaddr EthAddress
		ethaddr[0] = 0xff
		binary.BigEndian.PutUint64(ethaddr[12:], id)
		return ethaddr, true, nil
	case address.Delegated:
		payload := addr.Payload()
		namespace, n, err := varint.FromUvarint(payload)
		if err != nil {
			return EthAddress{}, false, fmt.Errorf("invalid delegated address namespace in: %s", addr)
		}
		payload = payload[n:]
		if namespace == builtintypes.EthereumAddressManagerActorID {
			addr, err := EthAddressFromBytes(payload)
			return addr, err == nil, err
		}
	}
	return EthAddress{}, false, nil
}

func EthAddressFromFilecoinAddress(addr address.Address) (EthAddress, error) {
	ethAddr, ok, err := TryEthAddressFromFilecoinAddress(addr, true)
	if !ok && err == nil {
		err = fmt.Errorf("failed to convert filecoin address %s to an equivalent eth address", addr)
	}
	return ethAddr, err
}

func EthAddressFromHex(s string) (EthAddress, error) {
	handlePrefix(&s)
	b, err := decodeHexString(s, ETH_ADDRESS_LENGTH)
	if err != nil {
		return EthAddress{}, err
	}
	var h EthAddress
	copy(h[ETH_ADDRESS_LENGTH-len(b):], b)
	return h, nil
}

func EthAddressFromBytes(b []byte) (EthAddress, error) {
	var a EthAddress
	if len(b) != ETH_ADDRESS_LENGTH {
		return EthAddress{}, fmt.Errorf("cannot initiate a new EthAddress: incorrect input length")
	}
	copy(a[:], b[:])
	return a, nil
}

type EthHash [ETH_HASH_LENGTH]byte

func (h EthHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h *EthHash) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	hash, err := EthHashFromHex(s)
	if err != nil {
		return err
	}
	copy(h[:], hash[:])
	return nil
}

func handlePrefix(s *string) {
	if strings.HasPrefix(*s, "0x") || strings.HasPrefix(*s, "0X") {
		*s = (*s)[2:]
	}
	if len(*s)%2 == 1 {
		*s = "0" + *s
	}
}

func decodeHexString(s string, length int) ([]byte, error) {
	b, err := hex.DecodeString(s)

	if err != nil {
		return []byte{}, fmt.Errorf("cannot parse hash: %w", err)
	}

	if len(b) > length {
		return []byte{}, fmt.Errorf("length of decoded bytes is longer than %d", length)
	}

	return b, nil
}

func EthHashFromCid(c cid.Cid) (EthHash, error) {
	return EthHashFromHex(c.Hash().HexString()[8:])
}

func EthHashFromHex(s string) (EthHash, error) {
	handlePrefix(&s)
	b, err := decodeHexString(s, ETH_HASH_LENGTH)
	if err != nil {
		return EthHash{}, err
	}
	var h EthHash
	copy(h[ETH_HASH_LENGTH-len(b):], b)
	return h, nil
}

func (h EthHash) String() string {
	return "0x" + hex.EncodeToString(h[:])
}

func (h EthHash) ToCid() cid.Cid {
	// err is always nil
	mh, _ := multihash.EncodeName(h[:], "blake2b-256")

	return cid.NewCidV1(cid.DagCBOR, mh)
}

type EthFeeHistory struct {
	OldestBlock   uint64         `json:"oldestBlock"`
	BaseFeePerGas []EthBigInt    `json:"baseFeePerGas"`
	GasUsedRatio  []float64      `json:"gasUsedRatio"`
	Reward        *[][]EthBigInt `json:"reward,omitempty"`
}
