package verifier

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/monitor"
)

type SnapshotVerifier struct {
	reader chain.Chain
}

func NewSnapshotVerifier() *SnapshotVerifier {
	// todo add chain chain
	verifier := &SnapshotVerifier{}
	return verifier
}

func (self *SnapshotVerifier) verifySelf(block *ledger.SnapshotBlock, stat *SnapshotBlockVerifyStat) error {
	defer monitor.LogTime("verify", "snapshotSelf", time.Now())

	if block.Height == types.GenesisHeight {
		snapshotBlock := ledger.GetGenesisSnapshotBlock()
		if block.Hash != snapshotBlock.Hash {
			stat.result = FAIL
			return errors.New("genesis block error.")
		}
	}
	return nil
}

func (self *SnapshotVerifier) verifyAccounts(block *ledger.SnapshotBlock, stat *SnapshotBlockVerifyStat) error {
	defer monitor.LogTime("verify", "snapshotAccounts", time.Now())
	for addr, b := range block.SnapshotContent {
		hash, e := self.reader.GetAccountBlockHashByHeight(&addr, b.AccountBlockHeight)
		if e != nil {
			return e
		}
		if hash == nil {
			stat.results[addr] = PENDING
		} else if *hash == b.AccountBlockHash {
			stat.results[addr] = SUCCESS
		} else {
			stat.results[addr] = FAIL
			stat.result = FAIL
			return errors.New(fmt.Sprintf("account[%s] fork, height:[%d], hash:[%s]",
				addr.String(), b.AccountBlockHeight, b.AccountBlockHash))
		}
	}
	return nil
}

func (self *SnapshotVerifier) verifyAccountsTimeout(block *ledger.SnapshotBlock, stat *SnapshotBlockVerifyStat) error {
	defer monitor.LogTime("verify", "snapshotAccountsTimeout", time.Now())
	head := self.reader.GetLatestSnapshotBlock()
	if head.Height != block.Height-1 {
		return errors.New("snapshot pending for height:" + strconv.FormatUint(head.Height, 10))
	}
	if head.Hash != block.PrevHash {
		return errors.New(fmt.Sprintf("block is not next. prevHash:%s, headHash:%s", block.PrevHash, head.Hash))
	}

	for addr, _ := range block.SnapshotContent {
		err := self.VerifyAccountTimeout(addr, block.Height)
		if err != nil {
			stat.result = FAIL
			return err
		}
	}
	return nil
}

func (self *SnapshotVerifier) VerifyAccountTimeout(addr types.Address, snapshotHeight uint64) error {
	defer monitor.LogTime("verify", "accountTimeout", time.Now())
	first, e := self.reader.GetFirstConfirmedAccountBlockBySbHeight(snapshotHeight, &addr)
	if e != nil {
		return e
	}
	if first == nil {
		return errors.New("account block is nil.")
	}
	refer, e := self.reader.GetSnapshotBlockByHash(&first.SnapshotHash)

	if e != nil {
		return e
	}
	if refer == nil {
		return errors.New("snapshot block is nil.")
	}

	ok := self.VerifyTimeout(snapshotHeight, refer.Height)
	if !ok {
		return errors.New("snapshot account block timeout.")
	}
	return nil
}

func (self *SnapshotVerifier) VerifyTimeout(nowHeight uint64, referHeight uint64) bool {
	if nowHeight-referHeight > 60*60*24 {
		return false
	}
	return true
}

func (self *SnapshotVerifier) VerifyReferred(block *ledger.SnapshotBlock) *SnapshotBlockVerifyStat {
	defer monitor.LogTime("verify", "snapshotBlock", time.Now())
	stat := self.newVerifyStat(block)

	err := self.verifySelf(block, stat)
	if err != nil {
		stat.errMsg = err.Error()
		return stat
	}
	err = self.verifyAccountsTimeout(block, stat)
	if err != nil {
		stat.errMsg = err.Error()
		return stat
	}

	err = self.verifyAccounts(block, stat)
	if err != nil {
		stat.errMsg = err.Error()
		return stat
	}

	return stat
}
func (self *SnapshotVerifier) VerifyProducer(block *ledger.SnapshotBlock) *SnapshotBlockVerifyStat {
	defer monitor.LogTime("verify", "snapshotProducer", time.Now())
	stat := self.newVerifyStat(block)
	return stat
}

type AccountHashH struct {
	Addr   *types.Address
	Hash   *types.Hash
	Height *big.Int
}

type SnapshotBlockVerifyStat struct {
	result       VerifyResult
	results      map[types.Address]VerifyResult
	errMsg       string
	accountTasks []*AccountPendingTask
	snapshotTask *SnapshotPendingTask
}

func (self *SnapshotBlockVerifyStat) AccountTasks() []*AccountPendingTask {
	return nil
}
func (self *SnapshotBlockVerifyStat) SnapshotTask() []*SnapshotPendingTask {
	return nil
}

func (self *SnapshotBlockVerifyStat) ErrMsg() string {
	return self.errMsg
}

func (self *SnapshotBlockVerifyStat) VerifyResult() VerifyResult {
	return self.result
}

func (self *SnapshotVerifier) newVerifyStat(b *ledger.SnapshotBlock) *SnapshotBlockVerifyStat {
	// todo init account hashH
	stat := &SnapshotBlockVerifyStat{result: PENDING}
	stat.results = make(map[types.Address]VerifyResult)
	return stat
}