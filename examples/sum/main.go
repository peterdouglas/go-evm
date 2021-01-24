/***************************************************************************
 *
 * Copyright (c) 2017 Baidu.com, Inc. All Rights Reserved
 * @author duanbing(duanbing@baidu.com)
 *
 **************************************************************************/

/**
 * @filename main.go
 * @desc
 * @create time 2018-04-19 15:49:26
**/
package main

import (
	"fmt"
	ec "github.com/duanbing/go-evm/core"
	"github.com/duanbing/go-evm/state"
	"github.com/duanbing/go-evm/vm"
	"time"

	"math/big"

	"github.com/duanbing/go-evm/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

var (
	//testHash    = common.StringToHash("duanbing")
	testAddress = common.StringToAddress("duanbing")
	toAddress   = common.StringToAddress("andone")
	amount      = big.NewInt(1)
	nonce       = uint64(0)
	gasLimit    = big.NewInt(100000)
	//generated by example
	codeStr = "0x6060604052341561000f57600080fd5b60b18061001d6000396000f300606060405260043610603f576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063c6888fa1146044575b600080fd5b3415604e57600080fd5b606260048080359060200190919050506078565b6040518082815260200191505060405180910390f35b60006007820290509190505600a165627a7a72305820c4ac950a92caa9944a7e07e030542e9ed7db92631adcc234d86a105c853b81a20029"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	data, err := hexutil.Decode(codeStr)
	must(err)
	msg := ec.NewMessage(testAddress, &toAddress, nonce, amount, gasLimit, big.NewInt(1), data, false)
	header := types.Header{
		// ParentHash: common.Hash{},
		// UncleHash:  common.Hash{},
		Coinbase: common.HexToAddress("0x0000000000000000000000000000000000000000"),
		//	Root:        common.Hash{},
		//	TxHash:      common.Hash{},
		//	ReceiptHash: common.Hash{},
		//	Bloom:      types.BytesToBloom([]byte("duanbing")),
		Difficulty: big.NewInt(1),
		Number:     big.NewInt(1),
		GasLimit:   gasLimit,
		GasUsed:    big.NewInt(1),
		Time:       big.NewInt(time.Now().Unix()),
		Extra:      nil,
		//MixDigest:  testHash,
		//Nonce:      types.EncodeNonce(1),
	}
	cc := ChainContext{}
	ctx := ec.NewEVMContext(msg, &header, cc, &testAddress)
	mdb, err := ethdb.NewMemDatabase()
	must(err)
	db := state.NewDatabase(mdb)
	statedb, err := state.New(common.Hash{}, db)
	//set balance
	statedb.GetOrNewStateObject(testAddress)
	statedb.GetOrNewStateObject(toAddress)
	statedb.AddBalance(testAddress, big.NewInt(1e18))
	testBalance := statedb.GetBalance(testAddress)
	fmt.Println("testBalance =", testBalance)
	must(err)

	//	config := params.TestnetChainConfig
	config := params.AllProtocolChanges
	logConfig := vm.LogConfig{}
	structLogger := vm.NewStructLogger(&logConfig)
	vmConfig := vm.Config{Debug: true, Tracer: structLogger, DisableGasMetering: false /*, JumpTable: vm.NewByzantiumInstructionSet()*/}
	//vmConfig := vm.Config{DisableGasMetering: false}

	evm := vm.NewEVM(ctx, statedb, config, vmConfig)
	contractRef := vm.AccountRef(testAddress)
	contractCode, _, gasLeftover, vmerr := evm.Create(contractRef, data, statedb.GetBalance(testAddress).Uint64(), big.NewInt(0))
	must(vmerr)
	statedb.SetBalance(testAddress, big.NewInt(0).SetUint64(gasLeftover))
	testBalance = statedb.GetBalance(testAddress)
	fmt.Println("after create contract, testBalance =", testBalance)
	// set input ,  formatted accocding to https://solidity.readthedocs.io/en/develop/abi-spec.html
	//encode methods := "multiply(uint)"
	inttypes, err := abi.NewType("uint")
	must(err)
	methods := abi.Method{
		Name:    "multiply",
		Const:   false,
		Inputs:  []abi.Argument{abi.Argument{Name: "a", Type: inttypes}},
		Outputs: []abi.Argument{abi.Argument{Name: "d", Type: inttypes}},
	}
	fmt.Println(hexutil.Encode(methods.Id()))

	//encode params := "0xa"
	pm := common.BigToHash(big.NewInt(-10)).Hex()
	fmt.Println(pm)
	//concat method and params
	inputstr := hexutil.Encode(methods.Id()) + pm[2:]
	input, err := hexutil.Decode(inputstr)
	must(err)
	fmt.Println("begin to exec contract")
	statedb.SetCode(testAddress, contractCode)
	outputs, gasLeftover, vmerr := evm.Call(contractRef, testAddress, input, statedb.GetBalance(testAddress).Uint64(), big.NewInt(0))
	must(vmerr)

	statedb.SetBalance(testAddress, big.NewInt(0).SetUint64(gasLeftover))
	testBalance = statedb.GetBalance(testAddress)
	fmt.Println("after call contract, testBalance =", testBalance)
	fmt.Printf("Output %#v\n", hexutil.Encode(outputs))

}

type ChainContext struct{}

func (cc ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	fmt.Println("(cc ChainContext) GetHeader(hash common.Hash, number uint64)")
	return nil
	//return &header
}
