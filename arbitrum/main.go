package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"math/big"
	"strings"
	"time"

	"claimTool/arbitrum/config"
	"claimTool/arbitrum/resource/contract/distributor"
	"claimTool/arbitrum/resource/contract/erc20"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/ethclient"
)

var cfg *config.ClaimArbConfig

func main() {

	// 获取配置文件路径 config参数
	filePath := flag.String("config", "", "")
	task := flag.String("task", "claim", "")
	flag.Parse()

	cfg = &config.ClaimArbConfig{}
	if err := cfg.LoadFromFile(*filePath); err != nil {
		fmt.Println("解析配置文件错误 ,注意配置文件路径", err.Error())
		return
	}
	fmt.Println("解析配置文件成功")
	fmt.Println()
	if *task == "claim" {
		Claim()
	} else if *task == "collect" {
		if len(cfg.CollectAddress) != 42 {
			fmt.Println("归集地址错误")
			return
		}
		Collect()
	}

}

func Collect() {

	collectAddress := common.HexToAddress(cfg.CollectAddress)
	fmt.Println("本次需要执行", len(cfg.PrivateKeys), "个地址")
	fmt.Println()

	for i, key := range cfg.PrivateKeys {
		// 随便过滤一下
		if len(key) <= 20 {
			continue
		}

		// 去除0x前缀
		key = strings.TrimPrefix(strings.ToUpper(key), "0X")
		privateKey, _ := crypto.HexToECDSA(key)
		publicKeyECDSA, _ := privateKey.Public().(*ecdsa.PublicKey)
		fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
		address := fromAddress.String()

		fmt.Println("开始执行第", i+1, "个地址:", address)

		client, ercInstance, err := GetErcInstance()
		if err != nil {
			fmt.Println("获取合约实例失败", err.Error())
			continue
		}
		fmt.Println("获取合约实例成功")

		auth, err := GetAuth(client, key)
		if err != nil {
			fmt.Println("获取GetAuth失败", err.Error())
			continue
		}
		balance, err := ercInstance.BalanceOf(nil, fromAddress)
		if err != nil {
			fmt.Println("获取arb余额失败", err.Error())
		} else {
			if balance != nil {
				balance = balance.Quo(balance, big.NewInt(1000000000000000000))
			}
			fmt.Println("arb余额:", balance)
		}

		var tx *types.Transaction
		var trErr error
		for i := 0; i < 3; i++ {
			tx, trErr = ercInstance.Transfer(auth, collectAddress, balance)
			if err != nil {
				continue
			}
			break
		}

		if trErr != nil {
			fmt.Println("归集失败", trErr.Error())
		} else {
			fmt.Println("归集成功", "成功哈希", tx.Hash().String())
		}
		fmt.Println()

	}
}

func Claim() {
	fmt.Println("准备进行以太坊时间校验")
	err := GetStart()
	if err != nil {
		fmt.Println("以太坊时间校验错误", err.Error())
	}
	fmt.Println()

	fmt.Println("本次需要执行", len(cfg.PrivateKeys), "个地址")
	fmt.Println()

	for i, key := range cfg.PrivateKeys {
		// 随便过滤一下
		if len(key) <= 20 {
			continue
		}

		// 去除0x前缀
		key = strings.TrimPrefix(strings.ToUpper(key), "0X")
		privateKey, _ := crypto.HexToECDSA(key)
		publicKeyECDSA, _ := privateKey.Public().(*ecdsa.PublicKey)
		fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
		address := fromAddress.String()
		fmt.Println("开始执行第", i+1, "个地址:", address)

		client, disInstance, _, err := GetNewInstance()
		if err != nil {
			fmt.Println("获取合约实例失败", err.Error())
			continue
		}
		fmt.Println("获取合约实例成功")

		auth, err := GetAuth(client, key)
		if err != nil {
			fmt.Println("获取GetAuth失败", err.Error())
			continue
		}

		balance, err := disInstance.ClaimableTokens(nil, fromAddress)
		if err != nil {
			fmt.Println("获取可claim余额失败", err.Error())
		} else {
			if balance != nil {
				balance = balance.Quo(balance, big.NewInt(1000000000000000000))
			}
			fmt.Println("可claim余额:", balance)
		}

		var tx *types.Transaction
		var claimErr error
		for i := 0; i < 3; i++ {
			tx, claimErr = disInstance.Claim(auth)
			if err != nil {
				continue
			}
			break
		}

		if claimErr != nil {
			fmt.Println("执行claim失败", claimErr.Error())
		} else {
			fmt.Println("执行claim成功", "成功哈希", tx.Hash().String())
		}
		fmt.Println()

	}
}

func GetStart() error {
	var err error
	var client *ethclient.Client

	for i := 0; i < 5; i++ {
		client, err = ethclient.Dial(cfg.EthereumRpc)
		if err != nil {
			continue
		}
		break
	}

	if err != nil {
		fmt.Println("获取以太坊客户端失败", err.Error())
		return err
	}
	fmt.Println("获取以太坊客户端成功")

	for {
		number, _ := client.BlockNumber(context.Background())
		fmt.Println("开始Block", cfg.StartBlockNumber, "当前Block", number)
		if int(number) >= cfg.StartBlockNumber {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func GetErcInstance() (*ethclient.Client, *erc20.Contract, error) {
	var err error
	var client *ethclient.Client
	var tokenInstance *erc20.Contract

	for i := 0; i < 3; i++ {
		client, err = ethclient.Dial(cfg.ArbiturmRpc)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		fmt.Println("连接Arbitrum网络失败", err.Error())
		return nil, nil, err
	}
	// 获取claim合约和arb代币合约地址
	tokenAddress := common.HexToAddress(cfg.TokenAddress)
	for i := 0; i < 3; i++ {
		//获取arb代币合约实例
		tokenInstance, err = erc20.NewContract(tokenAddress, client)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		fmt.Println("创造arb代币合约实例失败", err.Error())
		return nil, nil, err
	}

	return client, tokenInstance, nil
}

// GetNewInstance 返回两个合约的实例
func GetNewInstance() (*ethclient.Client, *distributor.Contract, *erc20.Contract, error) {
	var err error
	var client *ethclient.Client
	var disInstance *distributor.Contract
	var tokenInstance *erc20.Contract

	for i := 0; i < 3; i++ {
		client, err = ethclient.Dial(cfg.ArbiturmRpc)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		fmt.Println("连接Arbitrum网络失败", err.Error())
		return nil, nil, nil, err
	}

	// 获取claim合约和arb代币合约地址
	distributorAddress := common.HexToAddress(cfg.DistributorAddress)
	//tokenAddress := common.HexToAddress(cfg.TokenAddress)

	for i := 0; i < 3; i++ {
		//获取claim合约实例
		disInstance, err = distributor.NewContract(distributorAddress, client)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		fmt.Println("创造claim合约实例失败", err.Error())
		return nil, nil, nil, err
	}

	//for i := 0; i < 3; i++ {
	//	//获取arb代币合约实例
	//	tokenInstance, err = erc20.NewContract(tokenAddress, client)
	//	if err != nil {
	//		continue
	//	}
	//  break
	//}
	//if err != nil {
	//	fmt.Println("创造arb代币合约实例失败", err.Error())
	//	return nil, nil, nil, err
	//}

	return client, disInstance, tokenInstance, nil
}

func GetAuth(client *ethclient.Client, key string) (*bind.TransactOpts, error) {
	var nonce uint64
	var err error

	privateKey, _ := crypto.HexToECDSA(key)
	publicKeyECDSA, _ := privateKey.Public().(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	for i := 0; i < 3; i++ {
		nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			continue
		}
		break
	}

	if err != nil {
		fmt.Println("获取nonce失败", err.Error())
		return nil, err
	}
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		price = big.NewInt(int64(cfg.GasPrice))
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)           // in wei
	auth.GasLimit = uint64(cfg.GasLimit) // in units
	auth.GasPrice = price

	return auth, nil
}
