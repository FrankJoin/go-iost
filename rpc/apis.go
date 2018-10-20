package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/libp2p/go-libp2p-peer"
	"google.golang.org/grpc"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/cverifier"
	"github.com/iost-official/go-iost/consensus/pob"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/core/event"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/verifier"
	"github.com/iost-official/go-iost/vm/database"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/go-iost/new_rpc ApisServer

// GRPCServer GRPC rpc server
type GRPCServer struct {
	bc         blockcache.BlockCache
	p2pService p2p.Service
	txpool     txpool.TxPool
	bchain     block.Chain
	forkDB     db.MVCCDB
	visitor    *database.Visitor
	port       int
	bv         global.BaseVariable
}

// NewRPCServer create GRPC rpc server
func NewRPCServer(tp txpool.TxPool, bcache blockcache.BlockCache, _global global.BaseVariable, p2pService p2p.Service) *GRPCServer {
	forkDb := _global.StateDB().Fork()
	return &GRPCServer{
		p2pService: p2pService,
		txpool:     tp,
		bchain:     _global.BlockChain(),
		bc:         bcache,
		forkDB:     forkDb,
		visitor:    database.NewVisitor(0, forkDb),
		port:       _global.Config().RPC.GRPCPort,
		bv:         _global,
	}
}

// Start start GRPC server
func (s *GRPCServer) Start() error {
	port := strconv.Itoa(s.port)
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp4", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	if s == nil {
		return fmt.Errorf("failed to rpc NewServer")
	}

	RegisterApisServer(server, s)
	go server.Serve(lis)
	ilog.Info("RPCServer Start")
	return nil
}

// Stop stop GRPC server
func (s *GRPCServer) Stop() {
	return
}

// GetNodeInfo return the node info
func (s *GRPCServer) GetNodeInfo(ctx context.Context, empty *empty.Empty) (*NodeInfoRes, error) {
	netService, ok := s.p2pService.(*p2p.NetService)
	if !ok {
		return nil, fmt.Errorf("internal error: netService type conversion failed")
	}
	neighbors := netService.GetNeighbors()
	res := &NodeInfoRes{}
	res.Network = &NetworkInfo{}
	res.Network.PeerInfo = make([]*PeerInfo, 0)
	neighbors.Range(func(k, v interface{}) bool {
		res.Network.PeerInfo = append(res.Network.PeerInfo, &PeerInfo{ID: k.(peer.ID).Pretty(), Addr: v.(*p2p.Peer).GetAddr()})
		return true
	})
	res.Network.PeerCount = (int32)(len(res.Network.PeerInfo))
	res.Network.ID = s.p2pService.ID()
	res.GitHash = global.GitHash
	res.BuildTime = global.BuildTime
	res.Mode = s.bv.Mode().String()
	return res, nil
}

// GetChainInfo return the chain info
func (s *GRPCServer) GetChainInfo(ctx context.Context, empty *empty.Empty) (*ChainInfoRes, error) {
	return &ChainInfoRes{
		NetType:              s.bv.Config().Version.NetType,
		ProtocolVersion:      s.bv.Config().Version.ProtocolVersion,
		Height:               s.bchain.Length() - 1,
		WitnessList:          pob.GetStaticProperty().WitnessList,
		HeadBlock:            toBlockInfo(s.bc.Head(), false),
		LatestConfirmedBlock: toBlockInfo(s.bc.LinkedRoot(), false),
	}, nil
}

// GetTxByHash get tx by transaction hash
func (s *GRPCServer) GetTxByHash(ctx context.Context, hash *HashReq) (*TxRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash
	txHashBytes := common.Base58Decode(txHash)

	trx, err := s.bchain.GetTx(txHashBytes)

	if err != nil {
		return nil, err
	}

	return &TxRes{
		TxRaw: trx.ToTxRaw(),
		Hash:  common.Base58Encode(trx.Hash()),
	}, nil
}

// GetTxReceiptByHash get receipt by receipt hash
func (s *GRPCServer) GetTxReceiptByHash(ctx context.Context, hash *HashReq) (*TxReceiptRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	receiptHash := hash.Hash
	receiptHashBytes := common.Base58Decode(receiptHash)

	receipt, err := s.bchain.GetReceipt(receiptHashBytes)
	if err != nil {
		return nil, err
	}

	return &TxReceiptRes{
		TxReceiptRaw: receipt.ToTxReceiptRaw(),
		Hash:         common.Base58Encode(receiptHashBytes),
	}, nil
}

// GetTxReceiptByTxHash get receipt by transaction hash
func (s *GRPCServer) GetTxReceiptByTxHash(ctx context.Context, hash *HashReq) (*TxReceiptRes, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash
	txHashBytes := common.Base58Decode(txHash)

	receipt, err := s.bchain.GetReceiptByTxHash(txHashBytes)
	if err != nil {
		return nil, err
	}

	return &TxReceiptRes{
		TxReceiptRaw: receipt.ToTxReceiptRaw(),
		Hash:         common.Base58Encode(receipt.Hash()),
	}, nil
}

func toBlockInfo(blk *block.Block, complete bool) *BlockInfo {
	blkInfo := &BlockInfo{
		Head:   blk.Head,
		Hash:   common.Base58Encode(blk.HeadHash()),
		Txs:    make([]*tx.TxRaw, 0),
		Txhash: make([]string, 0),
	}
	for _, trx := range blk.Txs {
		if complete {
			blkInfo.Txs = append(blkInfo.Txs, trx.ToTxRaw())
		}
		blkInfo.Txhash = append(blkInfo.Txhash, common.Base58Encode(trx.Hash()))
	}
	for _, receipt := range blk.Receipts {
		if complete {
			blkInfo.Receipts = append(blkInfo.Receipts, receipt.ToTxReceiptRaw())
		}
		blkInfo.ReceiptHash = append(blkInfo.ReceiptHash, common.Base58Encode(receipt.Hash()))
	}
	return blkInfo
}

// GetBlockByHash get block by block head hash
func (s *GRPCServer) GetBlockByHash(ctx context.Context, blkHashReq *BlockByHashReq) (*BlockInfo, error) {
	if blkHashReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	hash := blkHashReq.Hash
	hashBytes := common.Base58Decode(hash)
	complete := blkHashReq.Complete

	blk, _ := s.bchain.GetBlockByHash(hashBytes)
	if blk == nil {
		blk, _ = s.bc.GetBlockByHash(hashBytes)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := toBlockInfo(blk, complete)
	return blkInfo, nil
}

// GetBlockByNum get block by block number
func (s *GRPCServer) GetBlockByNum(ctx context.Context, blkNumReq *BlockByNumReq) (*BlockInfo, error) {
	if blkNumReq == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	num := blkNumReq.Num
	complete := blkNumReq.Complete

	blk, _ := s.bchain.GetBlockByNumber(num)
	if blk == nil {
		blk, _ = s.bc.GetBlockByNumber(num)
	}
	if blk == nil {
		return nil, fmt.Errorf("cant find the block")
	}
	blkInfo := toBlockInfo(blk, complete)
	return blkInfo, nil
}

// GetContractStorage get contract storage from state db
func (s *GRPCServer) GetContractStorage(ctx context.Context, req *GetContractStorageReq) (*GetContractStorageRes, error) {
	if req == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if req.ContractID == "" {
		return nil, fmt.Errorf("contract id cannot be empty")
	}
	s.forkDB.Checkout(string(s.bc.LinkedRoot().HeadHash()))
	var value string
	if req.Field == "" {
		k := req.ContractID + database.Separator + req.Key
		value = s.visitor.BasicHandler.Get(k)
	} else {
		k := req.ContractID + database.Separator + req.Key
		value = s.visitor.MapHandler.MGet(k, req.Field)
	}
	data, err := json.Marshal(database.Unmarshal(value))
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal %v", value)
	}
	return &GetContractStorageRes{JsonStr: string(data)}, nil
}

// GetContract return a contract by contract id
func (s *GRPCServer) GetContract(ctx context.Context, key *GetContractReq) (*GetContractRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if key.ContractID == "" {
		return nil, fmt.Errorf("argument cannot be empty string")
	}
	if !strings.HasPrefix(key.ContractID, "Contract") {
		return nil, fmt.Errorf("Contract id should start with \"Contract\"")
	}

	txHashBytes := common.Base58Decode(key.ContractID[len("Contract"):])
	trx, err := s.bchain.GetTx(txHashBytes)

	if err != nil {
		return nil, err
	}
	// assume only one 'SetCode' action
	txActionName := trx.Actions[0].ActionName
	if trx.Actions[0].Contract != "iost.system" || txActionName != "SetCode" && txActionName != "UpdateCode" {
		return nil, fmt.Errorf("Not a SetCode or Update transaction")
	}
	js, err := simplejson.NewJson([]byte(trx.Actions[0].Data))
	if err != nil {
		return nil, err
	}
	contractStr, err := js.GetIndex(0).String()
	if err != nil {
		return nil, err
	}
	contract := &contract.Contract{}
	err = contract.B64Decode(contractStr)
	if err != nil {
		return nil, err
	}
	return &GetContractRes{Value: contract}, nil
	//return &GetContractRes{Value: s.visitor.Contract(key.Key)}, nil

}

// GetBalance get account balance
func (s *GRPCServer) GetBalance(ctx context.Context, key *GetBalanceReq) (*GetBalanceRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	if key.UseLongestChain {
		s.forkDB.Checkout(string(s.bc.Head().HeadHash())) // long
	} else {
		s.forkDB.Checkout(string(s.bc.LinkedRoot().HeadHash())) // confirm
	}
	return &GetBalanceRes{
		Balance: s.visitor.Balance(key.ID),
	}, nil
}

// SendRawTx send transaction to blockchain
func (s *GRPCServer) SendRawTx(ctx context.Context, rawTx *RawTxReq) (*SendRawTxRes, error) {
	if rawTx == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	err := trx.Decode(rawTx.Data)
	if err != nil {
		return nil, err
	}
	// add servi
	//tx.RecordTx(trx, tx.Data.Self())
	ret := s.txpool.AddTx(&trx)
	switch ret {
	case txpool.TimeError:
		return nil, fmt.Errorf("tx err:%v", "TimeError")
	case txpool.VerifyError:
		return nil, fmt.Errorf("tx err:%v", "VerifyError")
	case txpool.DupError:
		return nil, fmt.Errorf("tx err:%v", "DupError")
	case txpool.GasPriceError:
		return nil, fmt.Errorf("tx err:%v", "GasPriceError")
	case txpool.CacheFullError:
		return nil, fmt.Errorf("tx err:%v", "CacheFullError")
	default:
	}
	res := SendRawTxRes{}
	res.Hash = string(trx.Hash())
	return &res, nil
}

// ExecTx only exec the tx, but not put it onto chain
func (s *GRPCServer) ExecTx(ctx context.Context, rawTx *RawTxReq) (*ExecTxRes, error) {
	if rawTx == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	err := trx.Decode(rawTx.Data)
	if err != nil {
		return nil, err
	}
	_, head := s.txpool.TxIterator()
	topBlock := head.Block
	blk := block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: topBlock.HeadHash(),
			Number:     topBlock.Head.Number + 1,
			Witness:    "",
			Time:       time.Now().Unix() / common.SlotLength,
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	v := verifier.Verifier{}
	receipt, err := v.Try(blk.Head, s.forkDB, &trx, cverifier.TxExecTimeLimit)
	if err != nil {
		return nil, fmt.Errorf("exec tx failed: %v", err)
	}
	return &ExecTxRes{TxReceiptRaw: receipt.ToTxReceiptRaw()}, nil
}

// Subscribe used for event
func (s *GRPCServer) Subscribe(req *SubscribeReq, res Apis_SubscribeServer) error {
	ec := event.GetEventCollectorInstance()
	sub := event.NewSubscription(100, req.Topics)
	ec.Subscribe(sub)
	defer ec.Unsubscribe(sub)

	timerChan := time.NewTicker(time.Minute).C
forloop:
	for {
		select {
		case <-timerChan:
			ilog.Debugf("timeup in subscribe send")
			break forloop
		case ev := <-sub.ReadChan():
			err := res.Send(&SubscribeRes{Ev: ev})
			if err != nil {
				return err
			}
		default:
		}
	}
	return nil
}
