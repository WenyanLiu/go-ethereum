// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Contains all the wrappers from the bind package.

package geth

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// Signer is an interaface defining the callback when a contract requires a
// method to sign the transaction before submission.
type Signer interface {
	Sign(*Address, *Transaction) (tx *Transaction, _ error)
}

type signer struct {
	sign bind.SignerFn
}

func (s *signer) Sign(addr *Address, unsignedTx *Transaction) (signedTx *Transaction, _ error) {
	log.DebugLog()
	sig, err := s.sign(types.HomesteadSigner{}, addr.address, unsignedTx.tx)
	if err != nil {
		return nil, err
	}
	return &Transaction{sig}, nil
}

// CallOpts is the collection of options to fine tune a contract call request.
type CallOpts struct {
	opts bind.CallOpts
}

// NewCallOpts creates a new option set for contract calls.
func NewCallOpts() *CallOpts {
	log.DebugLog()
	return new(CallOpts)
}

func (opts *CallOpts) IsPending() bool {
	log.DebugLog()
	return opts.opts.Pending
}
func (opts *CallOpts) GetGasLimit() int64 {
	log.DebugLog()
	return 0 /* TODO(karalabe) */ }

// GetContext cannot be reliably implemented without identity preservation (https://github.com/golang/go/issues/16876)
// Even then it's awkward to unpack the subtleties of a Go context out to Java.
// func (opts *CallOpts) GetContext() *Context { log.DebugLog() return &Context{opts.opts.Context} }

func (opts *CallOpts) SetPending(pending bool) {
	log.DebugLog()
	opts.opts.Pending = pending
}
func (opts *CallOpts) SetGasLimit(limit int64) { log.DebugLog() /* TODO(karalabe) */ }
func (opts *CallOpts) SetContext(context *Context) {
	log.DebugLog()
	opts.opts.Context = context.context
}

// TransactOpts is the collection of authorization data required to create a
// valid Ethereum transaction.
type TransactOpts struct {
	opts bind.TransactOpts
}

func (opts *TransactOpts) GetFrom() *Address {
	log.DebugLog()
	return &Address{opts.opts.From}
}
func (opts *TransactOpts) GetNonce() int64 {
	log.DebugLog()
	return opts.opts.Nonce.Int64()
}
func (opts *TransactOpts) GetValue() *BigInt {
	log.DebugLog()
	return &BigInt{opts.opts.Value}
}
func (opts *TransactOpts) GetGasPrice() *BigInt {
	log.DebugLog()
	return &BigInt{opts.opts.GasPrice}
}
func (opts *TransactOpts) GetGasLimit() int64 {
	log.DebugLog()
	return int64(opts.opts.GasLimit)
}

// GetSigner cannot be reliably implemented without identity preservation (https://github.com/golang/go/issues/16876)
// func (opts *TransactOpts) GetSigner() Signer { log.DebugLog() return &signer{opts.opts.Signer} }

// GetContext cannot be reliably implemented without identity preservation (https://github.com/golang/go/issues/16876)
// Even then it's awkward to unpack the subtleties of a Go context out to Java.
//func (opts *TransactOpts) GetContext() *Context { log.DebugLog() return &Context{opts.opts.Context} }

func (opts *TransactOpts) SetFrom(from *Address) {
	log.DebugLog()
	opts.opts.From = from.address
}
func (opts *TransactOpts) SetNonce(nonce int64) {
	log.DebugLog()
	opts.opts.Nonce = big.NewInt(nonce)
}
func (opts *TransactOpts) SetSigner(s Signer) {
	log.DebugLog()
	opts.opts.Signer = func(signer types.Signer, addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
		sig, err := s.Sign(&Address{addr}, &Transaction{tx})
		if err != nil {
			return nil, err
		}
		return sig.tx, nil
	}
}
func (opts *TransactOpts) SetValue(value *BigInt) {
	log.DebugLog()
	opts.opts.Value = value.bigint
}
func (opts *TransactOpts) SetGasPrice(price *BigInt) {
	log.DebugLog()
	opts.opts.GasPrice = price.bigint
}
func (opts *TransactOpts) SetGasLimit(limit int64) {
	log.DebugLog()
	opts.opts.GasLimit = uint64(limit)
}
func (opts *TransactOpts) SetContext(context *Context) {
	log.DebugLog()
	opts.opts.Context = context.context
}

// BoundContract is the base wrapper object that reflects a contract on the
// Ethereum network. It contains a collection of methods that are used by the
// higher level contract bindings to operate.
type BoundContract struct {
	contract *bind.BoundContract
	address  common.Address
	deployer *types.Transaction
}

// DeployContract deploys a contract onto the Ethereum blockchain and binds the
// deployment address with a wrapper.
func DeployContract(opts *TransactOpts, abiJSON string, bytecode []byte, client *EthereumClient, args *Interfaces) (contract *BoundContract, _ error) {
	log.DebugLog()
	// Deploy the contract to the network
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	addr, tx, bound, err := bind.DeployContract(&opts.opts, parsed, common.CopyBytes(bytecode), client.client, args.objects...)
	if err != nil {
		return nil, err
	}
	return &BoundContract{
		contract: bound,
		address:  addr,
		deployer: tx,
	}, nil
}

// BindContract creates a low level contract interface through which calls and
// transactions may be made through.
func BindContract(address *Address, abiJSON string, client *EthereumClient) (contract *BoundContract, _ error) {
	log.DebugLog()
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	return &BoundContract{
		contract: bind.NewBoundContract(address.address, parsed, client.client, client.client, client.client),
		address:  address.address,
	}, nil
}

func (c *BoundContract) GetAddress() *Address {
	log.DebugLog()
	return &Address{c.address}
}
func (c *BoundContract) GetDeployer() *Transaction {
	log.DebugLog()
	if c.deployer == nil {
		return nil
	}
	return &Transaction{c.deployer}
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result.
func (c *BoundContract) Call(opts *CallOpts, out *Interfaces, method string, args *Interfaces) error {
	log.DebugLog()
	if len(out.objects) == 1 {
		result := out.objects[0]
		if err := c.contract.Call(&opts.opts, result, method, args.objects...); err != nil {
			return err
		}
		out.objects[0] = result
	} else {
		results := make([]interface{}, len(out.objects))
		copy(results, out.objects)
		if err := c.contract.Call(&opts.opts, &results, method, args.objects...); err != nil {
			return err
		}
		copy(out.objects, results)
	}
	return nil
}

// Transact invokes the (paid) contract method with params as input values.
func (c *BoundContract) Transact(opts *TransactOpts, method string, args *Interfaces) (tx *Transaction, _ error) {
	log.DebugLog()
	rawTx, err := c.contract.Transact(&opts.opts, method, args.objects...)
	if err != nil {
		return nil, err
	}
	return &Transaction{rawTx}, nil
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (c *BoundContract) Transfer(opts *TransactOpts) (tx *Transaction, _ error) {
	log.DebugLog()
	rawTx, err := c.contract.Transfer(&opts.opts)
	if err != nil {
		return nil, err
	}
	return &Transaction{rawTx}, nil
}
