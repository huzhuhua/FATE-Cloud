package job_service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fate.manager/comm/e"
	"fate.manager/comm/enum"
	"fate.manager/comm/http"
	"fate.manager/comm/logging"
	"fate.manager/comm/setting"
	"fate.manager/comm/util"
	"fate.manager/entity"
	"fate.manager/models"
	"fate.manager/services/federated_service"
	"fate.manager/services/k8s_service"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type CloudSystemAdd struct {
	SiteId           int64  `json:"id"`
	ComponentName    string `json:"installItems"`
	JobType          string `json:"type"`
	JobStatus        int    `json:"updateStatus"`
	UpdateTime       int64  `json:"updateTime"`
	ComponentVersion string `json:"version"`
}

func JobTask() {

	deployJob := models.DeployJob{
		Status: int(enum.JOB_STATUS_RUNNING),
	}
	deployJobList, err := models.GetDeployJob(deployJob, true)
	if err != nil {
		return
	}
	for i := 0; i < len(deployJobList); i++ {
		user := util.User{
			UserName: "admin",
			Password: "admin",
		}
		kubefateUrl := k8s_service.GetKubenetesUrl(deployJobList[i].FederatedId, deployJobList[i].PartyId)
		token, err := util.GetToken(kubefateUrl, user)
		if err != nil {
			continue
		}
		authorization := fmt.Sprintf("Bearer %s", token)
		head := make(map[string]interface{})
		head["Authorization"] = authorization
		kubenetesUrl := k8s_service.GetKubenetesUrl(deployJobList[i].FederatedId, deployJobList[i].PartyId)
		result, err := http.GET(http.Url(kubenetesUrl+"/v1/job/"+deployJobList[0].JobId), nil, head)
		if err != nil || result == nil {
			continue
		}
		var jobQueryResp entity.JobQueryResp
		index := bytes.IndexByte([]byte(result.Body), 0)
		err = json.Unmarshal([]byte(result.Body)[:index], &jobQueryResp)
		if err != nil {
			logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
			continue
		}

		var clusterConfig entity.ClusterConfig
		if result.StatusCode == 200 {

			deployComponent := models.DeployComponent{
				FederatedId: deployJobList[i].FederatedId,
				PartyId:     deployJobList[i].PartyId,
				ProductType: deployJobList[i].ProductType,
				IsValid:     int(enum.IS_VALID_YES),
			}
			deployComponentList, err := models.GetDeployComponent(deployComponent)
			if err != nil || len(deployComponentList) == 0 {
				logging.Debug("no site info")
				continue
			}
			deploySite := models.DeploySite{
				FederatedId: deployJobList[i].FederatedId,
				PartyId:     deployJobList[i].PartyId,
				ProductType: deployJobList[i].ProductType,
				IsValid:     int(enum.IS_VALID_YES),
			}
			deploySiteList, err := models.GetDeploySite(&deploySite)
			if err != nil || len(deploySiteList) == 0 {
				continue
			}

			deployJob := models.DeployJob{
				FederatedId: deployJobList[i].FederatedId,
				PartyId:     deployJobList[i].PartyId,
				ProductType: deployJobList[i].ProductType,
				JobId:       deployJobList[i].JobId,
			}
			deployStatus := enum.DeployStatus_INSTALLED
			logging.Debug(jobQueryResp)
			if jobQueryResp.Data.Status == "Success" {
				if err := json.Unmarshal([]byte(deploySiteList[0].Config), &clusterConfig); err != nil {
					continue
				}
				deployJob.Status = int(enum.JOB_STATUS_SUCCESS)
				var componentVersonMap = make(map[string]interface{})
				for j := 0; j < len(deployComponentList); j++ {
					componentVersonMap[deployComponentList[j].ComponentName] = deployComponentList[j].ComponentVersion
					autoTest := models.AutoTest{
						FederatedId: deployJobList[i].FederatedId,
						PartyId:     deployJobList[i].PartyId,
						ProductType: deployJobList[i].ProductType,
						FateVersion: deployComponentList[j].FateVersion,
						TestItem:    deployComponentList[j].ComponentName,
						CreateTime:  time.Now(),
						UpdateTime:  time.Now(),
					}
					autoTestList, _ := models.GetAutoTest(autoTest)
					if len(autoTestList) == 0 {
						autoTest.Status = int(enum.TEST_STATUS_WAITING)
						models.AddAutoTest(autoTest)
					} else {
						var data = make(map[string]interface{})
						data["status"] = int(enum.TEST_STATUS_WAITING)
						models.UpdateAutoTest(data, autoTest)
					}
					if j == len(deployComponentList)-1 {
						InserTestItem(autoTest, enum.TEST_ITEM_TOY)
						InserTestItem(autoTest, enum.TEST_ITEM_FAST)
						InserTestItem(autoTest, enum.TEST_ITEM_NORMAL)
						InserTestItem(autoTest, enum.TEST_ITEM_SINGLE)
					}
				}

				componentVersonMapjson, _ := json.Marshal(componentVersonMap)
				site := models.SiteInfo{
					FederatedId:        deployJobList[i].FederatedId,
					PartyId:            deployJobList[i].PartyId,
					FateVersion:        deploySiteList[0].FateVersion,
					ComponentVersion:   string(componentVersonMapjson),
					FateServingVersion: "1.2.1",
				}
				models.UpdateSite(&site)
			} else if jobQueryResp.Data.Status == "Running" {
				deployJob.Status = int(enum.JOB_STATUS_RUNNING)
				deployStatus = enum.DeployStatus_INSTALLING
			} else if jobQueryResp.Data.Status == "Failed" {
				deployJob.Status = int(enum.JOB_STATUS_FAILED)
				deployStatus = enum.DeployStatus_INSTALLED_FAILED
			} else if jobQueryResp.Data.Status == "Created" {
				deployStatus = enum.DeployStatus_INSTALLING
			}
			var data = make(map[string]interface{})
			data["status"] = deployJob.Status
			data["cluster_id"] = jobQueryResp.Data.ClusterId
			data["start_time"] = jobQueryResp.Data.StartTime
			data["end_time"] = jobQueryResp.Data.EndTime
			data["result"] = jobQueryResp.Data.Result
			data["deploy_status"] = int(deployStatus)
			data["update_time"] = time.Now()
			models.UpdateDeployJob(data, &deployJob)

			data = make(map[string]interface{})
			data["deploy_status"] = int(deployStatus)
			data["cluster_id"] = jobQueryResp.Data.ClusterId
			starttime := jobQueryResp.Data.StartTime.UnixNano() / 1e6
			endtime := jobQueryResp.Data.EndTime.UnixNano() / 1e6

			var duration int64
			duration = 0
			if endtime-starttime >= 0 {
				duration = endtime - starttime
			}
			data["duration"] = duration
			if jobQueryResp.Data.Status == "Success" {
				data["proxy_port"] = clusterConfig.Proxy.NodePort
			}
			deploySite.JobId = deployJobList[i].JobId
			data["click_type"] = int(enum.ClickType_PULL)
			models.UpdateDeploySite(data, deploySite)

			data = make(map[string]interface{})
			data["deploy_status"] = int(deployStatus)
			data["duration"] = duration
			data["start_time"] = jobQueryResp.Data.StartTime
			data["end_time"] = jobQueryResp.Data.EndTime
			models.UpdateDeployComponent(data, deployComponent)

			if jobQueryResp.Data.Status == "Success" || jobQueryResp.Data.Status == "Failed" {
				siteInfo, err := models.GetSiteInfo(deploySiteList[0].PartyId, deploySiteList[0].FederatedId)
				if err != nil {
					continue
				}
				federatedInfo, err := models.GetFederatedUrlById(deploySiteList[0].FederatedId)
				if err != nil {
					continue
				}
				models.GetFederatedUrlById(deploySiteList[0].FederatedId)
				var cloudSystemAddList []CloudSystemAdd
				for k := 0; k < len(deployComponentList); k++ {
					if deployJob.Status == int(enum.JOB_STATUS_SUCCESS) || deployJob.Status == int(enum.JOB_STATUS_FAILED) {
						cloudSystemAdd := CloudSystemAdd{
							SiteId:           siteInfo.SiteId,
							ComponentName:    deployComponentList[k].ComponentName,
							JobType:          enum.GetJobTypeString(enum.JobType(deployJobList[i].JobType)),
							JobStatus:        1,
							UpdateTime:       jobQueryResp.Data.EndTime.UnixNano() / 1e6,
							ComponentVersion: deployComponentList[k].ComponentVersion,
						}
						if deployJob.Status == int(enum.JOB_STATUS_FAILED) {
							cloudSystemAdd.JobStatus = 2
						}
						cloudSystemAddList = append(cloudSystemAddList, cloudSystemAdd)
					}
				}
				if len(cloudSystemAddList) > 0 {
					cloudSystemAddJson, _ := json.Marshal(cloudSystemAddList)
					headInfo := util.HeaderInfo{
						AppKey:    siteInfo.AppKey,
						AppSecret: siteInfo.AppSecret,
						PartyId:   siteInfo.PartyId,
						Body:      string(cloudSystemAddJson),
						Role:      siteInfo.Role,
						Uri:       setting.SystemAddUri,
					}
					headerInfoMap := util.GetHeaderInfo(headInfo)
					result, err := http.POST(http.Url(federatedInfo.FederatedUrl+setting.SystemAddUri), cloudSystemAddList, headerInfoMap)
					if err != nil {
						logging.Debug(e.GetMsg(e.ERROR_HTTP_FAIL))
						continue
					}
					if len(result.Body) > 0 {
						var updateResp entity.CloudCommResp
						err = json.Unmarshal([]byte(result.Body), &updateResp)
						if err != nil {
							logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
							return
						}
						if updateResp.Code == e.SUCCESS {
							msg := "federatedId:" + strconv.Itoa(siteInfo.FederatedId) + ",partyId:" + strconv.Itoa(siteInfo.PartyId) + ",status:system add success"
							logging.Debug(msg)
						}
					}
				}
			}
		}
	}
}

func InserTestItem(autoTest models.AutoTest, testItemType enum.TestItemType) {
	autoTest.TestItem = enum.GetTestItemString(testItemType)
	autoTestList, _ := models.GetAutoTest(autoTest)
	if len(autoTestList) == 0 {
		models.AddAutoTest(autoTest)
	}
}

func TestOnlyTask() {
	deploySite := models.DeploySite{
		ProductType: int(enum.PRODUCT_TYPE_FATE),
		ToyTestOnly: int(enum.ToyTestOnly_TESTING),
		IsValid:     int(enum.IS_VALID_YES),
	}
	deploySiteList, err := models.GetDeploySite(&deploySite)
	if err != nil {
		return
	}
	for i := 0; i < len(deploySiteList); i++ {
		logdir := fmt.Sprintf("./runtime/test/toy/fate-%d.log", deploySiteList[i].FederatedId, deploySiteList[i].PartyId)
		if !util.FileExists(logdir) {
			continue
		}
		file, err := os.Open(logdir)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		testtag := false
		existfile := false
		for scanner.Scan() {
			existfile = true
			lineText := scanner.Text()
			ret := strings.Index(lineText, "success to calculate secure_sum")
			if ret > 0 {
				testtag = true
				break
			}
		}
		if existfile {
			var data = make(map[string]interface{})
			data["toy_test_only_read"] = int(enum.ToyTestOnlyTypeRead_NO)
			data["toy_test_only"] = int(enum.ToyTestOnly_FAILED)
			if testtag {
				data["toy_test_only"] = int(enum.ToyTestOnly_SUCCESS)
			}
			deploySite := models.DeploySite{
				FederatedId: deploySiteList[i].FederatedId,
				PartyId:     deploySiteList[i].PartyId,
				ProductType: deploySiteList[i].ProductType,
				IsValid:     int(enum.IS_VALID_YES),
			}
			models.UpdateDeploySite(data, deploySite)
		}
	}
}

type ApplyResultReq struct {
	Institutions string `json:"institutions"`
}

func ApplyResultTask(info *models.AccountInfo) {
	applySiteInfo := models.ApplySiteInfo{
		ReadStatus: int(enum.APPLY_READ_STATUS_NOT_READ),
		Status:     int(enum.IS_VALID_ING),
	}
	applySiteInfoList, err := models.GetApplySiteInfo(applySiteInfo)
	if err != nil {
		return
	}
	if len(applySiteInfoList) == 0 {
		applySiteInfo = models.ApplySiteInfo{
			Status:     int(enum.IS_VALID_YES),
			ReadStatus: -1,
		}
		applySiteInfoList, err = models.GetApplySiteInfo(applySiteInfo)
		if err != nil || len(applySiteInfoList) == 0 {
			return
		}
		var applyResultReq ApplyResultReq
		applyResultReq.Institutions = info.Institutions
		applyResultReqJson, _ := json.Marshal(applyResultReq)
		headInfo := util.UserHeaderInfo{
			UserAppKey:    info.AppKey,
			UserAppSecret: info.AppSecret,
			FateManagerId: info.CloudUserId,
			Body:          string(applyResultReqJson),
			Uri:           setting.AuthorityResults,
		}
		headerInfoMap := util.GetUserHeadInfo(headInfo)
		federationList, err := federated_service.GetFederationInfo()
		if err != nil || len(federationList) == 0 {
			return
		}
		result, err := http.POST(http.Url(federationList[0].FederatedUrl+setting.AuthorityResults), applyResultReq, headerInfoMap)
		if err != nil {
			logging.Debug(e.GetMsg(e.ERROR_HTTP_FAIL))
			return
		}
		var applySiteResultResp entity.ApplySiteResultResp
		err = json.Unmarshal([]byte(result.Body), &applySiteResultResp)
		if err != nil {
			logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
			return
		}

		var auditResult []entity.IdPair
		err = json.Unmarshal([]byte(applySiteInfoList[0].Institutions), &auditResult)
		if err != nil {
			logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
			return
		}
		if applySiteResultResp.Code == e.SUCCESS {
			updateTag := false
			for l := 0; l < len(auditResult); l++ {
				hittag := false
				for k := 0; k < len(applySiteResultResp.Data); k++ {
					item := applySiteResultResp.Data[k]

					if auditResult[l].Desc == item.AuthorityInstitutions {
						if item.Status != auditResult[l].Code {
							updateTag = true
						}
						auditResult[l].Code = item.Status
						break
						hittag = true
					}
				}
				if !hittag {
					auditResult[l].Code = int(enum.AuditStatus_REJECTED)
					updateTag = true
				}
			}
			if updateTag {
				info := models.ApplySiteInfo{
					UserId:   applySiteInfoList[0].UserId,
					UserName: applySiteInfoList[0].UserName,
					Status:   int(enum.IS_VALID_YES),
				}
				var data = make(map[string]interface{})
				data["status"] = int(enum.IS_VALID_NO)
				data["update_time"] = time.Now()
				models.UpdateApplySiteInfo(data, info)

				auditResultJson, _ := json.Marshal(auditResult)
				applySiteInfo.Institutions = string(auditResultJson)
				applySiteInfo.UpdateTime = time.Now()
				applySiteInfo.ReadStatus = int(enum.APPLY_READ_STATUS_NOT_READ)
				applySiteInfo.Status = int(enum.IS_VALID_YES)

				models.AddApplySiteInfo(&applySiteInfo)
			}
		}
	} else {
		for i := 0; i < len(applySiteInfoList); i++ {
			var auditResult []entity.IdPair
			err = json.Unmarshal([]byte(applySiteInfoList[i].Institutions), &auditResult)
			if err != nil {
				logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
				continue
			}
			remoteTag := false
			for j := 0; j < len(auditResult); j++ {
				if auditResult[j].Code == int(enum.AuditStatus_AUDITING) {
					remoteTag = true
					break
				}
			}
			if remoteTag {
				var applyResultReq ApplyResultReq
				applyResultReq.Institutions = info.Institutions
				applyResultReqJson, _ := json.Marshal(applyResultReq)
				headInfo := util.UserHeaderInfo{
					UserAppKey:    info.AppKey,
					UserAppSecret: info.AppSecret,
					FateManagerId: info.CloudUserId,
					Body:          string(applyResultReqJson),
					Uri:           setting.AuthorityResults,
				}
				headerInfoMap := util.GetUserHeadInfo(headInfo)
				federationList, err := federated_service.GetFederationInfo()
				if err != nil || len(federationList) == 0 {
					continue
				}
				result, err := http.POST(http.Url(federationList[0].FederatedUrl+setting.AuthorityResults), applyResultReq, headerInfoMap)
				if err != nil {
					logging.Debug(e.GetMsg(e.ERROR_HTTP_FAIL))
					continue
				}
				var applySiteResultResp entity.ApplySiteResultResp
				err = json.Unmarshal([]byte(result.Body), &applySiteResultResp)
				if err != nil {
					logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
					continue
				}

				if applySiteResultResp.Code == e.SUCCESS {
					updateTag := false
					for k := 0; k < len(applySiteResultResp.Data); k++ {
						item := applySiteResultResp.Data[k]
						for l := 0; l < len(auditResult); l++ {
							if auditResult[l].Desc == item.AuthorityInstitutions {
								auditResult[l].Code = item.Status
								if item.Status != int(enum.AuditStatus_AUDITING) {
									updateTag = true
									//if item.Status == int(enum.AuditStatus_AGREED) {
									//	auditTag = true
									//}
								}
								break
							}
						}
					}
					if updateTag {
						info := models.ApplySiteInfo{
							UserId:   applySiteInfoList[i].UserId,
							UserName: applySiteInfoList[i].UserName,
							Status:   int(enum.IS_VALID_YES),
						}
						var data = make(map[string]interface{})
						data["status"] = int(enum.IS_VALID_NO)
						data["update_time"] = time.Now()
						models.UpdateApplySiteInfo(data, info)

						auditResultJson, _ := json.Marshal(auditResult)
						data["institutions"] = string(auditResultJson)
						data["update_time"] = time.Now()

						data["read_status"] = int(enum.APPLY_READ_STATUS_NOT_READ)

						data["status"] = int(enum.IS_VALID_YES)

						models.UpdateApplySiteInfo(data, applySiteInfo)
					}
				}
			}
		}
	}
}

type ApplyiedReq struct {
	Institutions string `json:"institutions"`
}

func AllowApplyTask(info *models.AccountInfo) {

	applyiedReq := ApplyiedReq{Institutions: info.Institutions}
	applyiedReqJson, _ := json.Marshal(applyiedReq)
	headInfo := util.UserHeaderInfo{
		UserAppKey:    info.AppKey,
		UserAppSecret: info.AppSecret,
		FateManagerId: info.CloudUserId,
		Body:          string(applyiedReqJson),
		Uri:           setting.AuthorityApplied,
	}
	headerInfoMap := util.GetUserHeadInfo(headInfo)
	federationList, err := federated_service.GetFederationInfo()
	if err != nil || len(federationList) == 0 {
		return
	}
	result, err := http.POST(http.Url(federationList[0].FederatedUrl+setting.AuthorityApplied), applyiedReq, headerInfoMap)
	if err != nil {
		logging.Debug(e.GetMsg(e.ERROR_HTTP_FAIL))
		return
	}
	var applySiteResultResp entity.AppliedResultResp
	err = json.Unmarshal([]byte(result.Body), &applySiteResultResp)
	if err != nil {
		logging.Debug(e.GetMsg(e.ERROR_PARSE_JSON_ERROR))
		return
	}

	if applySiteResultResp.Code == e.SUCCESS {
		var institutions string
		for i := 0; i < len(applySiteResultResp.Data); i++ {
			item := applySiteResultResp.Data[i]
			institutions = fmt.Sprintf("%s,%s", item, institutions)
		}
		if len(institutions) > 0 {
			var data = make(map[string]interface{})
			data["allow_instituions"] = institutions[0 : len(institutions)-1]
			accountInfo := models.AccountInfo{
				UserId:   info.UserId,
				UserName: info.UserName,
				Role:     int(enum.UserRole_ADMIN),
				Status:   int(enum.IS_VALID_YES),
			}
			models.UpdateAccountInfo(data, accountInfo)
		}
	}
}

func ComponentStatusTask() {

	deployComponent := models.DeployComponent{
		ProductType:  int(enum.PRODUCT_TYPE_FATE),
		DeployStatus: int(enum.DeployStatus_SUCCESS),
		IsValid:      int(enum.IS_VALID_YES),
	}
	deployComponentList, err := models.GetDeployComponent(deployComponent)
	if err != nil {
		return
	}
	for i := 0; i < len(deployComponentList); i++ {
		testname := deployComponentList[i].ComponentName
		if deployComponentList[i].ComponentName == "meta-service" {
			testname = "roll"
		}
		if deployComponentList[i].ComponentName == "fateboard" || deployComponentList[i].ComponentName == "fateflow" {
			testname = "python"
		}
		cmdSub := fmt.Sprintf("kubectl get pods -n fate-%d |grep %s* | grep Running |wc -l", deployComponentList[i].PartyId, testname)
		result, _ := util.ExecCommand(cmdSub)
		cnt, _ := strconv.Atoi(result[0:1])
		var data = make(map[string]interface{})
		if cnt < 1 {
			data["status"] = int(enum.SITE_RUN_STATUS_STOPPED)
		} else {
			data["status"] = int(enum.SITE_RUN_STATUS_RUNNING)
		}
		deployComponent = models.DeployComponent{
			FederatedId:   deployComponentList[i].FederatedId,
			PartyId:       deployComponentList[i].PartyId,
			ProductType:   deployComponentList[i].ProductType,
			ComponentName: deployComponentList[i].ComponentName,
			IsValid:       int(enum.IS_VALID_YES),
		}
		models.UpdateDeployComponent(data, deployComponent)
	}
}