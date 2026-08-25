package logic

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// DecryptUserInfo @title DecryptUserInfo
// @description	解析密文，获取详细的用户信息
// @auth	Snactop		2023-11-13	19:27
// @param	userInfo *TravelModel.TraUser, sessionKey, encryptedData, iv string
// @return	error	传出错误信息
func DecryptUserInfo(sessionKey, encryptedData, iv string) (map[string]interface{}, error) {
	// 1. Base64解码
	sessionKeyBytes, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("解密失败: 无法Base64解码sessionKey: %v", err)
	}

	encryptedDataBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("解密失败: 无法Base64解码encryptedData: %v", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("解密失败: 无法Base64解码iv: %v", err)
	}

	// 2. 创建AES解密块
	block, err := aes.NewCipher(sessionKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("解密失败: 无法创建AES块: %v", err)
	}

	if len(encryptedDataBytes)%aes.BlockSize != 0 {
		return nil, errors.New("encryptedData is not a multiple of the block size")
	}

	// 3. 解密过程
	if len(encryptedDataBytes)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("解密失败: 密文长度不是BlockSize的整数倍")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	decrypted := make([]byte, len(encryptedDataBytes))
	mode.CryptBlocks(decrypted, encryptedDataBytes)

	// 4. 去除填充
	decrypted = pkcs7Unpad(decrypted)
	if decrypted == nil {
		return nil, errors.New("decryption failed")
	}

	// 转换为 JSON 格式的 map
	var result map[string]interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// pkcs7Unpad @title pkcs7Unpad
// @description  pkcs7Unpad 对数据进行PKCS#7去填充
// @auth	Snactop		2023-11-13	19:27
// @param	data []byte
// @return	[]byte
func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil
	}
	return data[:(length - unpadding)]
}
