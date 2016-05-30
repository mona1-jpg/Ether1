package ens

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/swarm/services/ens/contract"
)

func init() {
	glog.SetV(6)
	glog.SetToStderr(true)
}

var (
	key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	name   = "my name on ENS"
	hash   = crypto.Sha3Hash([]byte("my content"))
	addr   = crypto.PubkeyToAddress(key.PublicKey)
)

func deploy(prvKey *ecdsa.PrivateKey, amount *big.Int, backend *backends.SimulatedBackend) (common.Address, error) {
	deployTransactor := bind.NewKeyedTransactor(prvKey)
	deployTransactor.Value = amount
	addr, _, _, err := contract.DeployResolver(deployTransactor, backend)
	if err != nil {
		return common.Address{}, err
	}
	backend.Commit()
	return addr, nil
}

func TestENS(t *testing.T) {
	contractBackend := backends.NewSimulatedBackend(core.GenesisAccount{addr, big.NewInt(1000000000)})
	transactOpts := bind.NewKeyedTransactor(key)
	contractAddr, err := deploy(key, big.NewInt(0), contractBackend)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	resolverAddr, err := deploy(key, big.NewInt(0), contractBackend)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ens := NewENS(transactOpts, contractAddr, contractBackend)
	_, err = ens.Register(name, resolverAddr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	contractBackend.Commit()

	_, err = ens.SetContentHash(name, hash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	contractBackend.Commit()

	vhost, err := ens.Resolve(name)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if vhost.Hex() != hash.Hex()[2:] {
		t.Fatalf("resolve error, expected %v, got %v", hash.Hex(), vhost)
		// t.Fatalf("resolve error, expected %v, got %v", transactOpts.From, hash)
	}

}