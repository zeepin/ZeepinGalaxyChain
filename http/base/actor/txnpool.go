/*
 * Copyright (C) 2018 The ZeepinChain Authors
 * This file is part of The ZeepinChain library.
 *
 * The ZeepinChain is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ZeepinChain is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ZeepinChain.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package actor privides communication with other actor
package actor

import (
	"errors"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/imZhuFei/zeepin/common"
	"github.com/imZhuFei/zeepin/common/log"
	"github.com/imZhuFei/zeepin/core/types"
	ontErrors "github.com/imZhuFei/zeepin/errors"
	tcomn "github.com/imZhuFei/zeepin/txnpool/common"
)

var txnPid *actor.PID
var txnPoolPid *actor.PID
var DisableSyncVerifyTx = false

func SetTxPid(actr *actor.PID) {
	txnPid = actr
}
func SetTxnPoolPid(actr *actor.PID) {
	txnPoolPid = actr
}

//append transaction to pool to txpool actor
func AppendTxToPool(txn *types.Transaction) (ontErrors.ErrCode, string) {
	if DisableSyncVerifyTx {
		txReq := &tcomn.TxReq{txn, tcomn.HttpSender, nil}
		txnPid.Tell(txReq)
		return ontErrors.ErrNoError, ""
	}
	ch := make(chan *tcomn.TxResult, 1)
	txReq := &tcomn.TxReq{txn, tcomn.HttpSender, ch}
	txnPid.Tell(txReq)
	if msg, ok := <-ch; ok {
		return msg.Err, msg.Desc
	}
	return ontErrors.ErrUnknown, ""
}

//GetTxsFromPool from txpool actor
func GetTxsFromPool(byCount bool) map[common.Uint256]*types.Transaction {
	future := txnPoolPid.RequestFuture(&tcomn.GetTxnPoolReq{ByCount: byCount}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return nil
	}
	txpool, ok := result.(*tcomn.GetTxnPoolRsp)
	if !ok {
		return nil
	}
	txMap := make(map[common.Uint256]*types.Transaction)
	for _, v := range txpool.TxnPool {
		txMap[v.Tx.Hash()] = v.Tx
	}
	return txMap

}

//GetTxFromPool from txpool actor
func GetTxFromPool(hash common.Uint256) (tcomn.TXEntry, error) {

	future := txnPid.RequestFuture(&tcomn.GetTxnReq{hash}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return tcomn.TXEntry{}, err
	}
	rsp, ok := result.(*tcomn.GetTxnRsp)
	if !ok {
		return tcomn.TXEntry{}, errors.New("fail")
	}
	if rsp.Txn == nil {
		return tcomn.TXEntry{}, errors.New("fail")
	}

	future = txnPid.RequestFuture(&tcomn.GetTxnStatusReq{hash}, REQ_TIMEOUT*time.Second)
	result, err = future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return tcomn.TXEntry{}, err
	}
	txStatus, ok := result.(*tcomn.GetTxnStatusRsp)
	if !ok {
		return tcomn.TXEntry{}, errors.New("fail")
	}
	txnEntry := tcomn.TXEntry{rsp.Txn, txStatus.TxStatus}
	return txnEntry, nil
}

//GetTxnCount from txpool actor
func GetTxnCount() ([]uint32, error) {
	future := txnPid.RequestFuture(&tcomn.GetTxnCountReq{}, REQ_TIMEOUT*time.Second)
	result, err := future.Result()
	if err != nil {
		log.Errorf(ERR_ACTOR_COMM, err)
		return []uint32{}, err
	}
	txnCnt, ok := result.(*tcomn.GetTxnCountRsp)
	if !ok {
		return []uint32{}, errors.New("fail")
	}
	return txnCnt.Count, nil
}
