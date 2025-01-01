package nakamoto

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/liamzebedee/tinychain-go/core"
	"github.com/triplewz/poseidon"

	"github.com/stretchr/testify/assert"
)

func TestPOWProofOfWorkSolver(t *testing.T) {
	// create a genesis block
	genesis_block := RawBlock{}
	nonce := new(big.Int)
	target := new(big.Int)
	target.SetString("0000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

	solution, err := SolvePOW(genesis_block, *nonce, *target, 1000000)
	if err != nil {
		t.Fatalf("Failed to solve proof of work")
	}
	fmt.Printf("Solution: %x\n", solution.String())
}

func TestPOWBuildChainOfBlocks(t *testing.T) {
	assert := assert.New(t)

	// Build a chain of 6 blocks.
	chain := make([]RawBlock, 0)
	curr_block := RawBlock{}

	// Fixed target for test.
	target := new(big.Int)
	target.SetString("0fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

	for {
		fmt.Printf("Mining block %x\n", curr_block.Hash())
		solution, err := SolvePOW(curr_block, *new(big.Int), *target, 100000000000)
		if err != nil {
			assert.Nil(t, err)
		}
		fmt.Printf("Solution: %x\n", solution.String())

		// Seal the block.
		curr_block.SetNonce(solution)

		// Append the block to the chain.
		chain = append(chain, curr_block)

		// Create a new block.
		timestamp := uint64(0)
		curr_block = RawBlock{
			ParentHash:      curr_block.Hash(),
			Timestamp:       timestamp,
			NumTransactions: 0,
			Transactions:    []RawTransaction{},
		}

		// Exit if the chain is long enough.
		if len(chain) >= 6 {
			break
		}
	}
}

// func TestSomething(t *testing.T) {
// 	// Setup the configuration for a consensus epoch.
// 	genesis_difficulty := new(big.Int)
// 	genesis_difficulty.SetString("0fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

// 	conf := ConsensusConfig{
// 		EpochLengthBlocks: 5,
// 		TargetEpochLengthMillis: 2000,
// 		GenesisDifficulty: *genesis_difficulty,
// 		GenesisParentBlockHash: [32]byte{},
// 		MaxBlockSizeBytes: 1000000,
// 	}
// 	difficulty := conf.GenesisDifficulty

// 	// Now mine 2 epochs worth of blocks.
// 	chain := make([]RawBlock, 0)
// 	curr_block := RawBlock{}
// 	for {
// 		fmt.Printf("Mining block %x\n", curr_block.Hash())
// 		solution, err := SolvePOW(curr_block, *new(big.Int), difficulty, 100000000000)
// 		if err != nil {
// 			t.Fatalf("Failed to solve proof of work")
// 		}
// 		fmt.Printf("Solution: %x\n", solution.String())

// 		// Seal the block.
// 		curr_block.SetNonce(solution)
// 		curr_block.Timestamp = Timestamp()

// 		// Append the block to the chain.
// 		chain = append(chain, curr_block)

// 		// Create a new block.
// 		curr_block = RawBlock{
// 			ParentHash: curr_block.Hash(),
// 			Timestamp: 0,
// 			NumTransactions: 0,
// 			Transactions: []RawTransaction{},
// 		}

// 		// Recompute the difficulty.
// 		if len(chain) % int(conf.EpochLengthBlocks) == 0 {
// 			// Compute the time taken to mine the last epoch.
// 			epoch_start := chain[len(chain) - int(conf.EpochLengthBlocks)].Timestamp
// 			epoch_end := chain[len(chain) - 1].Timestamp
// 			epoch_duration := epoch_end - epoch_start
// 			if epoch_duration == 0 {
// 				epoch_duration = 1
// 			}
// 			epoch_index := len(chain) / int(conf.EpochLengthBlocks)
// 			fmt.Printf("epoch i=%d start_time=%d end_time=%d duration=%d \n", epoch_index, epoch_start, epoch_end, epoch_duration)

// 			// Compute the target epoch length.
// 			target_epoch_length := conf.TargetEpochLengthMillis * conf.EpochLengthBlocks

// 			// Compute the new difficulty.
// 			// difficulty = difficulty * (epoch_duration / target_epoch_length)
// 			new_difficulty := new(big.Int)
// 			new_difficulty.Mul(&conf.GenesisDifficulty, big.NewInt(int64(epoch_duration)))
// 			new_difficulty.Div(new_difficulty, big.NewInt(int64(target_epoch_length)))

// 			fmt.Printf("New difficulty: %x\n", new_difficulty.String())

// 			// Update the difficulty.
// 			difficulty = *new_difficulty
// 		}

// 		fmt.Printf("Chain length: %d\n", len(chain))
// 		if len(chain) >= 4 * int(conf.EpochLengthBlocks) {
// 			break
// 		}
// 	}
// }

func TestCalculateWork(t *testing.T) {
	diff_target := new(big.Int)
	diff_target.SetString("0fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	// max_diff_target := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)

	acc_work := big.NewInt(0)
	block_template := RawBlock{}

	// https://bitcoin.stackexchange.com/questions/105213/how-is-cumulative-pow-calculated-to-decide-between-competing-chains
	// work = 2^256 / (target + 1)
	// difficulty = max_target / target

	// Solve 30 blocks, adjust difficulty every 10.
	for i := 0; i < 30; i++ {
		solution, err := SolvePOW(block_template, *new(big.Int), *diff_target, 100000000000)
		if err != nil {
			t.Fatalf("Failed to solve proof of work")
		}

		// Seal the block.
		block_template.SetNonce(solution)

		// Setup next block.
		block_template = RawBlock{
			ParentHash: block_template.Hash(),
			Timestamp:  0,
		}

		// Calculate the work.
		work := big.NewInt(2).Exp(big.NewInt(2), big.NewInt(256), nil)
		work.Div(work, big.NewInt(0).Add(diff_target, big.NewInt(1)))
		fmt.Printf("Work: %x\n", work.String())

		acc_work.Add(acc_work, work)
		fmt.Printf("Acc Work: %x\n", acc_work.String())
	}
}

// Poseidon is a ZK-friendly hash function.
//
// This is a benchmark from Starkware's original prover, ethSTARK, which was research done for the Ethereum Foundation.
// https://eprint.iacr.org/2021/582.pdf
//
// 1. Operating-System: Linux 5.3.0-51-generic x86_64.
// 2. CPU: Intel(R) Core(TM) i7-7700K @ 4.20GHz (4 cores, 2 threads per core).
// 3. RAM: 16GB DDR4 (8GB × 2, Speed: 2667 MHz)
// 4. STARK bits of security: 80 bits.
// 5. Hash function used: Rescue (another ZK-friendly hash function).
//
// Proving:
// - Number of hashes: 12,288
// - Prove time: 1s
// Verification:
// - Proof size: ~40kB
// - Verify time: 1.9mS
//
// We note these benchmarks are similar to the Poseidon paper's libsnark implementation.
// Field: BN254
// Proofing system: groth16
// https://eprint.iacr.org/2019/458.pdf
// Prove time: 43.1ms
// Proof size: 200 bytes (3 field elements, 252 bits each, ~3*8 bytes, 24 bytes)
// Verify time: 1.2ms
//
// Interestingly, you can use a ZK proof of revealing a hash preimage as a digital signature.
// How does this compare to digital signatures?
func TestPoseidonHashFunction(t *testing.T) {
	// poseidon hash with 3 input elements and 1 output element.
	input := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}

	// generate round constants for poseidon hash.
	// width=len(input)+1.
	cons, _ := poseidon.GenPoseidonConstants(4)

	// use OptimizedStatic hash mode.
	h1, _ := poseidon.Hash(input, cons, poseidon.OptimizedStatic)
	// use OptimizedDynamic hash mode.
	h2, _ := poseidon.Hash(input, cons, poseidon.OptimizedDynamic)
	// use Correct hash mode.
	h3, _ := poseidon.Hash(input, cons, poseidon.Correct)

	t.Logf("Poseidon hash with OptimizedStatic: %x", h1)
	t.Logf("Poseidon hash with OptimizedDynamic: %x", h2)
	t.Logf("Poseidon hash with Correct: %x", h3)
}

// 213.833µs
func TestECDSASignatureSignTiming(t *testing.T) {
	wallet, err := core.CreateRandomWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %s", err)
	}

	// Measure start time.
	start := time.Now()

	// Sign a message.
	message := []byte("hello world")
	_, err = wallet.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %s", err)
	}

	end := time.Now()
	elapsed := end.Sub(start)

	// Print elapsed time in ms.
	t.Logf("Elapsed time: %v", elapsed)
}

// 125.459µs
func TestECDSASignatureVerifyTiming(t *testing.T) {
	wallet, err := core.CreateRandomWallet()
	if err != nil {
		t.Fatalf("Failed to create wallet: %s", err)
	}

	// Sign a message.
	message := []byte("hello world")
	sig, err := wallet.Sign(message)
	if err != nil {
		t.Fatalf("Failed to sign message: %s", err)
	}

	pubkey := wallet.PubkeyBytes()

	// Measure start time.
	start := time.Now()

	// Verify signature.
	core.VerifySignature(pubkey, sig, message)
	if err != nil {
		t.Fatalf("Failed to verify signature: %s", err)
	}

	end := time.Now()
	elapsed := end.Sub(start)

	// Print elapsed time in ms.
	t.Logf("Elapsed time: %v", elapsed)
}
