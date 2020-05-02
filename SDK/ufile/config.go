package ufile

import (
	"encoding/json"
)

/* config 序列化字段
 */
type Config struct {
	PublicKey 		string `json:"public_key"`
	PrivateKey 		string `json:"private_key"`
	BucketName		string `json:"bucket_name"`
	FileHost		string `json:"file_host"`
	VerifyUploadMD5	bool	`json:"verify_upload_md5"`
}

func LoadConfig(jsonConfig []byte) (*Config, error) {
	config := new(Config)
	err := json.Unmarshal(jsonConfig, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}