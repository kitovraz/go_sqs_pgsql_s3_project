package apiserver

import "go_sqs_pqsql_s3_project/config"

type JwtManager struct {
	config *config.Config
}

func NewJwtManger(config *config.Config) *JwtManager {
	return &JwtManager{
		config: config,
	}
}
