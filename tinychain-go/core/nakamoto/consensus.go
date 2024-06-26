// Nakamoto consensus.

package nakamoto

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"crypto/sha256"
	"fmt"
	"time"
)

type ConsensusConfig struct {
	// The length of an epoch.
	EpochLengthBlocks uint64

	// The target block production rate in terms of 1 epoch.
	TargetEpochLengthMillis uint64

	// Initial difficulty target.
	InitialDifficulty big.Int
}

// A raw block is the block as transmitted on the network.
// It contains the block header and the block body.
// It does not contain any block metadata such as height, epoch, or difficulty.
type RawBlock struct {
	// Block header.
	ParentHash [32]byte
	Timestamp uint64
	NumTransactions uint64
	TransactionsMerkleRoot [32]byte
	Nonce [32]byte
	
	// Block body.
	Transactions []RawTransaction
}

func (b *RawBlock) SetNonce(i big.Int) {
	nonce := make([]byte, 32)
	i.FillBytes(nonce)
	b.Nonce = sha256.Sum256(nonce)
}

func (b *RawBlock) Envelope() []byte {
	// Encode canonically.
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, b.ParentHash)
	if err != nil { panic(err); }
	err = binary.Write(buf, binary.BigEndian, b.Timestamp)
	err = binary.Write(buf, binary.BigEndian, b.NumTransactions)
	if err != nil { panic(err); }
	err = binary.Write(buf, binary.BigEndian, b.TransactionsMerkleRoot)
	if err != nil { panic(err); }
	err = binary.Write(buf, binary.BigEndian, b.Nonce)
	if err != nil { panic(err); }

	return buf.Bytes()
}

func (b *RawBlock) Hash() [32]byte {
	// Hash the envelope.
	h := sha256.New()
	h.Write(b.Envelope())
	return sha256.Sum256(h.Sum(nil))
}

type RawTransaction struct {
	Sig [64]byte
	FromPubkey [32]byte
	Data []byte
}

func (tx *RawTransaction) Envelope() []byte {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, tx.Sig)
	if err != nil { panic(err); }
	err = binary.Write(buf, binary.BigEndian, tx.Data)
	if err != nil { panic(err); }

	return buf.Bytes()
}

func VerifyPOW(blockhash [32]byte, target big.Int) bool {
	hash := new(big.Int).SetBytes(blockhash[:])
	return hash.Cmp(&target) == -1
}

func SolvePOW(b RawBlock, startNonce big.Int, target big.Int, maxIterations uint64) (big.Int, error) {
	block := b
	nonce := startNonce
	var i uint64 = 0

	for {
		i++
		
		// Exit if iterations is reached.
		if maxIterations < i {
			return big.Int{}, fmt.Errorf("Solution not found in %d iterations", maxIterations)
		}

		// Increment nonce.
		nonce.Add(&nonce, big.NewInt(1))
		block.SetNonce(nonce)

		// Hash.
		h := block.Hash()
		hash := new(big.Int).SetBytes(h[:])

		// Check solution: hash < target.
		if hash.Cmp(&target) == -1 {
			fmt.Printf("Solved in %d iterations\n", i)
			fmt.Printf("Hash: %x\n", hash.String())
			fmt.Printf("Nonce: %x\n", nonce.String())
			return nonce, nil
		}
	}
}

func Timestamp() uint64 {
	now := time.Now()
	milliseconds := now.UnixMilli()
	return uint64(milliseconds)
}

