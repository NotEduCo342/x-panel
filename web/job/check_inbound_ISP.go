package job

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"x-panel/database"
	"x-panel/database/model"
	"x-panel/logger"
	"x-panel/web/service"
)

type CheckIspClientJob struct {
	xrayService    service.XrayService
	inboundService service.InboundService
}
type ResponseStruct struct {
	Isp string `json:"as"`
}

func NewCheckIspClientJob() *CheckIspClientJob {
	job := new(CheckIspClientJob)
	return job
}

func (j *CheckIspClientJob) Run() {
	SetIspInDatabase()
}

func GetAccessFile() string {
	return "./access.log"
}
func CheckForValue(Ip string, Ips map[int]map[string]interface{}) bool {
	for _, value := range Ips {
		if value["ip"] == Ip {
			return true
		}
	}
	return false
}
func ReturnIpMap(s string) map[int]map[string]interface{} {
	re := regexp.MustCompile("([0-9]+\\.[0-9]+\\.[0-9]+\\.[0-9]+):([0-9]+)")
	reg := re.FindAllStringSubmatch(s, -1)
	data := make(map[int]map[string]interface{})
	for a, match := range reg {
		if match[1] == "127.0.0.1" || match[1] == "0.0.0.0" || match[1] == "1.1.1.1" {
			continue
		}
		if CheckForValue(match[1], data) {
			continue
		}
		data[a] = make(map[string]interface{})
		data[a]["ip"] = match[1]
		data[a]["port"] = match[2]
	}
	return data
}
func ReturnIpByPort(InboundPort int, IpMap map[int]map[string]interface{}) string {
	for _, Ip := range IpMap {
		IspPort := fmt.Sprintf("%v", Ip["port"])
		GivenPort := fmt.Sprintf("%d", InboundPort)
		if IspPort == GivenPort {
			ip := fmt.Sprintf("%v", Ip["port"])
			return ip
		}
	}
	return "unknown"
}
func FindIspCompany(ip string) string {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	response, err := http.Get(url)
	if err != nil {
		logger.Warning("something went wrong at read data from \"http://ip-api.com/json\":", err)
	}
	defer response.Body.Close()
	responseBody, err2 := ioutil.ReadAll(response.Body)
	if err2 != nil {
		logger.Warning("something went wrong on convert request body to []byte:", err2)
	}
	var responseStruct ResponseStruct
	err3 := json.Unmarshal(responseBody, &responseStruct)
	if err3 != nil {
		fmt.Println("something went wrong on get isp:", err3)
	}
	return responseStruct.Isp
}
func ReadAccessFile(FileAddress string) string {
	fileContent, err := ioutil.ReadFile(FileAddress)
	if err != nil {
		logger.Warning("something went wrong on read access file:", err)
	}
	fileContentStr := string(fileContent)
	return fileContentStr
}
func SetIspInDatabase() {
	IpMap := ReturnIpMap(ReadAccessFile(GetAccessFile()))
	var db = database.GetDB()
	var Inbounds []model.Inbound
	db.Find(&Inbounds)
	for i := range Inbounds {
		ip := ReturnIpByPort(Inbounds[i].Port, IpMap)
		Inbounds[i].Isp = ip
		db.Save(&Inbounds)
	}
}

//File example : "2023/09/12 07:51:29 192.168.80.1:54818 accepted tcp:example.com:443 email: tuyjxjw0r5@gmail.com
//2023/09/12 07:51:29 192.168.80.1:54821 accepted tcp:example.com:443 email: tuyjxjw0r5@gmail.com
//2023/09/12 07:51:30 192.168.80.1:54824 accepted tcp:example.com:443 email: tuyjxjw0r5@gmail.com
//2023/09/12 07:51:30 192.168.80.1:54827 accepted tcp:example.com:443 email: tuyjxjw0r5@gmail.com
//2023/09/12 07:51:30 192.168.80.1:54830 accepted tcp:example.com:443 email: tuyjxjw0r5@gmail.com
//2023/09/12 07:51:34 127.0.0.1:39002 accepted tcp:127.0.0.1:0 [api -> api]
//
