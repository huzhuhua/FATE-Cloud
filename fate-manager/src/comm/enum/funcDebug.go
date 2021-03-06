/*
 * Copyright 2020 The FATE Authors. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package enum

import "fate.manager/entity"

type FuncDebug int32

const (
	FuncDebug_UNKNOWN FuncDebug = -1
	FuncDebug_ON      FuncDebug = 1
	FuncDebug_OFF     FuncDebug = 2
)

func GetFuncDebugString(p FuncDebug) string {
	switch p {
	case FuncDebug_ON:
		return "ON"
	case FuncDebug_OFF:
		return "OFF"
	default:
		return "unknown"
	}
}

func GetFuncDebug_OFFList() []entity.IdPair {
	var idPairList []entity.IdPair
	for i := 1; i < 3; i++ {
		idPair := entity.IdPair{i, GetFuncDebugString(FuncDebug(i))}
		idPairList = append(idPairList, idPair)
	}
	return idPairList
}
