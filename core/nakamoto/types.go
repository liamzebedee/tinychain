package nakamoto

import (
	"database/sql"
	"encoding/hex"
	"math/big"
	"strconv"
	"time"
)

// The Nakamoto consensus configuration, pertaining to difficulty readjustment, genesis block, and block size.
type ConsensusConfig struct {
	// The length of an epoch.
	EpochLengthBlocks uint64 `json:"epoch_length_blocks"`

	// The target block production rate in terms of 1 epoch.
	TargetEpochLengthMillis uint64 `json:"target_epoch_length_millis"`

	// Genesis difficulty target.
	GenesisDifficulty big.Int `json:"genesis_difficulty"`

	// The genesis parent block hash.
	GenesisParentBlockHash [32]byte `json:"genesis_block_hash"`

	// Maximum block size.
	MaxBlockSizeBytes uint64 `json:"max_block_size_bytes"`
}

// A raw block is the block as transmitted on the network.
// It contains the block header and the block body.
// It does not contain any block metadata such as height, epoch, or difficulty.
type RawBlock struct {
	// Block header.
	ParentHash             [32]byte `json:"parent_hash"`
	ParentTotalWork        [32]byte `json:"parent_total_work"`
	Difficulty             [32]byte `json:"difficulty"`
	Timestamp              uint64   `json:"timestamp"`
	NumTransactions        uint64   `json:"num_transactions"`
	TransactionsMerkleRoot [32]byte `json:"transactions_merkle_root"`
	Nonce                  [32]byte `json:"nonce"`
	Graffiti               [32]byte `json:"graffiti"`

	// Block body.
	Transactions []RawTransaction `json:"transactions"`
}

type RawTransaction struct {
	Version    byte     `json:"version"`
	Sig        [64]byte `json:"sig"`
	FromPubkey [65]byte `json:"from"`
	ToPubkey   [65]byte `json:"to"`
	Amount     uint64   `json:"amount"`
	Fee        uint64   `json:"fee"`
	Nonce      uint64   `json:"nonce"`
}

func (tx *RawTransaction) SizeBytes() uint64 {
	// Size of the transaction is the size of the envelope.
	return 1 + 65 + 65 + 8 + 8 + 8
}

// TODO embed in Block?
type BlockHeader struct {
	ParentHash             [32]byte
	ParentTotalWork        big.Int
	Timestamp              uint64
	NumTransactions        uint64
	TransactionsMerkleRoot [32]byte
	Nonce                  [32]byte
	Graffiti               [32]byte
}

type Block struct {
	// Block header.
	ParentHash             [32]byte
	ParentTotalWork        big.Int
	Timestamp              uint64
	NumTransactions        uint64
	TransactionsMerkleRoot [32]byte
	Nonce                  [32]byte
	Graffiti               [32]byte

	// Block body.
	Transactions []RawTransaction

	// Metadata.
	Height          uint64
	Epoch           string
	Work            big.Int
	SizeBytes       uint64
	Hash            [32]byte
	AccumulatedWork big.Int
}

func (b *Block) HashStr() string {
	sl := b.Hash[:]
	return hex.EncodeToString(sl)
}

type Transaction struct {
	Version    byte     `json:"version"`
	Sig        [64]byte `json:"sig"`
	FromPubkey [65]byte `json:"from"`
	ToPubkey   [65]byte `json:"to"`
	Amount     uint64   `json:"amount"`
	Fee        uint64   `json:"fee"`
	Nonce      uint64   `json:"nonce"`

	Hash      [32]byte
	Blockhash [32]byte
	TxIndex   uint64
}

type BlockDAGInterface interface {
	// Ingest block.
	IngestBlock(b Block) error

	// Get block.
	GetBlockByHash(hash [32]byte) (*Block, error)

	// Get block's transactions.
	GetBlockTransactions(hash [32]byte) (*[]Transaction, error)

	// Get epoch for block.
	GetEpochForBlockHash(blockhash [32]byte) (*Epoch, error)

	// Get the tip of the chain, given a minimum number of confirmations.
	GetLatestTip() (Block, error)

	// Get the raw bytes of a block.
	GetRawBlockDataByHash(hash [32]byte) ([]byte, error)
}

// The block DAG is the core data structure of the Nakamoto consensus protocol.
// It is a directed acyclic graph of blocks, where each block has a parent block.
// As it is infeasible to store the entirety of the blockchain in-memory,
// the block DAG is backed by a SQL database.
type BlockDAG struct {
	// The backing SQL database store, which stores:
	// - blocks
	// - epochs
	// - transactions
	db *sql.DB

	// The state machine.
	stateMachine StateMachineInterface

	// Consensus settings.
	consensus ConsensusConfig

	// Latest tip.
	Tip Block

	// OnNewTip handler.
	OnNewTip func(tip Block, prevTip Block)
}

type StateMachineInterface interface {
	VerifyTx(tx RawTransaction) error
}

type Epoch struct {
	// Epoch number.
	Number uint64

	// Epoch unique ID.
	Id string

	// Start block.
	StartBlockHash [32]byte
	// Start time.
	StartTime uint64
	// Start height.
	StartHeight uint64

	// Difficulty target.
	Difficulty big.Int
}

func GetIdForEpoch(startBlockHash [32]byte, startHeight uint64) string {
	return strconv.FormatUint(uint64(startHeight), 10) + "_" + hex.EncodeToString(startBlockHash[:])
}

// The epoch unique ID is the height ++ startblockhash.
func (e *Epoch) GetId() string {
	return GetIdForEpoch(e.StartBlockHash, e.StartHeight)
}

type PeerConfig struct {
	address        string
	port           string
	bootstrapPeers []string
}

func NewPeerConfig(address string, port string, bootstrapPeers []string) PeerConfig {
	return PeerConfig{address: address, port: port, bootstrapPeers: bootstrapPeers}
}

type NetworkMessage struct {
	Type string `json:"type"`
}

type HeartbeatMesage struct {
	Type                string `json:"type"` // "heartbeat"
	TipHash             string `json:"tipHash"`
	TipHeight           int    `json:"tipHeight"`
	ClientVersion       string `json:"clientVersion"`
	WireProtocolVersion uint   `json:"wireProtocolVersion"`
	ClientAddress       string `json:"clientAddress"`
	// TODO add chain/network ID.
	Time time.Time
}

// get_tip
type GetTipMessage struct {
	Type string      `json:"type"` // "get_tip"
	Tip  BlockHeader `json:"tip"`
}

// new_block
type NewBlockMessage struct {
	Type     string   `json:"type"` // "new_block"
	RawBlock RawBlock `json:"rawBlock"`
}

// new_transaction
type NewTransactionMessage struct {
	Type           string         `json:"type"` // "new_transaction"
	RawTransaction RawTransaction `json:"rawTransaction"`
}

// get_blocks
type GetBlocksMessage struct {
	Type        string   `json:"type"` // "get_blocks"
	BlockHashes []string `json:"blockHashes"`
}

type GetBlocksReply struct {
	Type          string   `json:"type"` // "get_blocks_reply"
	RawBlockDatas [][]byte `json:"rawBlockDatas"`
}

// has_block
type HasBlockMessage struct {
	Type      string `json:"type"` // "have_block"
	BlockHash string `json:"blockHash"`
}

type HasBlockReply struct {
	Type string `json:"type"` // "have_block_reply"
	Has  bool   `json:"has"`
}

// gossip_peers
type GossipPeersMessage struct {
	Type  string   `json:"type"` // "gossip_peers"
	Peers []string `json:"myPeers"`
}
