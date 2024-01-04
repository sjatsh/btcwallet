package main

import (
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/sjatsh/btcwallet/chain"
	"github.com/sjatsh/btcwallet/rpc/legacyrpc"
	"github.com/sjatsh/btcwallet/wallet"
	_ "github.com/sjatsh/btcwallet/walletdb/bdb"
	"net"
	"os"
	"time"
)

func main() {
	var loader *wallet.Loader
	// Set up a wallet.
	dir, err := os.MkdirTemp("", "test_wallet")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := loader.UnloadWallet(); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()

	seed, err := hdkeychain.GenerateSeed(hdkeychain.MinSeedBytes)
	if err != nil {
		panic(err)
	}

	pubPass := []byte("hello")
	privPass := []byte("world")

	loader = wallet.NewLoader(
		&chaincfg.TestNet3Params, dir, true, time.Second*10, 250,
		wallet.WithWalletSyncRetryInterval(10*time.Millisecond),
	)

	w, err := loader.CreateNewWallet(pubPass, privPass, seed, time.Now())
	if err != nil {
		panic(err)
	}
	if err := w.Unlock(privPass, time.After(10*time.Minute)); err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp4", ":18332")
	if err != nil {
		panic(err)
	}
	conn, err := chain.NewBitcoindConn(&chain.BitcoindConfig{
		ChainParams: &chaincfg.TestNet3Params,
		Host:        "go.getblock.io/9ebdf5dc1dfe4a63ab89aa1fc8aad2a0",
		User:        "root",
		Pass:        "root",
	})
	if err != nil {
		panic(err)
	}
	legacyServer := legacyrpc.NewServer(&legacyrpc.Options{
		Username: "root",
		Password: "root",
	}, loader, []net.Listener{listener})

	loader.RunAfterLoad(func(w *wallet.Wallet) {
		legacyServer.RegisterWallet(w)
		legacyServer.SetChainServer(conn.NewBitcoindClient())
	})

	select {}
}
