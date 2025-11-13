package core

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	acID         = "8"
	customB64ABC = "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA"
)

var servers = []string{"https://w.xidian.edu.cn", "https://10.255.44.33"}

// ================================================================================= //
//                                 主要业务流程                                      //
// ================================================================================= //

// performLogin 执行完整的登录流程
func PerformLogin(client *http.Client, username, password, domain string) {
	fullUsername := username + domain
	var success bool

	for _, host := range servers {
		uiLogf("正在尝试连接服务器: %s ...\n", host)

		userIP, err := getIpAddress(client, host)
		if err != nil {
			uiLogf("从 %s 获取IP失败: %v\n", host, err)
			continue
		}

		token, err := getChallengeToken(client, host, userIP, fullUsername)
		if err != nil {
			uiLogf("从 %s 获取Token失败: %v\n", host, err)
			continue
		}

		hmd5 := calculateHmacMD5(password, token)
		info := encodeUserInfo(userIP, fullUsername, password, token)
		chksum := calculateChecksum(token, userIP, hmd5, info, fullUsername)

		err = finalLogin(client, host, userIP, hmd5, info, chksum, fullUsername)
		if err != nil {
			uiLogf("向 %s 发起登录失败: %v\n", host, err)
			continue
		}

		// 只要有一个服务器成功，就标记并跳出循环
		uiLogf("login to \"%s\" success! IP \"%s\" is now authorized!\n", host, userIP)
		success = true
		break
	}

	if !success {
		uiLogln("错误: 所有服务器均尝试失败，请检查是否已正确连接到校园网。")
	}
}

// checkStatus 执行在线状态查询流程
func CheckStatus(client *http.Client) bool {
	var success bool
	loggedIn := false
	for _, host := range servers {
		uiLogf("正在尝试连接服务器: %s ...\n", host)

		userIP, err := getIpAddress(client, host)
		if err != nil {
			uiLogf("从 %s 获取IP失败: %v\n", host, err)
			continue
		}

		// 构造查询URL
		callback := fmt.Sprintf("jQuery11240%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000)
		apiURL := fmt.Sprintf("%s/cgi-bin/rad_user_info?callback=%s&ip=%s&_=%d", host, callback, userIP, time.Now().UnixMilli())

		respBody, err := makeRequest(client, apiURL)
		if err != nil {
			uiLogf("向 %s 查询状态失败: %v\n", host, err)
			continue
		}

		// 解析JSONP响应
		jsonStr := strings.TrimPrefix(string(respBody), callback+"(")
		jsonStr = strings.TrimSuffix(jsonStr, ")")

		var statusInfo map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &statusInfo); err != nil {
			uiLogf("解析来自 %s 的状态信息失败: %v\n", host, err)
			continue
		}

		// 检查并打印状态
		if errStr, ok := statusInfo["error"].(string); ok && errStr == "ok" {
			uiLogln("--- 当前在线状态 ---")
			uiLogf("账号: %v\n", statusInfo["user_name"])
			uiLogf("姓名: %v\n", statusInfo["real_name"])
			uiLogf("已用流量: %.2f MB\n", statusInfo["sum_bytes"].(float64)/(1024*1024))
			uiLogf("已用时长: %.2f 分钟\n", statusInfo["sum_seconds"].(float64)/60)
			uiLogf("账户余额: %.2f\n", statusInfo["user_balance"].(float64))
			uiLogf("当前IP: %v\n", statusInfo["online_ip"])
			uiLogln("--------------------")
			loggedIn = true
		} else {
			uiLogf("当前未登录或状态异常。服务器消息: %v\n", statusInfo["error_msg"])
		}

		success = true
		break
	}
	if !success {
		uiLogln("错误: 所有服务器均尝试失败，请检查是否已正确连接到校园网。")
	}
	return loggedIn
}

// ================================================================================= //
//                            核心认证与加密函数 (已验证)                            //
// ================================================================================= //

func getIpAddress(client *http.Client, host string) (string, error) {
	body, err := makeRequest(client, host+"/srun_portal_pc?ac_id="+acID)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`ip\s*:\s*"(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", errors.New("在页面中未找到 IP 地址")
	}
	return matches[1], nil
}

func getChallengeToken(client *http.Client, host, userIP, fullUsername string) (string, error) {
	callback := fmt.Sprintf("jQuery11240%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000)
	apiURL := fmt.Sprintf("%s/cgi-bin/get_challenge?callback=%s&username=%s&ip=%s&_=%d", host, callback, url.QueryEscape(fullUsername), userIP, time.Now().UnixMilli())
	body, err := makeRequest(client, apiURL)
	if err != nil {
		return "", err
	}

	jsonStr := strings.TrimPrefix(string(body), callback+"(")
	jsonStr = strings.TrimSuffix(jsonStr, ")")

	var res struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		return "", fmt.Errorf("解析JSON失败: %w", err)
	}
	if res.Challenge == "" {
		return "", errors.New("获取到的 challenge 为空")
	}
	return res.Challenge, nil
}

func finalLogin(client *http.Client, host, userIP, hmd5, info, chksum, fullUsername string) error {
	params := url.Values{}
	params.Set("callback", fmt.Sprintf("jQuery11240%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000))
	params.Set("action", "login")
	params.Set("username", fullUsername)
	params.Set("password", "{MD5}"+hmd5)
	params.Set("os", "Windows 10")
	params.Set("name", "Windows")
	params.Set("double_stack", "0")
	params.Set("chksum", chksum)
	params.Set("info", info)
	params.Set("ac_id", acID)
	params.Set("ip", userIP)
	params.Set("n", "200")
	params.Set("type", "1")
	params.Set("_", fmt.Sprintf("%d", time.Now().UnixMilli()))

	finalURL := host + "/cgi-bin/srun_portal?" + params.Encode()
	body, err := makeRequest(client, finalURL)
	if err != nil {
		return err
	}

	if !strings.Contains(string(body), "\"error\":\"ok\"") && !strings.Contains(string(body), "\"suc_msg\":\"login_ok\"") {
		return fmt.Errorf("登录失败，服务器响应: %s", string(body))
	}
	return nil
}

func calculateHmacMD5(data, key string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func encodeUserInfo(userIP, fullUsername, password, token string) string {
	userInfo := map[string]string{
		"username": fullUsername,
		"password": password,
		"ip":       userIP,
		"acid":     acID,
		"enc_ver":  "srun_bx1",
	}
	jsonData, _ := json.Marshal(userInfo)
	encryptedData := srunXXTEAEncrypt(string(jsonData), token)
	customEncoder := base64.NewEncoding(customB64ABC)
	b64EncodedData := customEncoder.EncodeToString(encryptedData)
	return "{SRBX1}" + b64EncodedData
}

func calculateChecksum(token, userIP, hmd5, info string, fullUsername string) string {
	var builder strings.Builder
	builder.WriteString(token)
	builder.WriteString(fullUsername)
	builder.WriteString(token)
	builder.WriteString(hmd5)
	builder.WriteString(token)
	builder.WriteString(acID)
	builder.WriteString(token)
	builder.WriteString(userIP)
	builder.WriteString(token)
	builder.WriteString("200")
	builder.WriteString(token)
	builder.WriteString("1")
	builder.WriteString(token)
	builder.WriteString(info)
	h := sha1.New()
	h.Write([]byte(builder.String()))
	return hex.EncodeToString(h.Sum(nil))
}
