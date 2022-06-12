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

package bind

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/external"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ErrNoChainID is returned whenever the user failed to specify a chain id.
var ErrNoChainID = errors.New("no chain id specified")

// ErrNotAuthorized is returned when an account is not properly unlocked.
var ErrNotAuthorized = errors.New("not authorized to sign this account")

// NewKeyStoreTransactorWithChainID is a utility method to easily create a transaction signer from
// an decrypted key from a keystore.
func NewKeyStoreTransactorWithChainID(keystore *keystore.KeyStore, account accounts.Account, chainID *big.Int, blockNumber *big.Int) (*TransactOpts, error) {
	if chainID == nil {
		return nil, ErrNoChainID
	}
	signer := types.LatestSignerForChainID(func(*big.Int) *big.Int { return chainID })
	return &TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, ErrNotAuthorized
			}
			signature, err := keystore.SignHash(account, signer.Hash(tx, blockNumber).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature, blockNumber)
		},
		Context: context.Background(),
	}, nil
}

// NewKeyedTransactorWithChainID is a utility method to easily create a transaction signer
// from a single private key.
func NewKeyedTransactorWithChainID(key *ecdsa.PrivateKey, chainID *big.Int, blockNumber *big.Int) (*TransactOpts, error) {
	keyAddr := crypto.PubkeyToAddress(key.PublicKey)
	if chainID == nil {
		return nil, ErrNoChainID
	}
	signer := types.LatestSignerForChainID(func(*big.Int) *big.Int { return chainID })
	return &TransactOpts{
		From: keyAddr,
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != keyAddr {
				return nil, ErrNotAuthorized
			}
			signature, err := crypto.Sign(signer.Hash(tx, blockNumber).Bytes(), key)
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature, blockNumber)
		},
		Context: context.Background(),
	}, nil
}

// NewClefTransactor is a utility method to easily create a transaction signer
// with a clef backend.
func NewClefTransactor(clef *external.ExternalSigner, account accounts.Account) *TransactOpts {
	return &TransactOpts{
		From: account.Address,
		Signer: func(address common.Address, transaction *types.Transaction) (*types.Transaction, error) {
			if address != account.Address {
				return nil, ErrNotAuthorized
			}
			return clef.SignTx(account, transaction, nil, new(big.Int)) // Clef enforces its own chain id
		},
		Context: context.Background(),
	}
}
