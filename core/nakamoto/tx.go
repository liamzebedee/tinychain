package nakamoto

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/liamzebedee/tinychain-go/core"
)

type RawTransaction struct {
	Version    byte     `json:"version"`
	Sig        [64]byte `json:"sig"`
	FromPubkey [65]byte `json:"from"`
	ToPubkey   [65]byte `json:"to"`
	Amount     uint64   `json:"amount"`
	Fee        uint64   `json:"fee"`
	Nonce      uint64   `json:"nonce"`
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

func (tx *Transaction) ToRawTransaction() RawTransaction {
	return RawTransaction{
		Version:    tx.Version,
		Sig:        tx.Sig,
		FromPubkey: tx.FromPubkey,
		ToPubkey:   tx.ToPubkey,
		Amount:     tx.Amount,
		Fee:        tx.Fee,
		Nonce:      tx.Nonce,
	}
}

func (tx *RawTransaction) SizeBytes() uint64 {
	// Size of the transaction is the size of the envelope.
	return 1 + 65 + 65 + 8 + 8 + 8
}

func (tx *RawTransaction) Bytes() []byte {
	buf := make([]byte, 0)
	buf = append(buf, tx.Version)
	buf = append(buf, tx.Sig[:]...)
	buf = append(buf, tx.FromPubkey[:]...)
	buf = append(buf, tx.ToPubkey[:]...)

	amount := make([]byte, 8)
	binary.BigEndian.PutUint64(amount, tx.Amount)
	buf = append(buf, amount...)

	fee := make([]byte, 8)
	binary.BigEndian.PutUint64(fee, tx.Fee)
	buf = append(buf, fee...)

	nonce := make([]byte, 8)
	binary.BigEndian.PutUint64(nonce, tx.Nonce)
	buf = append(buf, nonce...)

	return buf
}

func (tx *RawTransaction) Envelope() []byte {
	buf := make([]byte, 0)
	buf = append(buf, tx.Version)
	buf = append(buf, tx.FromPubkey[:]...)
	buf = append(buf, tx.ToPubkey[:]...)

	amount := make([]byte, 8)
	binary.BigEndian.PutUint64(amount, tx.Amount)
	buf = append(buf, amount...)

	fee := make([]byte, 8)
	binary.BigEndian.PutUint64(fee, tx.Fee)
	buf = append(buf, fee...)

	nonce := make([]byte, 8)
	binary.BigEndian.PutUint64(nonce, tx.Nonce)
	buf = append(buf, nonce...)

	return buf
}

func (tx *RawTransaction) Hash() [32]byte {
	// Hash the envelope.
	h := sha256.New()
	h.Write(tx.Bytes())
	return sha256.Sum256(h.Sum(nil))
}

func MakeTransferTx(from [65]byte, to [65]byte, amount uint64, fee uint64, wallet *core.Wallet) RawTransaction {
	tx := RawTransaction{
		Version:    1,
		Sig:        [64]byte{},
		FromPubkey: from,
		ToPubkey:   to,
		Amount:     amount,
		Fee:        fee,
		Nonce:      0,
	}
	// Sign tx.
	sig, err := wallet.Sign(tx.Envelope())
	if err != nil {
		panic(err)
	}
	copy(tx.Sig[:], sig)
	return tx
}

func GetMerkleRootForTxs(txs []RawTransaction) [32]byte {
	envs := make([][]byte, 0)
	for _, tx := range txs {
		envs = append(envs, tx.Bytes())
	}
	return core.ComputeMerkleHash(envs)
}
