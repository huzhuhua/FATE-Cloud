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

package controllers

import (
	"fate.manager/comm/app"
	"fate.manager/comm/e"
	"fate.manager/services/monitor_service"
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary Get Monitor Total
// @Tags MonitorController
// @Accept  json
// @Produce  json
// @Success 200 {object} app.MonitorResponse
// @Failure 500 {object} app.Response
// @Router /fate-manager/api/monitor/total [get]
func GetMonitorTotal(c *gin.Context) {
	appG := app.Gin{C: c}
	result, err := monitor_service.GetMonitorTotal()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_MONITOR_TOTAL_FAIL, nil)
		return
	}
	appG.Response(http.StatusOK, e.SUCCESS, result)
}

// @Summary Get Institution based statistics
// @Tags MonitorController
// @Accept  json
// @Produce  json
// @Success 200 {object} app.InstitutionBaseStaticsResponse
// @Failure 500 {object} app.Response
// @Router /fate-manager/api/monitor/statistics/institution [get]
func GetInstitutionBaseStatics(c *gin.Context) {
	appG := app.Gin{C: c}
	result, err := monitor_service.GetInstitutionBaseStatics()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_MONITOR_TOTAL_FAIL, nil)
		return
	}
	appG.Response(http.StatusOK, e.SUCCESS, result)
}

// @Summary Get Site based statistics
// @Tags MonitorController
// @Accept  json
// @Produce  json
// @Success 200 {object} app.SiteBaseStatisticsResponse
// @Failure 500 {object} app.Response
// @Router /fate-manager/api/monitor/statistics/site [get]
func GetSiteBaseStatistics(c *gin.Context) {
	appG := app.Gin{C: c}
	result, err := monitor_service.GetSiteBaseStatistics()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_MONITOR_TOTAL_FAIL, nil)
		return
	}
	appG.Response(http.StatusOK, e.SUCCESS, result)
}