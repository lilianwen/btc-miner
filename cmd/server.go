/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import "C"
import (
	"btcnetwork/common"
	"btcnetwork/miner"
	"btcnetwork/node"
	"btcnetwork/storage"
	"fmt"
	"github.com/btcsuite/btclog"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	sigs               chan os.Signal
	log                btclog.Logger
	exitingFlag        int32
	defaultLogFilename = "btcminer.log"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "service for btc network",
	Long:  `service for btc network`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig(cmd, args)
		runServer(cfg)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//serverCmd.PersistentFlags().BoolP("mine", "m", false, "mine blocks")
	serverCmd.PersistentFlags().StringP("config", "c", "", "config file for node")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runServer(cfg *common.Config) {
	var wg sync.WaitGroup

	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go sigHandler(&wg)

	storage.Start(cfg) //启动存储服务
	node.Start(cfg)
	miner.Start(cfg)

	wg.Wait()
}

func sigHandler(wg *sync.WaitGroup) {
	defer wg.Done()
	sig := <-sigs
	log.Info("acquire signal:", sig)
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		stopServer()
	default:
		log.Error("unsupport signal handle:", sig)
	}
}

func stopServer() {
	//防止多次按下ctrl+C导致多次执行这个函数
	if atomic.AddInt32(&exitingFlag, 1) != 1 {
		return
	}
	log.Info("stoping server...")

	miner.Stop()
	node.Stop()
	storage.Stop()
	close(sigs)
}

func loadConfig(cmd *cobra.Command, args []string) *common.Config {
	_ = args
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	s := strings.Split(configFile, ".")
	if len(s) <= 1 {
		panic("config file without extension")
	}

	cfgType := s[len(s)-1]
	viper.SetConfigFile(configFile)
	viper.SetConfigType(cfgType)
	fmt.Println("read data form config file:", configFile)
	if err = viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var c = common.Config{}
	if err = viper.Unmarshal(&c); err != nil {
		panic(err)
	}

	//初始化日志系统
	initLogRotator(filepath.Join(c.LogDir, defaultLogFilename))
	setLogLevels(c.LogLevel)

	return &c
}
